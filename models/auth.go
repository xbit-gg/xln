package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Auth struct {
	K1 string `gorm:"primaryKey"`

	// belongs to strictly either wallet or user.
	UserUsername   *string
	User           *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WalletUsername *string
	WalletID       *string
	Wallet         *Wallet `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Link   bool
	Expiry time.Time
	Authed bool
	Label  string
}

func (r *repository) CreateAuth(tx *gorm.DB, auth *Auth) error {
	if err := tx.Create(auth).Error; err != nil {
		if auth.WalletID != nil {
			log.WithError(err).WithFields(log.Fields{
				"k1":     auth.K1,
				"user":   *auth.UserUsername,
				"wallet": *auth.WalletID,
			}).Error(MsgCreateAuthFailed)
			return ErrInternal
		} else {
			log.WithError(err).WithFields(log.Fields{
				"k1":   auth.K1,
				"user": auth.UserUsername,
			}).Error(MsgCreateAuthFailed)
			return ErrInternal
		}
	} else {
		return nil
	}
}

func (r *repository) Authenticate(tx *gorm.DB, k1 string) error {
	if res := tx.Model(&Auth{}).Where("k1 = ?", k1).UpdateColumn("authed", true); res.Error != nil {
		log.WithError(res.Error).WithField("k1", k1).Error(MsgAuthFailed)
		return ErrInternal
	} else if res.RowsAffected == 0 {
		log.WithError(ErrAuthNotFound).WithField("k1", k1).Error(MsgAuthFailed)
		return ErrAuthNotFound
	} else {
		return nil
	}
}

func (r *repository) GetAuth(tx *gorm.DB, k1 string) (*Auth, error) {
	var auth *Auth
	now := time.Now().UTC()
	res := tx.Where("k1 = ? AND expiry >= ?", k1, now).First(&auth)
	if res.Error == gorm.ErrRecordNotFound {
		return auth, fmt.Errorf("invalid auth k1 '%s'", k1)
	}
	return auth, res.Error
}
