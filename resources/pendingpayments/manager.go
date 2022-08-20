package pendingpayments

import (
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
)

type Manager interface {
	// ListWalletPendingPayments lists wallet's pending payments
	ListWalletPendingPayments(username, walletId string) ([]*models.PendingPayment, error)

	// GetPendingPayment returns the pending payment matching the paymentHash.
	GetPendingPayment(paymentHash string) (*models.PendingPayment, error)

	// ListPendingPayments lists all pending payments.
	ListPendingPayments() ([]*models.PendingPayment, error)
}

type manager struct {
	db *db.DB
}

func NewManager(db *db.DB) Manager {
	return &manager{db: db}
}

func (m *manager) ListWalletPendingPayments(username, walletId string) ([]*models.PendingPayment, error) {
	return m.db.Repo.ListWalletPendingPayments(m.db.DB, username, walletId)
}

func (m *manager) GetPendingPayment(paymentHash string) (*models.PendingPayment, error) {
	return m.db.Repo.GetPendingPayment(m.db.DB, paymentHash)
}

func (m *manager) ListPendingPayments() ([]*models.PendingPayment, error) {
	return m.db.Repo.ListPendingPayments(m.db.DB)
}
