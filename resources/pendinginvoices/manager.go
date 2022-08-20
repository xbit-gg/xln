package pendinginvoices

import (
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
)

type Manager interface {
	// ListWalletPendingInvoices lists pending invoices
	ListWalletPendingInvoices(username, walletId string) ([]*models.PendingInvoice, error)

	// GetPendingInvoice returns the pending invoice matching id.
	GetPendingInvoice(paymentHash string) (*models.PendingInvoice, error)

	// ListPendingInvoices lists all pending invoices.
	ListPendingInvoices() ([]*models.PendingInvoice, error)
}

type manager struct {
	db *db.DB
}

func NewManager(db *db.DB) Manager {
	return &manager{db: db}
}

func (m *manager) ListWalletPendingInvoices(username, walletId string) ([]*models.PendingInvoice, error) {
	return m.db.Repo.ListWalletPendingInvoices(m.db.DB, username, walletId)
}

func (m *manager) GetPendingInvoice(paymentHash string) (*models.PendingInvoice, error) {
	return m.db.Repo.GetPendingInvoice(m.db.DB, paymentHash)
}

func (m *manager) ListPendingInvoices() ([]*models.PendingInvoice, error) {
	return m.db.Repo.ListPendingInvoices(m.db.DB)
}
