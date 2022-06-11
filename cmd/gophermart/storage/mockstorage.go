package storage

import (
	"context"

	"github.com/kmx0/project1/internal/crypto"
	"github.com/kmx0/project1/internal/errors"
	"github.com/kmx0/project1/internal/types"
	"github.com/sirupsen/logrus"
)

type MockStorage struct {
	Users    map[string]string
	Sessions map[int]string
	IDs      map[string]int
}

func NewMockStorage() Storage {
	users := make(map[string]string)
	ids := make(map[string]int)
	sessions := make(map[int]string)
	users["henry"] = "1qaz@WSX"
	users["bobby"] = "bababa"
	ids["henry"] = 0
	ids["bobby"] = 1

	return &MockStorage{Users: users, IDs: ids, Sessions: sessions}
}
func (ms *MockStorage) PingDB(ctx context.Context, urlExample string) bool {
	return true
}

func (ms *MockStorage) RegisterUser(user types.User) (id int, err error) {
	if _, ok := ms.Users[user.Login]; ok {
		return -1, errors.ErrStatusConflict
	}
	ms.Users[user.Login] = user.Password
	return 2, nil

}
func (ms *MockStorage) WriteUserCookie(user types.User, id int) error {
	ms.IDs[user.Login] = id
	logrus.Info(ms.Sessions[id])
	ms.Sessions[id] = user.Cookie
	logrus.Info(ms.Sessions[id])
	return nil
}

func (ms *MockStorage) LoginUser(user types.User) (id int, cookie string, err error) {
	if _, ok := ms.Users[user.Login]; !ok {
		return -1, "", errors.ErrStatusUnauthorized
	}
	if ms.Users[user.Login] != user.Password {
		return -1, "", errors.ErrStatusUnauthorized
	}
	if id, ok := ms.IDs[user.Login]; ok {
		ms.Sessions[id] = crypto.CookieHash("", "", user.Login)
		return id, ms.Sessions[id], nil
	}
	return -1, "", errors.ErrStatusUnauthorized

}
func (ms *MockStorage) GetOrdersList(cookie string) ([]types.Order, error) {
	return nil, nil
}

func (ms *MockStorage) DeleteCookie(cookie string) error {
	return nil
}
func (ms *MockStorage) CheckCookie(cookie, ip, userAgent string) error {
	return nil
}

func (ms *MockStorage) LoadNewOrder(cookie string, order string) error {
	return nil
}

func (ms *MockStorage) WriteAccrual(accrual types.AccrualO) error {
	return nil
}
func (ms *MockStorage) GetBalance(cookie string) (balance float64, err error) {
	return balance, nil
}

func (ms *MockStorage) GetUserID(cookie string) (id int, err error) {
	return id, err
}
func (ms *MockStorage) GetSUMWithdraws(userID int) (withdrawals float64, err error) {
	return withdrawals, nil
}
func (ms *MockStorage) GetWithdrawals(userID int) ([]types.Withdraw, error) {
	return nil, nil
}

func (ms *MockStorage) ChangeBalanceValue(value float64, action string, userID int) error {
	return nil
}
func (ms *MockStorage) WriteWithdraw(withdraw types.Withdraw, userID int) error {
	return nil
}
