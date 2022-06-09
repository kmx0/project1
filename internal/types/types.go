package types

import (
	"time"
)

type User struct {
	Login     string `json:"login"`                
	Password  string `json:"password"`             
	IP        string `json:"ip,omitempty"`         
	UserAgent string `json:"user-agent,omitempty"` 
	Cookie    string `json:"cookie,omitempty"`     
	ID        int    `json:"id,omitempty"`         
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
