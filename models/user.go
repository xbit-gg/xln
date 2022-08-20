package models

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type User struct {
	Username  string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	ApiKey    string         `gorm:"size:44;unique;<-:create"`

	Wallets []Wallet `gorm:"foreignKey:Username;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// ln auth
	LinkKey   *string `gorm:"index;unique"`
	LinkLabel *string
}

func (r *repository) CreateUser(tx *gorm.DB, user *User) error {
	if err := tx.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), gormMsgSubstrUniqueConstraintFailed) ||
			strings.Contains(err.Error(), gormMsgSubstrDuplicateKey) {
			return fmt.Errorf("%s. Reason: %v", MsgCreateUserFailed, MsgDuplicateUsername)
		} else {
			log.WithError(err).WithField("user", user.Username).Error(MsgCreateUserFailed)
			return fmt.Errorf("%s. Reason: %v", MsgCreateUserFailed, ErrInternal)
		}
	} else {
		return nil
	}
}

func (r *repository) DeleteUser(tx *gorm.DB, username string) error {
	if res := tx.Where("username = ?", username).Delete(&User{}); res.Error != nil {
		log.WithError(res.Error).WithField("user", username).Error(MsgDeleteUserFailed)
		return fmt.Errorf("%s. Reason: %v", MsgDeleteUserFailed, ErrInternal)
	} else if res.RowsAffected == 0 {
		return ErrUserNotFound
	} else {
		log.WithFields(log.Fields{
			"user": username,
		}).Info("user deleted")
		return nil
	}
}

func (r *repository) GetUser(tx *gorm.DB, username string) (*User, error) {
	user := &User{}
	if err := tx.First(user, "username = ?", username).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrUserNotFound
	} else if err != nil {
		log.WithError(err).WithField("user", username).Error(MsgGetUserFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetUserFailed, ErrInternal)
	} else {
		return user, nil
	}
}

func (r *repository) UserExists(tx *gorm.DB, username string) (bool, error) {
	if _, err := r.GetUser(tx, username); err == gorm.ErrRecordNotFound {
		return false, err
	} else if err != nil {
		log.WithError(err).WithField("user", username).Error(MsgUserExistsFailed)
		return false, fmt.Errorf("%s. Reason: %v", MsgUserExistsFailed, ErrInternal)
	} else {
		return true, nil
	}
}

func (r *repository) GetUserWithApiKey(tx *gorm.DB, apiKey string) (*User, error) {
	user := &User{}
	if err := tx.First(user, "api_key = ?", apiKey).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrUserNotFound
	} else if err != nil {
		log.WithError(err).Error(MsgGetUserWithApiKeyFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetUserWithApiKeyFailed, ErrInternal)
	} else {
		return user, nil
	}
}

func (r *repository) GetLinkedUser(tx *gorm.DB, key string) (*User, error) {
	user := &User{}
	if err := tx.Take(&user, "link_key = ?", key).Error; err == gorm.ErrRecordNotFound {
		return nil, ErrLinkedUserNotFound
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"key": key,
		}).Error(MsgGetLinkedUserFailed)
		return nil, fmt.Errorf("%s. Reason: %v", MsgGetLinkedUserFailed, ErrInternal)
	} else {
		return user, nil
	}
}

func (r *repository) UpdateUserLink(tx *gorm.DB, username string, key, label *string) error {
	if label != nil && key == nil {
		log.WithError(ErrInvalidParams).WithFields(log.Fields{
			"user":      username,
			"linkKey":   key,
			"linkLabel": *label,
		}).Error(MsgCannotHaveLabelForNilValue)
		return ErrInvalidParams
	}
	if err := tx.Model(&User{}).Where("username = ?", username).
		Updates(User{LinkKey: key, LinkLabel: label}).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":      username,
			"linkKey":   key,
			"linkLabel": label,
		}).Error(MsgUpdateUserFailed)
		return fmt.Errorf("%s. Reason: %v", MsgUpdateUserFailed, ErrInternal)
	} else {
		return nil
	}
}
