package models

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Transaction struct {
	ID        string         `gorm:"primaryKey" sql:"type:uuid"`
	CreatedAt time.Time      // time when transaction was created
	UpdatedAt time.Time      // time when transaction was last updated
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// if From and To are both not Nil then it is an internal transfer
	// From and To are never both Nil
	FromID       *string `gorm:"index,priority:4"`
	FromUsername *string `gorm:"index,priority:5"`
	From         *Wallet `gorm:"foreignKey:from_id,from_username;references:id,username;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ToID         *string `gorm:"index,priority:6"`
	ToUsername   *string `gorm:"index,priority:7"`
	To           *Wallet `gorm:"foreignKey:to_id,to_username;references:id,username;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Label    *string
	Amount   uint64
	FeesPaid uint64

	InvoiceID *string
	Invoice   *Invoice // if Nil then it is an internal transfer
}

func (r *repository) CreateTransaction(tx *gorm.DB, transaction *Transaction) error {
	if transaction == nil {
		return fmt.Errorf("%s. Reason: %v", MsgCreateTransactionFailed, MsgReceivedNil)
	} else if err := tx.Create(transaction).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"transaction.amount": transaction.Amount,
			"invoice":            transaction.InvoiceID,
			"fromId":             transaction.FromID,
			"fromUsername":       transaction.FromUsername,
			"toId":               transaction.ToID,
			"toUsername":         transaction.ToUsername,
		}).Error(MsgCreateTransactionFailed)
		return fmt.Errorf("%s. Reason: %v", MsgCreateTransactionFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) CreateTransactions(tx *gorm.DB, transactions *[]Transaction) error {
	if transactions == nil {
		return fmt.Errorf(
			"%s. Reason: %v",
			MsgCreateTransactionsFailed,
			"expected transaction records. Received nil instead.")
	}
	if err := tx.Create(transactions).Error; err != nil {
		numTxns := len(*transactions)
		for i, txn := range *transactions {
			log.WithError(err).WithFields(log.Fields{
				"transaction.amount": txn.Amount,
				"invoice":            txn.InvoiceID,
				"from":               txn.FromID,
				"to":                 txn.To,
				"i":                  i,
				"totalTranscations":  numTxns,
			}).Error(MsgCreateTransactionFailed)
		}
		log.WithError(err).WithField("totalTranscations", numTxns).Error(MsgCreateTransactionsFailed)
		return fmt.Errorf("%s. Reason: %v", MsgCreateTransactionsFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) GetTransaction(tx *gorm.DB, transactionId string) (*Transaction, error) {
	transaction := Transaction{}
	err := tx.Preload("Invoice").
		Where("id = ?", transactionId).Take(&transaction).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrTransactionNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"transaction": transactionId,
		}).Error(MsgGetTransactionFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetTransactionFailed, ErrInternal)
	} else {
		return &transaction, nil
	}
}

func (r *repository) GetWalletTransaction(tx *gorm.DB, username, walletId, transactionId string) (*Transaction, error) {
	transaction := Transaction{}
	err := tx.Preload("Invoice").
		Where("id = ?", transactionId).
		Where(
			tx.Where(
				"from_id = ? AND from_username = ?", walletId, username).
				Or("to_id = ? AND to_username = ?", walletId, username),
		).Take(&transaction).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrTransactionNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet":      walletId,
			"user":        username,
			"transaction": transactionId,
		}).Error(MsgGetTransactionFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetTransactionFailed, ErrInternal)
	} else {
		return &transaction, nil
	}
}

func (r *repository) ListWalletTransactions(tx *gorm.DB, username, walletId string, startTime, endTime time.Time,
	offset, limit uint, descending bool) (txns []*Transaction, nextOffset int, total uint64, err error) {
	var dbLimit int
	if limit == 0 { // cancel the gorm limit
		dbLimit = -1
	} else {
		dbLimit = int(limit)
	}
	order := ""
	if descending {
		order = " desc"
	}

	res := tx.Order("created_at"+order).
		Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Where(
			tx.Where(
				"from_id = ? AND from_username = ?", walletId, username).
				Or("to_id = ? AND to_username = ?", walletId, username),
		).Offset(int(offset)).Limit(dbLimit).Find(&txns)
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet":     walletId,
			"user":       username,
			"startTime":  startTime,
			"endTime":    endTime,
			"offset":     offset,
			"limit":      limit,
			"descending": descending,
		}).Error(MsgListWalletTransactionsFailed)
		return nil, 0, 0, fmt.Errorf("%s. Reason: %v", MsgListWalletTransactionsFailed, ErrInternal)
	} else if offset+limit >= math.MaxUint32 {
		return nil, 0, 0, ErrMaxNumResultsExceeded
	}

	// Count currently does not work with offset/limit and just returns 0
	var dbTotal int64
	err = tx.Model(Transaction{}).
		Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Where(
			tx.Where(
				"from_id = ? AND from_username = ?", walletId, username).
				Or("to_id = ? AND to_username = ?", walletId, username),
		).Count(&dbTotal).Error
	total = uint64(dbTotal)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet":     walletId,
			"user":       username,
			"startTime":  startTime,
			"endTime":    endTime,
			"offset":     offset,
			"limit":      limit,
			"descending": descending,
		}).Error(MsgListWalletTransactionsFailed)
		return nil, 0, 0, fmt.Errorf("%s. Reason: %v", MsgListWalletTransactionsFailed, ErrInternal)
	} else if limit == 0 { // return results as is. There is no nextOffset/ there is no sequence.
		return txns, 0, total, nil
	} else if uint64(offset+limit) > total || uint64(len(txns)) == total { // end of sequence
		return txns, -1, total, nil
	} else { // set index of next record in sequence
		return txns, int(offset + limit), total, nil
	}
}

func (r *repository) ListUserTransactions(tx *gorm.DB, username string, startTime, endTime time.Time,
	offset, limit uint, descending bool) (txns []*Transaction, nextOffset int, total uint64, err error) {
	var dbLimit int
	if limit == 0 { // cancel the gorm limit
		dbLimit = -1
	} else {
		dbLimit = int(limit)
	}
	order := ""
	if descending {
		order = " desc"
	}

	res := tx.Order("created_at"+order).
		Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Where(
			tx.Where("from_username = ?", username).Or("to_username = ?", username),
		).Limit(dbLimit).Offset(int(offset)).Find(&txns)
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"user":       username,
			"startTime":  startTime,
			"endTime":    endTime,
			"offset":     offset,
			"limit":      limit,
			"descending": descending,
		}).Error(MsgListUserTransactionsFailed)
		return nil, 0, 0, fmt.Errorf("%s. Reason: %v", MsgListUserTransactionsFailed, ErrInternal)
	} else if offset+limit >= math.MaxUint32 {
		return nil, 0, 0, ErrMaxNumResultsExceeded
	}
	// Count currently does not work with offset/limit and just returns 0
	var dbTotal int64
	err = tx.Model(Transaction{}).
		Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Where(
			tx.Where("from_username = ?", username).Or("to_username = ?", username),
		).Count(&dbTotal).Error
	total = uint64(dbTotal)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":       username,
			"startTime":  startTime,
			"endTime":    endTime,
			"offset":     offset,
			"limit":      limit,
			"descending": descending,
		}).Error(MsgListUserTransactionsFailed)
		return nil, 0, 0, fmt.Errorf("%s. Reason: %v", MsgListUserTransactionsFailed, ErrInternal)
	} else if limit == 0 { // return results as is. There is no nextOffset/ there is no sequence.
		return txns, 0, total, nil
	} else if uint64(offset+limit) > total || uint64(len(txns)) == total { // end of sequence
		return txns, -1, total, nil
	} else { // set index of next record in sequence
		return txns, int(offset + limit), total, nil
	}
}

func (r *repository) NullifyTransactionRecipient(tx *gorm.DB, username, id string) error {
	var err error
	if id == "" {
		err = tx.Model(&Transaction{}).Where("to_username = ?", username).Select("to_id", "to_username").
			Updates(&Transaction{ToID: nil, ToUsername: nil}).Error
	} else {
		err = tx.Model(&Transaction{}).Where("to_username = ? AND to_id = ?", username, id).Select("to_id", "to_username").
			Updates(&Transaction{ToID: nil, ToUsername: nil}).Error
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": id,
		}).Error(MsgNullifyTranscationRecipientFailed)
		return fmt.Errorf("%s. Reason: %v", MsgNullifyTranscationRecipientFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) NullifyTransactionSender(tx *gorm.DB, username, id string) error {
	var err error
	if id == "" {
		err = tx.Model(&Transaction{}).Where("from_username = ?", username).Select("from_id", "from_username").
			Updates(&Transaction{FromID: nil, FromUsername: nil}).Error
	} else {
		err = tx.Model(&Transaction{}).Where("from_username = ? AND from_id = ?", username, id).Select("from_id", "from_username").
			Updates(&Transaction{FromID: nil, FromUsername: nil}).Error
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": id,
		}).Error(MsgNullifyTransactionSenderFailed)
		return fmt.Errorf("%s. Reason: %v", MsgNullifyTransactionSenderFailed, ErrInternal)
	} else {
		return nil
	}
}
