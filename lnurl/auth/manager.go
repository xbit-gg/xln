package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/gorm"
)

const (
	DefaultExpiryTime = time.Minute * 60
	LNURLAuthEndpoint = "https://%s/lnurl/auth?tag=login&k1=%s&action=%s"
)

type Manager interface {
	UserAuth(username string) (lnurl string, err error)
	WalletAuth(username string, walletId *string) (lnurl string, err error)

	UserLinkAuth(username string, label string) (lnurl string, err error)
	WalletLinkAuth(username string, walletId *string, label string) (lnurl string, err error)

	GetAuthKey(k1 string) (apikey string, err error)

	LNURLAuthenticate(k1 string, sig string, key string) error
}

type manager struct {
	hostname string
	db       *db.DB
}

func NewManager(restHost string, db *db.DB) Manager {
	return &manager{hostname: restHost, db: db}
}

func (m *manager) UserAuth(username string) (string, error) {
	return m.createAuth(username, nil, false, "")
}

func (m *manager) WalletAuth(username string, walletId *string) (string, error) {
	return m.createAuth(username, walletId, false, "")
}

func (m *manager) UserLinkAuth(username string, label string) (string, error) {
	return m.createAuth(username, nil, true, label)
}

func (m *manager) WalletLinkAuth(username string, walletId *string, label string) (string, error) {
	return m.createAuth(username, walletId, true, label)
}

func (m *manager) LNURLAuthenticate(k1 string, sig string, key string) error {
	var (
		valid bool
		err   error
	)
	if os.Getenv(cfg.EnvVarNameRuntime) == cfg.EnvValRuntimeDev {
		valid, err = func() (bool, error) {
			return sig == key, nil
		}()
	} else {
		valid, err = lnurl.VerifySignature(k1, sig, key)
	}
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid signature for key")
	}
	auth, err := m.db.Repo.GetAuth(m.db.DB, k1)
	if err != nil {
		return err
	}

	if auth.Link {
		err = m.db.Transaction(func(tx *gorm.DB) error {
			err := m.db.Repo.Authenticate(tx, k1)
			if err != nil {
				return err
			}
			if auth.UserUsername != nil { // if user
				return m.db.Repo.UpdateUserLink(tx, *auth.UserUsername, &key, &auth.Label)
			} else { // if wallet
				return m.db.Repo.UpdateWalletLink(tx, *auth.WalletID, *auth.UserUsername, &key, &auth.Label)
			}
		})
	} else {
		err = m.db.Transaction(func(tx *gorm.DB) error {
			if auth.UserUsername != nil { // if user
				if user, err := m.db.Repo.GetLinkedUser(tx, key); err == models.ErrLinkedUserNotFound {
					return errors.New("user could not be authenticated")
				} else if err != nil {
					return err
				} else if user.Username == *auth.UserUsername {
					return m.db.Repo.Authenticate(tx, k1)
				} else {
					return errors.New("user could not be authenticated")
				}
			} else { // if wallet
				if wallet, err := m.db.Repo.GetLinkedWallet(tx, key); err == models.ErrLinkedWalletNotFound {
					return errors.New("wallet could not be authenticated")
				} else if err != nil {
					return err
				} else if wallet.Username == *auth.UserUsername && wallet.ID == *auth.WalletID {
					return m.db.Repo.Authenticate(tx, k1)
				} else {
					return errors.New("wallet could not be authenticated")
				}
			}
		})
	}
	return err
}

func (m *manager) GetAuthKey(k1 string) (string, error) {
	auth, err := m.db.Repo.GetAuth(m.db.DB, k1)
	if err != nil {
		return "", err
	}
	if auth.Authed {
		if auth.UserUsername != nil { // if user
			if user, err := m.db.Repo.GetUser(m.db.DB, *auth.UserUsername); err != nil {
				return "", err
			} else {
				return user.ApiKey, nil
			}
		} else { // if wallet
			if wal, err := m.db.Repo.GetWallet(m.db.DB, *auth.UserUsername, *auth.WalletID); err != nil {
				return "", err
			} else {
				return wal.ApiKey, nil
			}
		}
	} else {
		return "", fmt.Errorf("auth token %s is unauthenticated", k1)
	}
}

func (m *manager) createAuth(username string, walletId *string, link bool, label string) (string, error) {
	if username == "" {
		return "", errors.New("username must be specified")
	}

	k1 := lnurl.RandomK1()
	var auth models.Auth
	if walletId == nil { // if wallet
		auth = models.Auth{
			K1:           k1,
			UserUsername: &username,
			Link:         link,
			Label:        label,
			Expiry:       time.Now().Add(DefaultExpiryTime).UTC(),
		}
	} else { // if user
		auth = models.Auth{
			K1:             k1,
			WalletID:       walletId,
			WalletUsername: &username,
			Link:           link,
			Label:          label,
			Expiry:         time.Now().Add(DefaultExpiryTime).UTC(),
		}
	}

	if err := m.db.Repo.CreateAuth(m.db.DB, &auth); err != nil {
		return "", err
	}

	var action string
	if link {
		action = "link"
	} else {
		action = "login"
	}

	return lnurl.LNURLEncode(fmt.Sprintf(LNURLAuthEndpoint, m.hostname, k1, action))
}
