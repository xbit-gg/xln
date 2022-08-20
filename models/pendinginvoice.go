package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PendingInvoice struct {
	PaymentHash string `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	WalletID       string `gorm:"index,priority:1"`
	WalletUsername string `gorm:"index,priority:2"`
	Wallet         Wallet `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Amount uint64
}

func (r *repository) CreatePendingInvoice(tx *gorm.DB, pendingInv *PendingInvoice) error {
	if pendingInv == nil {
		log.Errorf("%s. Reason: %v", MsgCreatePendingInvoiceFailed, "expected to receive pending invoice record. Received nil instead.")
		return nil
	} else if err := tx.Create(pendingInv).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"pendingInvoice": pendingInv.PaymentHash,
			"wallet":         pendingInv.WalletID,
			"user":           pendingInv.WalletUsername,
		}).Error(MsgCreatePendingInvoiceFailed)
		return fmt.Errorf("%s. Reason: %v", MsgCreatePendingInvoiceFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) DeletePendingInvoice(tx *gorm.DB, paymentHash string) error {
	if res := tx.Where("payment_hash = ?", paymentHash).Delete(&PendingInvoice{}); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"pendingInvoice": paymentHash,
		}).Error(MsgDeletePendingInvoiceFailed)
		return fmt.Errorf("%s. Reason: %v", MsgDeletePendingInvoiceFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrPendingInvoiceNotFound
	} else {
		return nil
	}
}

func (r *repository) ListWalletPendingInvoices(tx *gorm.DB, username, walletId string) ([]*PendingInvoice, error) {
	var pendingInvoices []*PendingInvoice
	if err := tx.Find(&pendingInvoices, "wallet_username = ? AND wallet_id = ?", username, walletId).Error; err != nil {
		log.WithError(err).Error(MsgListWalletPendingInvoicesFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListWalletPendingInvoicesFailed, ErrInternal)
	} else {
		return pendingInvoices, nil
	}
}

func (r *repository) ListPendingInvoices(tx *gorm.DB) ([]*PendingInvoice, error) {
	var pendingInvoices []*PendingInvoice
	if err := tx.Find(&pendingInvoices).Error; err != nil {
		log.WithError(err).Error(MsgListPendingInvoicesFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListPendingInvoicesFailed, ErrInternal)
	} else {
		return pendingInvoices, nil
	}
}

func (r *repository) GetPendingInvoice(tx *gorm.DB, paymentHash string) (*PendingInvoice, error) {
	pendingInvoice := PendingInvoice{}
	if err := tx.Take(&pendingInvoice, "payment_hash = ?", paymentHash).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrPendingInvoiceNotFound
	} else if err != nil {
		log.WithError(err).WithField("paymentHash", paymentHash).Error(MsgGetPendingInvoiceFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetPendingInvoiceFailed, ErrInternal)
	} else {
		return &pendingInvoice, nil
	}
}
