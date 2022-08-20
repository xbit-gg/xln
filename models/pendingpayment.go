package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PendingPayment struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	WalletID       string `gorm:"index,priority:1"`
	WalletUsername string `gorm:"index,priority:2"`
	Wallet         Wallet `gorm:"constraint:OnUpdate:CASCADE,OnDelete:NO ACTION;"`

	PaymentHash string `gorm:"index,priority:3"`
	Amount      uint64

	WithdrawK1 *string
	Withdraw   *Withdraw
}

func (r *repository) ListPendingPayments(tx *gorm.DB) ([]*PendingPayment, error) {
	var pendingPayments []*PendingPayment
	if err := tx.Find(&pendingPayments).Error; err != nil {
		log.WithError(err).Error(MsgListPendingPaymentsFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListPendingPaymentsFailed, ErrInternal)
	} else {
		return pendingPayments, nil
	}
}

func (r *repository) ListWalletPendingPayments(tx *gorm.DB, username, walletId string) ([]*PendingPayment, error) {
	var pendingPayments []*PendingPayment
	res := tx.Where("wallet_username = ? AND wallet_id = ?", username, walletId).Find(&pendingPayments)
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgListWalletPendingPaymentsFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListWalletPendingPaymentsFailed, ErrInternal)
	} else {
		return pendingPayments, nil
	}
}

func (r *repository) ListPendingWithdraws(tx *gorm.DB, username, walletId, k1 string) ([]*PendingPayment, error) {
	var pendingPayments []*PendingPayment
	res := tx.Where("wallet_username = ? AND wallet_id = ? AND withdraw_k1 = ?", username, walletId, k1).Find(&pendingPayments)
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet":   walletId,
			"user":     username,
			"withdraw": k1,
		}).Error(MsgListWalletPendingWithdrawsFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListWalletPendingWithdrawsFailed, ErrInternal)
	} else {
		return pendingPayments, nil
	}
}

func (r *repository) CreatePendingPayment(tx *gorm.DB, pendingPayment *PendingPayment) error {
	if pendingPayment == nil {
		log.Errorf("%s. Reason: %v", MsgCreatePendingPaymentFailed, MsgReceivedNil)
		return nil
	} else if err := tx.Create(pendingPayment).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"pendingPayment": pendingPayment.ID,
			"wallet":         pendingPayment.WalletID,
			"user":           pendingPayment.WalletUsername,
		}).Error(MsgCreatePendingPaymentFailed)
		return fmt.Errorf("%s. Reason: %v", MsgCreatePendingPaymentFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) DeletePendingPayment(tx *gorm.DB, pendingPayment *PendingPayment) error {
	if pendingPayment == nil {
		log.Errorf("%s. Reason: %v", MsgDeletePendingPaymentFailed, "expected to receive pending payment record. Received nil instead.")
		return nil
	} else if res := tx.Delete(pendingPayment); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"pendingPayment": pendingPayment.ID,
		}).Error(MsgDeletePendingPaymentFailed)
		return fmt.Errorf("%s. Reason: %v", MsgDeletePendingPaymentFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrPendingPaymentNotFound
	} else {
		return nil
	}
}

func (r *repository) GetPendingPayment(tx *gorm.DB, paymentHash string) (*PendingPayment, error) {
	pendingPayment := PendingPayment{}
	if err := tx.Take(&pendingPayment, "payment_hash = ?", paymentHash).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrPendingPaymentNotFound
	} else if err != nil {
		log.WithError(err).WithField("paymentHash", paymentHash).Error(MsgGetPendingPaymentFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetPendingPaymentFailed, ErrInternal)
	} else {
		return &pendingPayment, nil
	}
}
