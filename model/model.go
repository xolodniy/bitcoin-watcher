package model

import "time"

type InvoiceStatus string

const (
	StatusPending   InvoiceStatus = "pending"
	StatusSeen      InvoiceStatus = "seen_in_mempool"
	StatusConfirmed InvoiceStatus = "confirmed"
	StatusExpired   InvoiceStatus = "expired"
)

type Invoice struct {
	Address    string
	ScriptHash string
	Amount     int64 // satoshis
	Status     InvoiceStatus
	CreatedAt  time.Time

	// ID        string
	// ExpiresAt     time.Time
	TxID                  string
	TxHeight              int
	RequiredConfirmations int
	Confirmations         int // required
}
