package types

import (
	"time"
)

type User struct {
	Login     string `json:"login"`                // имя метрики
	Password  string `json:"password"`             // параметр, принимающий значение gauge или counter
	IP        string `json:"ip,omitempty"`         // параметр, принимающий значение gauge или counter
	UserAgent string `json:"user-agent,omitempty"` // параметр, принимающий значение gauge или counter
	Cookie    string `json:"cookie,omitempty"`     // параметр, принимающий значение gauge или counter
	ID        int    `json:"id,omitempty"`         // параметр, принимающий значение gauge или counter
	// Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	// Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	// Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
