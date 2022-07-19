package repo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"

	"github.com/o1egl/pidor-bot/domain"
)

func TestBoltRepo_Users(t *testing.T) {
	db := createDB(t)
	boltRepo := NewBoltRepo(db)
	user := domain.User{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "@john.doe",
		IsActive:  true,
	}

	anotherUser := domain.User{
		ID:        2,
		FirstName: "Chuck",
		LastName:  "Norris",
		Username:  "",
		IsActive:  true,
	}

	t.Run("should upsert user to chat 1", func(t *testing.T) {
		err := boltRepo.UpsertUser(context.Background(), 1, user)
		require.NoError(t, err)

		t.Run("should return user from chat 1", func(t *testing.T) {
			users, err := boltRepo.GetUsers(context.Background(), 1)
			require.NoError(t, err)
			require.Equal(t, []domain.User{user}, users)
		})
	})

	t.Run("should upsert user to chat 2", func(t *testing.T) {
		err := boltRepo.UpsertUser(context.Background(), 2, user)
		require.NoError(t, err)

		t.Run("should return user from chat 2", func(t *testing.T) {
			users, err := boltRepo.GetUsers(context.Background(), 2)
			require.NoError(t, err)
			require.Equal(t, []domain.User{user}, users)
		})
	})

	t.Run("should upsert another user to chat 1", func(t *testing.T) {
		err := boltRepo.UpsertUser(context.Background(), 1, anotherUser)
		require.NoError(t, err)

		t.Run("should return all users from chat 1", func(t *testing.T) {
			users, err := boltRepo.GetUsers(context.Background(), 1)
			require.NoError(t, err)
			require.Equal(t, []domain.User{user, anotherUser}, users)
		})
	})
}

func TestBoltRepo_Votes(t *testing.T) {
	db := createDB(t)
	boltRepo := NewBoltRepo(db)

	allVotes := []domain.Vote{
		{
			UserID: 1,
			Time:   time.Date(2020, 1, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			UserID: 1,
			Time:   time.Date(2022, 1, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			UserID: 1,
			Time:   time.Date(2022, 2, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			UserID: 1,
			Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
		},
		{
			UserID: 2,
			Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
		},
	}

	t.Run("should create votes", func(t *testing.T) {
		for i, vote := range allVotes {
			t.Run(fmt.Sprintf("vote %d", i), func(t *testing.T) {
				err := boltRepo.CreateVote(context.Background(), 1, vote)
				require.NoError(t, err)
			})
		}

		t.Run("should return all votes", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
			require.NoError(t, err)
			require.Equal(t, allVotes, gotVotes)
		})

		t.Run("should return votes with year filter", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 1, WithYear(2020))
			wantVotes := []domain.Vote{allVotes[0]}
			require.NoError(t, err)
			require.Equal(t, wantVotes, gotVotes)
		})

		t.Run("should return votes with year and month filter", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 1, WithYear(2022), WithMonth(2))
			wantVotes := []domain.Vote{allVotes[2], allVotes[3], allVotes[4]}
			require.NoError(t, err)
			require.Equal(t, wantVotes, gotVotes)
		})
		t.Run("should return votes with year, month and day filter", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 1, WithYear(2022), WithMonth(2), WithDay(2))
			wantVotes := []domain.Vote{allVotes[3], allVotes[4]}
			require.NoError(t, err)
			require.Equal(t, wantVotes, gotVotes)
		})
		t.Run("should return no votes for unsatisfied filters", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 1, WithYear(2019))
			wantVotes := make([]domain.Vote, 0)
			require.NoError(t, err)
			require.Equal(t, wantVotes, gotVotes)
		})
		t.Run("should return no votes for unexciting chat", func(t *testing.T) {
			gotVotes, err := boltRepo.GetVotes(context.Background(), 2)
			wantVotes := make([]domain.Vote, 0)
			require.NoError(t, err)
			require.Equal(t, wantVotes, gotVotes)
		})
	})
}

func TestBoltRepo_DeleteVotes(t *testing.T) {
	t.Run("should delete votes for a day", func(t *testing.T) {
		db := createDB(t)
		boltRepo := NewBoltRepo(db)
		votes := []domain.Vote{
			{
				UserID: 1,
				Time:   time.Date(2022, 1, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 1,
				Time:   time.Date(2022, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 2,
				Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
			},
		}
		createVotes(t, boltRepo, 1, votes)

		err := boltRepo.DeleteVotes(context.Background(), 1, WithYear(2022), WithMonth(2), WithDay(1))
		assert.NoError(t, err)

		gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, []domain.Vote{votes[0], votes[2]}, gotVotes)
	})

	t.Run("should delete votes for a month", func(t *testing.T) {
		db := createDB(t)
		boltRepo := NewBoltRepo(db)
		votes := []domain.Vote{
			{
				UserID: 1,
				Time:   time.Date(2022, 1, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 1,
				Time:   time.Date(2022, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 2,
				Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
			},
		}
		createVotes(t, boltRepo, 1, votes)

		err := boltRepo.DeleteVotes(context.Background(), 1, WithYear(2022), WithMonth(2))
		assert.NoError(t, err)

		gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, []domain.Vote{votes[0]}, gotVotes)
	})

	t.Run("should delete votes for a year", func(t *testing.T) {
		db := createDB(t)
		boltRepo := NewBoltRepo(db)
		votes := []domain.Vote{
			{
				UserID: 1,
				Time:   time.Date(2022, 1, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 1,
				Time:   time.Date(2022, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 2,
				Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
			},
		}
		createVotes(t, boltRepo, 1, votes)

		err := boltRepo.DeleteVotes(context.Background(), 1, WithYear(2022))
		assert.NoError(t, err)

		gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, []domain.Vote{}, gotVotes)
	})

	t.Run("should not delete all votes", func(t *testing.T) {
		db := createDB(t)
		boltRepo := NewBoltRepo(db)
		votes := []domain.Vote{
			{
				UserID: 1,
				Time:   time.Date(2020, 1, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 1,
				Time:   time.Date(2022, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			{
				UserID: 2,
				Time:   time.Date(2022, 2, 2, 10, 30, 0, 0, time.UTC),
			},
		}
		createVotes(t, boltRepo, 1, votes)

		err := boltRepo.DeleteVotes(context.Background(), 1)
		assert.NoError(t, err)

		gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, []domain.Vote{}, gotVotes)
	})

	t.Run("should not delete votes for other chats", func(t *testing.T) {
		db := createDB(t)
		boltRepo := NewBoltRepo(db)
		votes := []domain.Vote{
			{
				UserID: 1,
				Time:   time.Date(2022, 1, 1, 10, 30, 0, 0, time.UTC),
			},
		}
		createVotes(t, boltRepo, 1, votes)

		err := boltRepo.DeleteVotes(context.Background(), 2, WithYear(2022))
		assert.NoError(t, err)

		gotVotes, err := boltRepo.GetVotes(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, votes, gotVotes)
	})
}

func createVotes(t *testing.T, boltRepo Repo, chatID int64, votes []domain.Vote) {
	t.Helper()

	for _, vote := range votes {
		err := boltRepo.CreateVote(context.Background(), chatID, vote)
		require.NoError(t, err)
	}
}

func createDB(t *testing.T) *bbolt.DB {
	t.Helper()

	file, err := os.CreateTemp("", "*.db")
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})

	db, err := bbolt.Open(file.Name(), 0600, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}
