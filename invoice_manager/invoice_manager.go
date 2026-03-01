package invoice_manager

import (
	"bitcoin-watcher/electrum_client"
	"bitcoin-watcher/model"
	"encoding/json"
	"fmt"
	"sync"
)

type InvoiceManager struct {
	invoices      map[string]*model.Invoice // key = scripthash
	currentHeight int
	client        *electrum_client.ElectrumClient
	mu            sync.Mutex
}

func NewInvoiceManager(c *electrum_client.ElectrumClient) *InvoiceManager {
	return &InvoiceManager{
		invoices: make(map[string]*model.Invoice),
		client:   c,
	}
}

func (m *InvoiceManager) CreateInvoice(hash string, amount int64, confs int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv := &model.Invoice{
		// ID:             hash,
		ScriptHash:    hash,
		Amount:        amount,
		Confirmations: confs,
		Status:        model.StatusPending,
	}

	m.invoices[hash] = inv

	return m.client.SubscribeScriptHash(hash)
}
func (m *InvoiceManager) UpdateHeight(height int) {
	m.mu.Lock()
	m.currentHeight = height
	m.mu.Unlock()

	// m.checkConfirmations()
}
func (m *InvoiceManager) HandleScriptHash(hash string) {
	historyRaw, err := m.client.CallHashHistory(hash)
	if err != nil {
		fmt.Println("history error:", err)
		return
	}

	var history []struct {
		TxHash string `json:"tx_hash"`
		Height int    `json:"height"`
	}

	if err := json.Unmarshal(historyRaw, &history); err != nil {
		return
	}

	if len(history) == 0 {
		return
	}

	// Берём последнюю tx
	tx := history[len(history)-1]

	m.mu.Lock()
	defer m.mu.Unlock()

	inv := m.invoices[hash]
	if inv == nil {
		return
	}

	inv.TxID = tx.TxHash
	inv.TxHeight = tx.Height

	if tx.Height == 0 {
		inv.Status = model.StatusSeen
		fmt.Println("Invoice seen in mempool:", inv.ScriptHash)
	} else {
		m.updateInvoiceConfirmations(inv)
	}
}

func (m *InvoiceManager) updateInvoiceConfirmations(inv *model.Invoice) {
	if inv.TxHeight <= 0 {
		return
	}

	confs := m.currentHeight - inv.TxHeight + 1
	if confs >= inv.RequiredConfirmations {
		inv.Status = model.StatusConfirmed
		fmt.Println("Invoice confirmed:", inv.ScriptHash)
	}
}

func (m *InvoiceManager) checkConfirmations() {
	for _, inv := range m.invoices {
		m.updateInvoiceConfirmations(inv)
	}
}
