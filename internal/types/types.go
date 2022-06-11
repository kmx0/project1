package types

import (
	"time"
)

type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	IP        string  `json:"ip,omitempty"`
	UserAgent string  `json:"user-agent,omitempty"`
	Cookie    string  `json:"cookie,omitempty"`
	ID        int     `json:"id,omitempty"`
	Balance   float64 `json:"balance,omitempty"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type AccrualO struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAT time.Time `json:"processed_at,omitempty"`
}
