package user

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/gorm"
)

var (
	errDeletingUserMsg = "could not delete user. Reason: %v"
)

type Manager interface {

	// CreateUser creates a user.
	CreateUser(username string) (*models.User, error)

	// DeleteUser deletes a user.
	DeleteUser(username string, isPostgres bool) error

	// GetUser returns a user with corresponding username if it exists. Otherwise it errors and returns nil
	GetUser(username string) (*models.User, error)

	// ListUsers lists all the users.
	ListUsers() ([]*models.User, error)

	// ListUserTransactions gets the list of transactions for a user with username within the
	// provided time range. Start time is by default the earliest date retrievable. End time is
	// by default the current time. If offset is nonzero, then the search results will skip that amount;
	// if limit is nonzero, then the number of results will be at most equal to the limit;
	// if descending is false then results will be returned chronollogically by creation time.
	// Returns: the corresponding list of transactions; the next index of the search result sequence;
	// the total number of records that belong to the user in the given time frame.
	ListUserTransactions(username string, startTime time.Time, endTime time.Time, offset uint,
		limit uint, descending bool) (txns []*models.Transaction, nextOffset int, total uint64, err error)

	// GetUserWithApiKey gets user with api key
	GetUserWithApiKey(apiKey string) (*models.User, error)
}

type manager struct {
	db *db.DB
}

func NewManager(db *db.DB) Manager {
	return &manager{db: db}
}

func (m *manager) CreateUser(username string) (*models.User, error) {
	user := &models.User{}
	if username != "" {
		user.Username = username
	}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if err := m.db.Repo.CreateUser(tx, user); err != nil {
			return err
		}
		wallet := &models.Wallet{ID: username, Name: &username, Username: username}
		if err := m.db.Repo.CreateWallet(tx, wallet); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": user.Username,
			}).Error("failed to create default wallet when attempting to create new user")
			return err
		}
		user.Wallets = append(user.Wallets, *wallet)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *manager) DeleteUser(username string, isPostgres bool) error {
	exists, err := m.db.Repo.UserExists(m.db.DB, username)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": username,
		}).Info("could not delete user")
		return fmt.Errorf(errDeletingUserMsg, err)
	} else if !exists {
		return models.ErrUserNotFound
	}
	err = m.db.Transaction(func(tx *gorm.DB) error {
		if err := m.db.Repo.DeleteUser(tx, username); err != nil {
			return fmt.Errorf("error when deleting user from database. Reason: %v", err)
		} else if err := m.db.Repo.DeleteUserWallets(tx, username); err != nil {
			return fmt.Errorf("error when deleting user's wallets from database. Reason: %v", err)
		}
		// sqlite does not cascade constraints due to users model soft deletion mode
		if !isPostgres {
			if err := m.db.Repo.NullifyInvoiceRecipient(tx, username, ""); err != nil {
				return err
			}
			if err := m.db.Repo.NullifyInvoiceSender(tx, username, ""); err != nil {
				return err
			}
			if err := m.db.Repo.NullifyTransactionRecipient(tx, username, ""); err != nil {
				return err
			}
			if err := m.db.Repo.NullifyInvoiceSender(tx, username, ""); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (m *manager) GetUser(username string) (*models.User, error) {
	return m.db.Repo.GetUser(m.db.DB, username)
}

func (m *manager) ListUsers() ([]*models.User, error) {
	var users []*models.User
	res := m.db.Find(&users)
	return users, res.Error
}

func (m *manager) ListUserTransactions(username string, startTime time.Time, endTime time.Time, offset uint,
	limit uint, descending bool) (txns []*models.Transaction, nextOffset int, total uint64, err error) {
	if startTime.Unix() >= endTime.Unix() {
		err = errors.New("invalid time range")
		return
	}
	txns, nextOffset, total, err = m.db.Repo.ListUserTransactions(m.db.DB, username, startTime, endTime, offset, limit, descending)
	return
}

func (m *manager) GetUserWithApiKey(apiKey string) (*models.User, error) {
	if user, err := m.db.Repo.GetUserWithApiKey(m.db.DB, apiKey); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}
