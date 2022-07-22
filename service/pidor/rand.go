package pidor

import (
	"crypto/rand"
	"math/big"

	"github.com/o1egl/pidor-bot/domain"
)

func randString(ss []string) (string, error) {
	n, err := randInt(int64(len(ss)))
	if err != nil {
		return "", err
	}
	return ss[n], nil
}

func randUser(us []domain.User) (domain.User, error) {
	n, err := randInt(int64(len(us)))
	if err != nil {
		return domain.User{}, err
	}
	return us[n], nil
}

func randInt(max int64) (int64, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return num.Int64(), nil
}
