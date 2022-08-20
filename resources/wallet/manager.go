package wallet

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/gorm"
)

type Manager interface {
	// CreateWallet creates a wallet.
	// Does not allow duplicate wallet names for a given user.
	CreateWallet(username, id, name string) (*models.Wallet, error)

	// DeleteWallet deletes wallet with walletId if wallet has zero balance
	DeleteWallet(username, walletId string, isPostgres bool) error

	// DeleteWallet deletes wallet with walletId
	AdminDeleteWallet(username, walletId string, isPostgres bool) error

	// UpdateWalletOptions update wallet with select wallet options
	UpdateWalletOptions(username, walletId string, walletOptions *models.WalletOptions) error

	// GetWallet returns the wallet matching walletId.
	// Errors if there is no matching wallet.
	GetWallet(username, walletId string) (*models.Wallet, error)

	// GetTransaction gets a transaction for a given wallet
	GetTransaction(username, walletId string, transactionId string) (*models.Transaction, error)

	// ListWalletTransactions gets the list of transactions for a user with username within the
	// provided time range. Start time is by default the earliest date retrievable. End time is
	// by default the current time. If offset is nonzero, then the search results will skip that amount;
	// if limit is nonzero, then the number of results will be at most equal to the limit;
	// if descending is false then results will be returned chronollogically by creation time.
	// Returns: the corresponding list of transactions; the next index of the search result sequence;
	// the total number of records that belong to the user in the given time frame.
	ListWalletTransactions(username, walletId string, startTime, endTime time.Time, offset, limit uint,
		descending bool) (txns []*models.Transaction, nextOffset int, total uint64, err error)

	// ListWallets lists all the wallets for a given user
	ListWallets(username string) ([]*models.Wallet, error)

	// Transfer transfers money from one wallet to another if they belong to the same user.
	Transfer(username, walletId, toWalletId string, amount uint64) (*models.Transaction, error)

	// GetWalletWithApiKey gets wallet with api key
	GetWalletWithApiKey(apiKey string) (*models.Wallet, error)
}

type manager struct {
	db *db.DB
}

func NewManager(db *db.DB) Manager {
	return &manager{db: db}
}

func (m *manager) CreateWallet(username, id, name string) (*models.Wallet, error) {
	wallet := &models.Wallet{ID: id, Username: username}
	if name != "" {
		wallet.Name = &name
	}

	err := m.db.Transaction(func(tx *gorm.DB) error {
		return m.db.Repo.CreateWallet(tx, wallet)
	})
	if err != nil {
		return nil, err
	}
	return wallet, err
}

func (m *manager) DeleteWallet(username, walletId string, isPostgres bool) error {
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if _, err := m.isUpdatable(tx, username, walletId); err != nil {
			return err
		}
		if err := m.db.Repo.DeleteWalletZeroBalance(tx, username, walletId); err != nil {
			return err
		}
		if !isPostgres {
			if err := m.db.Repo.NullifyInvoiceRecipient(tx, username, walletId); err != nil {
				return err
			}
			if err := m.db.Repo.NullifyInvoiceSender(tx, username, walletId); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (m *manager) AdminDeleteWallet(username, walletId string, isPostgres bool) error {
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if _, err := m.isUpdatable(tx, username, walletId); err != nil {
			return err
		}
		if err := m.db.Repo.DeleteWallet(tx, username, walletId); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (m *manager) UpdateWalletOptions(username, walletId string, walletOptions *models.WalletOptions) error {
	if walletOptions.Name != nil && *walletOptions.Name != "" && walletId == username {
		return errors.New("Cannot edit the name of the main wallet")
	}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		_, err := m.isUpdatable(tx, username, walletId)
		unlockingLocked := err == models.ErrCannotUpdateLockedWallet && walletOptions.Locked != nil && !*walletOptions.Locked
		if err == nil || unlockingLocked {
			err = m.db.Repo.UpdateWalletOptions(tx, username, walletId, walletOptions)
		}
		return err
	})
	return err
}

func (m *manager) GetWallet(username, walletId string) (*models.Wallet, error) {
	return m.db.Repo.GetWallet(m.db.DB, username, walletId)
}

func (m *manager) GetTransaction(username, walletId string, transactionId string) (*models.Transaction, error) {
	transaction, err := m.db.Repo.GetWalletTransaction(m.db.DB, username, walletId, transactionId)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (m *manager) ListWalletTransactions(username, walletId string, startTime, endTime time.Time, offset, limit uint,
	descending bool) (txns []*models.Transaction, nextOffset int, total uint64, err error) {
	if startTime.Unix() >= endTime.Unix() {
		err = errors.New("invalid time range")
		return
	}
	txns, nextOffset, total, err = m.db.Repo.ListWalletTransactions(m.db.DB, username, walletId, startTime, endTime, offset, limit, descending)
	return
}

func (m *manager) ListWallets(username string) ([]*models.Wallet, error) {
	wallets, err := m.db.Repo.ListUserWallets(m.db.DB, username)
	return wallets, err
}

// Transfers an amount from a wallet to another wallet of the same user.
// It returns the transaction if it was created successfully, and the error, if any.
func (m *manager) Transfer(username, walletId, toWalletId string, amount uint64) (*models.Transaction, error) {
	if amount == 0 {
		return nil, errors.New("transfer amount must be non-zero")
	}
	if walletId == toWalletId {
		return nil, errors.New("cannot transfer to the same wallet")
	}
	walletsFromSameUser, _, err := m.db.Repo.WalletsFromUser(m.db.DB, username, []string{walletId, toWalletId})
	if err != nil {
		return nil, err
	}
	if !walletsFromSameUser {
		return nil, errors.New("either wallets are not from same user or one or more wallets do not exist")
	}
	var transaction models.Transaction
	err = m.db.Transaction(func(tx *gorm.DB) error {
		// validate wallets can be updated
		_, err := m.isUpdatable(tx, username, walletId)
		if err != nil {
			return err
		}
		_, err = m.isUpdatable(tx, username, toWalletId)
		if err != nil {
			return err
		}
		// Update sending wallet balance
		err = m.db.Repo.DecrementWalletBalance(tx, username, walletId, amount)
		if err != nil {
			return err
		}

		// Update receiving wallet balance
		err = m.db.Repo.IncrementWalletBalance(tx, username, toWalletId, amount)
		if err != nil {
			return err
		}

		// Record senders and receivers transaction
		transaction = models.Transaction{
			FromID:       &walletId,
			FromUsername: &username,
			ToID:         &toWalletId,
			ToUsername:   &username,
			Amount:       amount,
			FeesPaid:     0,
		}
		if err := m.db.Repo.CreateTransaction(tx, &transaction); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"from":     walletId,
				"to":       toWalletId,
				"username": username,
			}).Error("failed to write new transaction to database")
			return err
		} else {
			return nil
		}
	})
	if err == models.ErrCannotUpdateLockedWallet {
		return nil, models.ErrCannotTransactWithLockedWallet
	} else if err != nil {
		return nil, err
	} else {
		return &transaction, nil
	}
}

// isUpdatable returns an error if the wallet is not is not updatable. Otherwise returns the wallet
func (m *manager) isUpdatable(tx *gorm.DB, username, walletId string) (*models.Wallet, error) {
	wallet, err := m.db.Repo.LockWalletRecordForUpdate(tx, username, walletId)
	if err == models.ErrCannotUpdateLockedWallet {
		log.WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Info("wallet is not updatable")
	} else if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet": walletId,
			"user":   username,
		}).Error("wallet is not updatable")
	}
	return wallet, err
}

func (m *manager) GetWalletWithApiKey(apiKey string) (*models.Wallet, error) {
	if wallet, err := m.db.Repo.GetWalletWithApiKey(m.db.DB, apiKey); err != nil {
		return nil, err
	} else {
		return wallet, nil
	}
}
