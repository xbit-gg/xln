package models

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/xbit-gg/xln/util"
	"gorm.io/gorm"
)

func (user *User) BeforeCreate(tx *gorm.DB) error {
	if user.Username == "" {
		id, err := createUUID()
		if err != nil {
			tx.Logger.Error(tx.Statement.Context, "Failed to create user because UUID could not be generated")
			return err
		}
		tx.Statement.SetColumn("Username", id)
	}
	key, err := util.GenKey()
	if err != nil {
		tx.Logger.Error(tx.Statement.Context,
			"Failed to create user because key could not be generated")
		return err
	}
	tx.Statement.SetColumn("ApiKey", key)
	return nil
}

func (wallet *Wallet) BeforeCreate(tx *gorm.DB) error {
	var id string
	if wallet.ID == "" {
		var err error
		id, err = createUUID()
		if err != nil {
			tx.Logger.Error(tx.Statement.Context,
				"Failed to create wallet because UUID could not be generated")
			return err
		}
		tx.Statement.SetColumn("ID", id)

	} else {
		id = wallet.ID
	}
	if wallet.Name == nil {
		tx.Statement.SetColumn("Name", id)
	}
	key, err := util.GenKey()
	if err != nil {
		tx.Logger.Error(tx.Statement.Context, "Failed to create wallet because key could not be generated")
		return err
	}
	tx.Statement.SetColumn("ApiKey", key)
	return nil
}

func (wallet *Transaction) BeforeCreate(tx *gorm.DB) error {
	id, err := createUUID()
	if err != nil {
		tx.Logger.Error(tx.Statement.Context, "Failed to create transaction because UUID could not be generated")
		return err
	}
	tx.Statement.SetColumn("ID", id)
	return nil
}

func (auth *Auth) BeforeCreate(tx *gorm.DB) error {
	// auth must belong to strictly either wallet or user.
	if auth.UserUsername == auth.WalletID && auth.WalletUsername == auth.UserUsername {
		tx.Logger.Error(tx.Statement.Context, "invalid auth struct")
		return errors.New("invalid auth struct")
	}
	return nil
}

func (user *Withdraw) BeforeCreate(tx *gorm.DB) error {
	k1, err := util.GenURLRandStr(32)
	if err != nil {
		tx.Logger.Error(tx.Statement.Context,
			"Failed to create withdraw because rand id could not be generated")
		return err
	}
	tx.Statement.SetColumn("k1", k1)
	return nil
}

func createUUID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
