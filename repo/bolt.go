//go:generate $TOOLS_BIN/mockgen -package mocks -source $GOFILE -destination ./mocks/$GOFILE
package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/bbolt"

	"github.com/o1egl/pidor-bot/domain"
)

const keySeparator = ':'

var (
	ErrNotFound = errors.New("not found")
)

type statsFilters struct {
	year  int
	month int
	day   int
}

type StatsFilterOpts func(f *statsFilters)

func WithYear(year int) StatsFilterOpts {
	return func(f *statsFilters) {
		f.year = year
	}
}

func WithMonth(month int) StatsFilterOpts {
	return func(f *statsFilters) {
		if f.year == 0 {
			f.year = time.Now().Year()
		}
		f.month = month
	}
}

func WithDay(day int) StatsFilterOpts {
	return func(f *statsFilters) {
		if f.year == 0 {
			f.year = time.Now().Year()
		}
		if f.month == 0 {
			f.month = int(time.Now().Month())
		}
		f.day = day
	}
}

type Repo interface {
	UpsertUser(ctx context.Context, chatID int64, user domain.User) error
	GetUsers(ctx context.Context, chatID int64) ([]domain.User, error)
	GetUser(ctx context.Context, chatID, userID int64) (domain.User, error)

	CreateVote(ctx context.Context, chatID int64, vote domain.Vote) error
	GetVotes(ctx context.Context, chatID int64, opts ...StatsFilterOpts) ([]domain.Vote, error)
}

type BoltRepo struct {
	db *bbolt.DB
}

func NewBoltRepo(db *bbolt.DB) *BoltRepo {
	return &BoltRepo{db: db}
}

func (b *BoltRepo) UpsertUser(ctx context.Context, chatID int64, user domain.User) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		chatBucket, err := tx.CreateBucketIfNotExists(b.channelBucket(chatID))
		if err != nil {
			return err
		}

		usersBucket, err := chatBucket.CreateBucketIfNotExists(b.usersBucket())
		if err != nil {
			return err
		}

		userJSON, err := json.Marshal(user)
		if err != nil {
			return err
		}

		userKey := []byte(strconv.FormatInt(user.ID, 10))
		return usersBucket.Put(userKey, userJSON)
	})
}

func (b *BoltRepo) GetUsers(ctx context.Context, chatID int64) ([]domain.User, error) {
	users := make([]domain.User, 0)
	err := b.db.View(func(tx *bbolt.Tx) error {
		chatBucket := tx.Bucket(b.channelBucket(chatID))
		if chatBucket == nil {
			return nil
		}

		usersBucket := chatBucket.Bucket(b.usersBucket())
		if usersBucket == nil {
			return nil
		}

		return usersBucket.ForEach(func(k, v []byte) error {
			var user domain.User
			if err := json.Unmarshal(v, &user); err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (b *BoltRepo) GetUser(ctx context.Context, chatID, userID int64) (domain.User, error) {
	var user domain.User
	err := b.db.View(func(tx *bbolt.Tx) error {
		chatBucket := tx.Bucket(b.channelBucket(chatID))
		if chatBucket == nil {
			return nil
		}

		usersBucket := chatBucket.Bucket(b.usersBucket())
		if usersBucket == nil {
			return nil
		}

		userKey := []byte(strconv.FormatInt(userID, 10))
		userJSON := usersBucket.Get(userKey)
		if userJSON == nil {
			return ErrNotFound
		}

		return json.Unmarshal(userJSON, &user)
	})
	return user, err
}

func (b *BoltRepo) CreateVote(ctx context.Context, chatID int64, vote domain.Vote) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		chatBucket, err := tx.CreateBucketIfNotExists(b.channelBucket(chatID))
		if err != nil {
			return err
		}

		votesBucket, err := chatBucket.CreateBucketIfNotExists(b.votesBucket())
		if err != nil {
			return err
		}

		voteID, err := votesBucket.NextSequence()
		if err != nil {
			return err
		}
		voteKey := b.buildKey(
			[]byte(strconv.Itoa(vote.Time.Year())),
			[]byte(strconv.Itoa(int(vote.Time.Month()))),
			[]byte(strconv.Itoa(vote.Time.Day())),
			[]byte(strconv.Itoa(int(vote.UserID))),
			[]byte(strconv.Itoa(int(voteID))),
		)

		voteJSON, err := json.Marshal(vote)
		if err != nil {
			return err
		}
		return votesBucket.Put(voteKey, voteJSON)
	})
}

func (b *BoltRepo) GetVotes(ctx context.Context, chatID int64, opts ...StatsFilterOpts) ([]domain.Vote, error) {
	filters := statsFilters{}
	for _, opt := range opts {
		opt(&filters)
	}
	var statsPrefix []byte
	if filters.year > 0 {
		statsPrefix = b.buildKey([]byte(strconv.Itoa(filters.year)))
	}
	if filters.month > 0 {
		statsPrefix = b.buildKey(statsPrefix, []byte(strconv.Itoa(filters.month)))
	}
	if filters.day > 0 {
		statsPrefix = b.buildKey(statsPrefix, []byte(strconv.Itoa(filters.day)))
	}

	votes := make([]domain.Vote, 0)
	err := b.db.View(func(tx *bbolt.Tx) error {
		chatBucket := tx.Bucket(b.channelBucket(chatID))
		if chatBucket == nil {
			return nil
		}
		statsBucket := chatBucket.Bucket(b.votesBucket())
		if statsBucket == nil {
			return nil
		}

		c := statsBucket.Cursor()
		for k, v := c.Seek(statsPrefix); k != nil; k, v = c.Next() {
			if !bytes.HasPrefix(k, statsPrefix) {
				break
			}
			var vote domain.Vote
			if err := json.Unmarshal(v, &vote); err != nil {
				return err
			}
			votes = append(votes, vote)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return votes, nil
}

func (b *BoltRepo) channelBucket(chatID int64) []byte {
	return []byte(fmt.Sprintf("chat_%d", chatID))
}

func (b *BoltRepo) usersBucket() []byte {
	return []byte("users")
}

func (b *BoltRepo) votesBucket() []byte {
	return []byte("votes")
}

func (b *BoltRepo) buildKey(parts ...[]byte) []byte {
	if len(parts) == 0 {
		return nil
	}
	buf := bytes.NewBuffer(parts[0])
	for _, part := range parts[1:] {
		buf.WriteByte(keySeparator)
		buf.Write(part)
	}
	return buf.Bytes()
}
