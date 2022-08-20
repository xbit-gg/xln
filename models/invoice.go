package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Invoice is a Lighning Invoice. If the invoice is pending or failed, thre is no Transaction associated.
// If it is successful then it has a Transaction associated it
type Invoice struct {
	PaymentHash string `gorm:"primaryKey" sql:"type:uuid"`

	Timestamp time.Time // the first time we encountered the invoice. Either on (our) creation or when entered into XLN
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Preimage       *string // is null for invoices from other nodes unless it has been settled.
	Pubkey         string  // belongs to initiating node. Could be our node ID
	Memo           string
	PaymentRequest string
	Settled        time.Time // is 0 if unsettled or the time of settlement
	Amount         uint64

	// if not Nil then the invoice is created by XLN and Sender
	RecipientID       *string `gorm:"index;"`
	RecipientUsername *string `gorm:"index;"`
	Recipient         *Wallet `gorm:"foreignKey:recipient_id,recipient_username;references:id,username;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	// if Nil then invoice origin is external
	SenderID       *string `gorm:"index"`
	SenderUsername *string `gorm:"index"`
	Sender         *Wallet `gorm:"foreignKey:sender_id,sender_username;references:id,username;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (r *repository) CreateInvoice(tx *gorm.DB, invoice *Invoice) error {
	if invoice == nil {
		return fmt.Errorf("%s. Reason: %v", MsgCreateInvoiceFailed, MsgReceivedNil)
	} else if err := tx.Create(invoice).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"paymentHash": invoice.PaymentHash,
		}).Error(MsgCreateInvoiceFailed)
		return fmt.Errorf("%s. Reason: %v", MsgCreateInvoiceFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) ListWalletInvoices(tx *gorm.DB, username, walletId string) ([]*Invoice, error) {
	var invoices []*Invoice
	if err := tx.Where("recipient_username = ? AND recipient_id = ?", username, walletId).
		Or("sender_username = ? AND sender_id = ?", username, walletId).Find(&invoices).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": walletId,
		}).Error(MsgListInvoicesFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListInvoicesFailed, ErrInternal)
	} else {
		return invoices, nil
	}
}

func (r *repository) GetInvoice(tx *gorm.DB, paymentHash string) (*Invoice, error) {
	var invoice Invoice
	if err := tx.Take(&invoice, "payment_hash = ?", paymentHash).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrInvoiceNotFound
	} else if err != nil {
		log.WithError(err).WithField("paymentHash", paymentHash).Error(MsgGetInvoiceFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetInvoiceFailed, ErrInternal)
	} else {
		return &invoice, nil
	}
}

func (r *repository) GetWalletInvoice(tx *gorm.DB, username, walletId, paymentHash string) (*Invoice, error) {
	var invoice Invoice
	if err := tx.Where(tx.Where("payment_hash = ?", paymentHash).Where(tx.
		Where("recipient_username = ? AND recipient_id = ?", username, walletId).
		Or("sender_username = ? AND sender_id = ?", username, walletId)),
	).Take(&invoice).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrInvoiceNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":        username,
			"wallet":      walletId,
			"paymentHash": paymentHash,
		}).Error(MsgGetInvoiceFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetInvoiceFailed, ErrInternal)
	} else {
		return &invoice, nil
	}
}

func (r *repository) UpdateInvoiceSettleTime(tx *gorm.DB, paymentHash string, time *time.Time) error {
	if err := tx.Model(&Invoice{}).Where("payment_hash = ?", paymentHash).UpdateColumn("settled", time).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"paymentHash": paymentHash,
		}).Error(MsgUpdateInvoiceSettledDateFailed)
		return fmt.Errorf("%s. Reason: %v", MsgUpdateInvoiceSettledDateFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) NullifyInvoiceRecipient(tx *gorm.DB, username, id string) error {
	var err error
	if id == "" {
		err = tx.Model(&Invoice{}).Where("recipient_username = ?", username).Select("recipient_id", "recipient_username").
			Updates(&Invoice{RecipientID: nil, RecipientUsername: nil}).Error
	} else {
		err = tx.Model(&Invoice{}).Where("recipient_username = ? AND recipient_id = ?", username, id).Select("recipient_id", "recipient_username").
			Updates(&Invoice{RecipientID: nil, RecipientUsername: nil}).Error
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": id,
		}).Error(MsgNullifyInvoiceRecipientFailed)
		return fmt.Errorf("%s. Reason: %v", MsgNullifyInvoiceRecipientFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) NullifyInvoiceSender(tx *gorm.DB, username, id string) error {
	var err error
	if id == "" {
		err = tx.Model(&Invoice{}).Where("sender_username = ?", username).Select("sender_id", "sender_username").
			Updates(&Invoice{SenderID: nil, SenderUsername: nil}).Error
	} else {
		err = tx.Model(&Invoice{}).Where("sender_username = ? AND sender_id = ?", username, id).Select("sender_id", "sender_username").
			Updates(&Invoice{SenderID: nil, SenderUsername: nil}).Error
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": id,
		}).Error(MsgNullifyInvoiceSenderFailed)
		return fmt.Errorf("%s. Reason: %v", MsgNullifyInvoiceSenderFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) SetInvoiceSenderAmount(tx *gorm.DB, paymentHash, username, id string, amount int64) error {
	err := tx.Model(&Invoice{}).Where("payment_hash = ?", paymentHash).
		Select("sender_id", "sender_username", "amount", "settled").
		Updates(&Invoice{
			SenderID:       &id,
			SenderUsername: &username,
			Amount:         uint64(amount),
			Settled:        time.Now().UTC(),
		}).Error
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": id,
		}).Error(MsgUpdateInvoiceSenderFailed)
		return fmt.Errorf("%s. Reason: %v", MsgUpdateInvoiceSenderFailed, ErrInternal)
	} else {
		return nil
	}
}
