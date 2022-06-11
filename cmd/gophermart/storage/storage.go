package storage

import (
	"context"

	"github.com/kmx0/project1/internal/types"
)

type Storage interface {
	// Update(metric, name, value string) error
	PingDB(ctx context.Context, urlExample string) bool
	RegisterUser(user types.User) (id int, err error)
	WriteUserCookie(user types.User, id int) error
	LoginUser(user types.User) (id int, cookie string, err error)
	DeleteCookie(cookie string) error
	CheckCookie(cookie, ip, userAgent string) error
	LoadNewOrder(cookie string, order string) error
	GetOrdersList(cookie string) ([]types.Order, error)
	WriteAccrual(accrual types.AccrualO) error
	GetUserID(cookie string) (id int, err error)
	GetBalance(cookie string) (balance float64, err error)
	GetSUMWithdraws(userID int) (withdrawals float64, err error)
	GetWithdrawals(userID int) ([]types.Withdraw, error)
	ChangeBalanceValue(value float64, action string, userID int) error
	WriteWithdraw(withdraw types.Withdraw, userID int) error
}
