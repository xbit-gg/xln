package models

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Wallet struct {
	ID       string `gorm:"primaryKey" sql:"type:uuid"`
	Username string `gorm:"primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Name   *string
	ApiKey string `gorm:"size:44;unique"`

	Balance      uint64
	Transactions []Transaction `gorm:"many2many:wallet_transactions;constraint:OnDelete:CASCADE"`

	Locked bool

	// ln auth
	LinkKey   *string `gorm:"index;unique"`
	LinkLabel *string
}

type WalletOptions struct {
	Name    *string
	Locked  *bool
	Balance *uint64
}

func (r *repository) CreateWallet(tx *gorm.DB, wallet *Wallet) error {
	if wallet == nil {
		return fmt.Errorf("%s. Reason: %v", MsgCreateWalletFailed, MsgReceivedNil)
	} else if err := tx.Create(wallet).Error; err != nil {
		if strings.Contains(err.Error(), gormMsgSubstrUniqueConstraintFailed) ||
			strings.Contains(err.Error(), gormMsgSubstrDuplicateKey) {
			log.WithError(err).WithFields(log.Fields{
				"wallet":      wallet.ID,
				"wallet.Name": *wallet.Name,
				"user":        wallet.Username,
			}).Info(MsgCreateWalletFailed)
			return fmt.Errorf("%s. Reason: %v", MsgCreateWalletFailed, MsgWalletIDAlreadyExists)
		} else {
			log.WithError(err).WithFields(log.Fields{
				"wallet":      wallet.ID,
				"wallet.Name": *wallet.Name,
				"user":        wallet.Username,
			}).Error(MsgCreateWalletFailed)
			return fmt.Errorf("%s. Reason: %v", MsgCreateWalletFailed, ErrInternal)
		}
	} else {
		return nil
	}
}

func (r *repository) DeleteWalletZeroBalance(tx *gorm.DB, username, walletId string) error {
	if res := tx.Where("username = ? AND id = ? AND balance = 0", username, walletId).Delete(&Wallet{}); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgWalletDeleteFailed)
		return fmt.Errorf("%s. Reason: %v", MsgWalletDeleteFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		if _, err := r.GetWallet(tx, username, walletId); err != nil {
			return err
		} else {
			return ErrNonZeroBalance
		}
	} else {
		return nil
	}
}

func (r *repository) DeleteWallet(tx *gorm.DB, username, walletId string) error {
	if res := tx.Where("username = ? AND id = ?", username, walletId).Delete(&Wallet{}); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgWalletDeleteFailed)
		return fmt.Errorf("%s. Reason: %v", MsgWalletDeleteFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrWalletNotFound
	} else {
		return nil
	}
}

func (r *repository) DeleteUserWallets(tx *gorm.DB, username string) error {
	wallets := []Wallet{}
	if res := tx.Where("username = ?", username).Delete(&wallets); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"user": username,
		}).Error(MsgMultipleWalletDeleteFailed)
		return fmt.Errorf("%s. Reason: %v", MsgMultipleWalletDeleteFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrUserNotFound
	} else {
		return nil
	}
}

func (r *repository) GetWallet(tx *gorm.DB, username, walletId string) (*Wallet, error) {
	wallet := Wallet{}
	if err := tx.Take(&wallet, "username = ? AND id = ?", username, walletId).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrWalletNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet":   walletId,
			"username": username,
		}).Error(MsgGetWalletFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetWalletFailed, ErrInternal)
	} else {
		return &wallet, nil
	}
}

func (r *repository) ListUserWallets(tx *gorm.DB, username string) ([]*Wallet, error) {
	var wallets []*Wallet
	if res := tx.Find(&wallets, "username = ?", username); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"user": username,
		}).Error(MsgListWalletFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListWalletFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return nil, ErrUserNotFound
	} else {
		return wallets, nil
	}
}

func (r *repository) UpdateWalletOptions(tx *gorm.DB, username, walletId string, walletOptions *WalletOptions) error {
	wallet := &Wallet{}
	updateAttributes := make(map[string]interface{})
	if walletOptions.Name != nil {
		updateAttributes["name"] = *walletOptions.Name
	}
	if walletOptions.Locked != nil {
		updateAttributes["locked"] = *walletOptions.Locked
	}
	if walletOptions.Balance != nil {
		updateAttributes["balance"] = *walletOptions.Balance
	}
	if res := tx.Model(wallet).Where("username = ? AND id = ?", username, walletId).Updates(updateAttributes); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgUpdateWalletFailed)
		return fmt.Errorf("%s. Reason: %v", MsgUpdateWalletFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrWalletNotFound
	} else {
		return nil
	}
}

func (r *repository) UpdateWalletWithBalance(tx *gorm.DB, username string, walletId string, newBalance uint64) (*Wallet, error) {
	wallet := Wallet{}
	if res := tx.Model(&wallet).Where("username = ? AND id = ?", username, walletId).UpdateColumn("balance", newBalance); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet":     walletId,
			"user":       username,
			"newBalance": newBalance,
		}).Error(MsgUpdateWalletBalanceFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgUpdateWalletBalanceFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return nil, ErrWalletNotFound
	} else {
		return &wallet, nil
	}
}

func (r *repository) IncrementWalletBalance(tx *gorm.DB, username, walletId string, deltaBalance uint64) error {
	wallet := Wallet{}
	if res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&wallet).Where("username = ? AND id = ?", username, walletId).UpdateColumn("balance", gorm.Expr("balance + ?", deltaBalance)); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet":       walletId,
			"user":         username,
			"deltaBalance": deltaBalance,
		}).Error(MsgIncrementWalletFailed)
		return fmt.Errorf("%s. Reason: %v", MsgIncrementWalletFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrWalletNotFound
	} else {
		return nil
	}
}

func (r *repository) DecrementWalletBalance(tx *gorm.DB, username string, walletId string, deltaBalance uint64) error {
	var (
		pendingPayments     []*PendingPayment
		pendingPaymentTotal uint64
	)
	// get pending payments while locking rows
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("amount").Where("wallet_id = ?", walletId).Find(&pendingPayments).Error
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet":       walletId,
			"user":         username,
			"deltaBalance": deltaBalance,
		}).Errorf("%s. Details: %s", MsgDecrementWalletFailed, MsgListWalletPendingPaymentsFailed)
		return fmt.Errorf("%s. Reason: %v", MsgDecrementWalletFailed, ErrInternal)
	}
	for _, pendingPayment := range pendingPayments {
		pendingPaymentTotal += pendingPayment.Amount
	}
	minBalance := deltaBalance + pendingPaymentTotal
	// update wallet
	wallet := Wallet{}
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&wallet).Where("username = ? AND id = ? AND balance >= ?", username, walletId, minBalance).UpdateColumn("balance", gorm.Expr("balance - ?", deltaBalance))
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet":       walletId,
			"user":         username,
			"deltaBalance": deltaBalance,
		}).Errorf("%s. Details: %s", MsgDecrementWalletFailed, MsgUpdateWalletBalanceFailed)
		return fmt.Errorf("%s. Reason: %v", MsgDecrementWalletFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return fmt.Errorf("%s. Reason: %v", MsgDecrementWalletFailed, ErrInsufficientBalance)
	}
	return nil
}

func (r *repository) WalletsFromUser(tx *gorm.DB, username string, walletIds []string) (bool, []Wallet, error) {
	var (
		wallets     []Wallet
		walletCount int64
	)
	err := tx.Where("username = ? AND id IN ?", username, walletIds).Find(&wallets).Count(&walletCount).Error
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallets": walletIds,
			"user":    username,
		}).Error(MsgDeterminingIfWalletsFromUserFailed)
		return false, nil, fmt.Errorf("%s. Reason %v", MsgDeterminingIfWalletsFromUserFailed, ErrInternal)
	} else if int64(len(walletIds)) == walletCount {
		return true, wallets, nil
	} else {
		return false, wallets, nil
	}
}

func (r *repository) LockWalletRecordForUpdate(tx *gorm.DB, username string, walletId string) (*Wallet, error) {
	wallet := Wallet{}
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&wallet, "username = ? AND id = ?", username, walletId)
	if res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgLockWalletRecordFailed)
		return nil, res.Error // don't add error context. this method is not user-facing
	} else if res.RowsAffected == 0 {
		return nil, ErrWalletNotFound
	}
	if wallet.Locked {
		return &wallet, ErrCannotUpdateLockedWallet
	}
	return &wallet, nil
}

func (r *repository) GetWalletWithApiKey(tx *gorm.DB, apiKey string) (*Wallet, error) {
	wallet := &Wallet{}
	if err := tx.First(wallet, "api_key = ?", apiKey).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrWalletNotFound
	} else if err != nil {
		log.WithError(err).Error(MsgGetUserWithApiKeyFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetUserWithApiKeyFailed, ErrInternal)
	} else {
		return wallet, nil
	}
}

func (r *repository) UpdateWalletLink(tx *gorm.DB, username, walletId string, key, label *string) error {
	if label != nil && key == nil {
		log.WithError(ErrInvalidParams).WithFields(log.Fields{
			"user":      username,
			"wallet":    walletId,
			"linkKey":   key,
			"linkLabel": *label,
		}).Error(MsgCannotHaveLabelForNilValue)
		return ErrInvalidParams
	}
	if err := tx.Model(&Wallet{}).Where("username = ? AND id = ?", username, walletId).
		Updates(Wallet{LinkKey: key, LinkLabel: label}).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":      username,
			"wallet":    walletId,
			"linkKey":   key,
			"linkLabel": label,
		}).Error(MsgUpdateWalletFailed)
		return fmt.Errorf("%s. Reason: %v", MsgUpdateWalletFailed, ErrInternal)
	} else {
		return nil
	}
}

func (r *repository) GetLinkedWallet(tx *gorm.DB, key string) (*Wallet, error) {
	wallet := &Wallet{}
	if err := tx.Take(&wallet, "link_key = ?", key).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrLinkedWalletNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"key": key,
		}).Error(MsgGetLinkedWalletFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetLinkedWalletFailed, ErrInternal)
	} else {
		return wallet, nil
	}
}

func (r *repository) ListUserLinkedWallets(tx *gorm.DB, username string) ([]*Wallet, error) {
	var wallet []*Wallet
	if err := tx.Find(&wallet, "username = ? AND link_key IS NOT NULL", username).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": username,
		}).Error(MsgListLinkedWalletFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgListLinkedWalletFailed, ErrInternal)
	} else {
		return wallet, nil
	}
}

func (r *repository) GetConfirmedBalance(tx *gorm.DB, username, walletId string) (uint64, error) {
	wallet, err := r.LockWalletRecordForUpdate(tx, username, walletId)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgGetConfirmedBalanceFailed)
		return 0, fmt.Errorf("%s. Reason: %v", MsgGetConfirmedBalanceFailed, err)
	}
	confirmedBalance := wallet.Balance
	pendingPayments, err := r.ListWalletPendingPayments(tx, wallet.Username, walletId)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error(MsgGetConfirmedBalanceFailed)
		return 0, fmt.Errorf("%s. Reason: %v", MsgGetConfirmedBalanceFailed, err)
	}
	for _, pp := range pendingPayments {
		confirmedBalance -= pp.Amount
	}

	return confirmedBalance, err
}
