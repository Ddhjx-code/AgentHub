package model

import "time"

type Transaction struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	AgentID   *int64    `json:"agent_id,omitempty" db:"agent_id"`
	AgentName string    `json:"agent_name,omitempty" db:"agent_name"`
	Amount    int       `json:"amount" db:"amount"`
	Status    string    `json:"status" db:"status"`
	Note      string    `json:"note,omitempty" db:"note"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

const (
	TxTypeUse      = "use"
	TxTypeRecharge = "recharge"

	TxStatusSuccess = "success"
	TxStatusFailed  = "failed"
)
