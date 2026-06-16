package model

import "time"

type Wallet struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Balance   int64     `json:"balance" db:"balance"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RechargePlan struct {
	ID     int64  `json:"id" db:"id"`
	Amount int    `json:"amount" db:"amount"`
	Bonus  int    `json:"bonus" db:"bonus"`
	Label  string `json:"label" db:"label"`
	Hot    bool   `json:"hot" db:"hot"`
}
