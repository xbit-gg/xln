package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Withdraw struct {
	K1 string `gorm:"primaryKey" sql:"type:uuid"`

	CreatedAt time.Time
	UpdatedAt time.Time

	WalletID string `gorm:"index,priority:1"`
	Username string `gorm:"index,priority:2"`
	Wallet   Wallet `gorm:"foreignKey:username,wallet_id;references:username,id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	MinMsat     uint64
	MaxMsat     uint64
	MaxUse      uint
	Uses        uint
	Description string

	Expiry time.Time
}

func (r *repository) CreateWithdraw(tx *gorm.DB, withdraw *Withdraw) (*Withdraw, error) {
	if withdraw == nil {
		return nil, fmt.Errorf("%s. Reason: %v", MsgCreateWithdrawFailed, MsgReceivedNil)
	} else if err := tx.Create(withdraw).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"description": withdraw.Description,
			"user":        withdraw.Username,
			"wallet":      withdraw.WalletID,
			"withdraw":    withdraw.K1,
		}).Error(MsgCreateWithdrawFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgCreateWithdrawFailed, ErrInternal)
	} else {
		return withdraw, nil
	}
}

func (r *repository) GetWalletWithdraw(tx *gorm.DB, username, walletId, k1 string, includeExpired bool) (*Withdraw, error) {
	var (
		err      error
		withdraw Withdraw
	)
	if !includeExpired {
		// WHERE k1 = ? AND (expiry >= ? OR expiry == ?)
		err = tx.Take(&withdraw, "k1 = ? AND username = ? AND wallet_id = ? AND (expiry >= ? OR expiry = ?)",
			k1, username, walletId, time.Now().UTC(), time.Time{}.UTC()).Error
	} else {
		err = tx.Take(&withdraw, "k1 = ? AND username = ? AND wallet_id = ?", k1, username, walletId).Error
	}
	if err == gorm.ErrRecordNotFound {
		return nil, ErrWithdrawNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"k1": k1,
		}).Error(MsgGetWithdrawFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetWithdrawFailed, ErrInternal)
	} else {
		return &withdraw, nil
	}
}

func (r *repository) GetWithdraw(tx *gorm.DB, k1 string, includeExpired bool) (*Withdraw, error) {
	var (
		err      error
		withdraw Withdraw
	)
	if !includeExpired {
		err = tx.Take(&withdraw, "k1 = ? AND (expiry >= ? OR expiry = ?)", k1, time.Now().UTC(), time.Time{}).Error
	} else {
		err = tx.Take(&withdraw, "k1 = ?", k1).Error
	}
	if err == gorm.ErrRecordNotFound {
		return nil, ErrWithdrawNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"k1": k1,
		}).Error(MsgGetWithdrawFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetWithdrawFailed, ErrInternal)
	} else {
		return &withdraw, nil
	}
}

func (r *repository) IncrementWithdrawCount(tx *gorm.DB, k1 string) error {
	withdraw := Withdraw{}
	if res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&withdraw).
		Where("k1 = ?", k1).
		UpdateColumn("uses", gorm.Expr("uses + ?", 1)); res.Error != nil {
		log.WithError(res.Error).WithFields(log.Fields{}).Error(MsgIncrementWalletFailed)
		return fmt.Errorf("%s. Reason: %v", MsgIncrementWalletFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrWalletNotFound
	} else {
		return nil
	}
}
