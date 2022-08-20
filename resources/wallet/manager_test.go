package wallet

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type WalletManagerSuite struct {
	suite.Suite
	DB       *gorm.DB
	mock     sqlmock.Sqlmock
	mgr      Manager
	mockRepo MockRepo
}

func (s *WalletManagerSuite) SetupSuite() {
	var (
		sqlDB *sql.DB
		err   error
	)

	sqlDB, s.mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Require().NoError(err)
	s.DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	s.Require().NoError(err)
	s.mockRepo = MockRepo{}
	s.mgr = NewManager(&db.DB{DB: s.DB, Repo: &s.mockRepo})
}

func (s *WalletManagerSuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.mockRepo = MockRepo{}
}

func TestWalletManager(t *testing.T) {
	suite.Run(t, new(WalletManagerSuite))
}

func (s *WalletManagerSuite) TestTransfer() {
	s.Run("successfully processes valid transfer and updates wallets", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testAmount       = uint64(10)
			testUsername     = "testusername"
		)

		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			wallets := []models.Wallet{{
				ID:       testFromWalletId,
				Username: testUsername,
			}, {
				ID:       testToWalletId,
				Username: testUsername,
			}}
			return true, wallets, nil
		}

		s.mockRepo.MockDecrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testFromWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		s.mockRepo.MockIncrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testToWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		s.mockRepo.MockCreateTransaction = func(transactions *models.Transaction) error {
			s.Require().Equal(testAmount, transactions.Amount)
			s.Require().Equal(testFromWalletId, *transactions.FromID)
			s.Require().Equal(testUsername, *transactions.FromUsername)
			s.Require().Equal(testToWalletId, *transactions.ToID)
			s.Require().Equal(testUsername, *transactions.ToUsername)
			return nil
		}

		s.mockRepo.MockLockWalletRecordForUpdate = func(username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{ID: walletId, Username: testUsername}, nil
		}

		// mock start of db transaction
		s.mock.ExpectBegin()
		s.mock.ExpectCommit().WillReturnError(nil)

		// Validations
		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, testAmount)
		require.NoError(s.T(), err, "valid internal same-user wallet transfers should not error")
		s.Require().NotNil(txn, "successful transfer should return txn")
	})
	s.Run("fails when amount is zero", func() {
		txn, err := s.mgr.Transfer("wallet-id", "to-wallet-id", "testusername", uint64(0))
		s.Require().Nil(txn, "failed transfer should not return the transaction")
		s.Require().Error(err, "internal same-user wallet transfers should not allow zero amount")
	})
	s.Run("fails when wallet ids are equal", func() {

		txn, err := s.mgr.Transfer("testusername", "same-wallet-id", "same-wallet-id", uint64(10))
		s.Require().Nil(txn, "failed transfer should not return the transaction")
		s.Require().Error(err, "internal same-user wallet transfers should not require unique sender and receiver wallet ids")
	})
	s.Run("fails when user does not own receiving wallet", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testUsername     = "testusername"
		)

		// mock database return calls

		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			wallets := []models.Wallet{{
				ID:       walletIds[0],
				Username: "username-1",
			}}
			return false, wallets, nil
		}

		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, uint64(10))
		s.Require().Error(err, "internal same-user wallet transfers must be by the same user")
		s.Require().Nil(txn, "failed transfer should not return the transaction")
	})

	s.Run("fails when the check for if wallets are from same user fails", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testUsername     = "testusername"
		)

		expectedErr := errors.New("get wallet errored")
		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			return false, nil, expectedErr
		}

		// Validations
		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, uint64(10))
		s.Require().Error(err, "internal same-user wallet transfers should error when db repository fails")
		s.Require().Contains(err.Error(), expectedErr.Error())
		s.Require().Nil(txn, "failed transfer should not return the transaction")

	})

	s.Run("fails when sender's funds are insufficient", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testAmount       = uint64(10)
			testUsername     = "testusername"
		)

		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			wallets := []models.Wallet{{
				ID:       testFromWalletId,
				Username: testUsername,
			}, {
				ID:       testToWalletId,
				Username: testUsername,
			}}
			return true, wallets, nil
		}

		s.mockRepo.MockLockWalletRecordForUpdate = func(username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{ID: walletId}, nil
		}

		s.mock.ExpectBegin() // mock start of db transaction

		expectedErr := errors.New("Insufficient wallet balance")
		s.mockRepo.MockDecrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testFromWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return expectedErr
		}

		// Validations
		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, testAmount)
		s.Require().Error(err, "internal same-user wallet transfers should error when db repository fails")
		s.Require().Contains(err.Error(), expectedErr.Error())
		s.Require().Nil(txn, "failed transfer should not return the transaction")

	})

	s.Run("fails when creating the transfer transaction fails", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testAmount       = uint64(10)
			testUsername     = "testusername"
		)

		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			wallets := []models.Wallet{{
				ID:       testFromWalletId,
				Username: testUsername,
			}, {
				ID:       testToWalletId,
				Username: testUsername,
			}}
			return true, wallets, nil
		}

		s.mockRepo.MockDecrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testFromWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		s.mockRepo.MockIncrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testToWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		expectedErr := errors.New("Create wallet transaction entries failed")
		s.mockRepo.MockCreateTransaction = func(transactions *models.Transaction) error {
			s.Require().Equal(testAmount, transactions.Amount)
			s.Require().Equal(testFromWalletId, *transactions.FromID)
			s.Require().Equal(testUsername, *transactions.FromUsername)
			s.Require().Equal(testToWalletId, *transactions.ToID)
			s.Require().Equal(testUsername, *transactions.ToUsername)
			return expectedErr
		}

		s.mockRepo.MockLockWalletRecordForUpdate = func(username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{ID: walletId, Username: testUsername}, nil
		}

		// mock start of db transaction
		s.mock.ExpectBegin()

		// Validations
		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, testAmount)
		s.Require().Error(err, "internal same-user wallet transfers should error when gorm transaction fails")
		s.Require().Contains(err.Error(), expectedErr.Error())
		s.Require().Nil(txn, "failed transfer should not return the transaction")

	})

	s.Run("fails when the db transaction final commit fails", func() {
		var (
			testFromWalletId = "wallet-id"
			testToWalletId   = "to-wallet-id"
			testAmount       = uint64(10)
			testUsername     = "testusername"
		)

		s.mockRepo.MockWalletsFromUser = func(username string, walletIds []string) (bool, []models.Wallet, error) {
			s.Require().Equal(testFromWalletId, walletIds[0])
			s.Require().Equal(testToWalletId, walletIds[1])
			wallets := []models.Wallet{{
				ID:       testFromWalletId,
				Username: testUsername,
			}, {
				ID:       testToWalletId,
				Username: testUsername,
			}}
			return true, wallets, nil
		}

		s.mockRepo.MockDecrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testFromWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		s.mockRepo.MockIncrementWalletBalance = func(username, walletId string, amount uint64) error {
			s.Require().Equal(testToWalletId, walletId)
			s.Require().Equal(testAmount, amount)
			return nil
		}

		s.mockRepo.MockCreateTransaction = func(transactions *models.Transaction) error {
			s.Require().Equal(testAmount, transactions.Amount)
			s.Require().Equal(testFromWalletId, *transactions.FromID)
			s.Require().Equal(testUsername, *transactions.FromUsername)
			s.Require().Equal(testToWalletId, *transactions.ToID)
			s.Require().Equal(testUsername, *transactions.ToUsername)
			return nil
		}

		s.mockRepo.MockLockWalletRecordForUpdate = func(username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{ID: walletId, Username: testUsername}, nil
		}

		// mock start of db transaction
		s.mock.ExpectBegin()
		expectedErr := errors.New("DB transaction commit failed")
		s.mock.ExpectCommit().WillReturnError(expectedErr)

		// Validations
		txn, err := s.mgr.Transfer(testUsername, testFromWalletId, testToWalletId, testAmount)
		s.Require().Error(err, "internal same-user wallet transfers should error when gorm transaction fails")
		s.Require().Contains(err.Error(), expectedErr.Error())
		s.Require().Nil(txn, "failed transfer should not return the transaction")
	})

}

type MockRepo struct {
	models.Repository

	MockWalletsFromUser           func(username string, walletIds []string) (bool, []models.Wallet, error)
	MockDecrementWalletBalance    func(username, walletId string, amount uint64) error
	MockIncrementWalletBalance    func(username, walletId string, amount uint64) error
	MockCreateTransaction         func(transactions *models.Transaction) error
	MockLockWalletRecordForUpdate func(username, walletId string) (*models.Wallet, error)
}

func (m *MockRepo) WalletsFromUser(_ *gorm.DB, username string, walletIds []string) (bool, []models.Wallet, error) {
	return m.MockWalletsFromUser(username, walletIds)
}

func (m *MockRepo) DecrementWalletBalance(_ *gorm.DB, username, walletId string, amount uint64) error {
	return m.MockDecrementWalletBalance(username, walletId, amount)
}

func (m *MockRepo) IncrementWalletBalance(_ *gorm.DB, username, walletId string, amount uint64) error {
	return m.MockIncrementWalletBalance(username, walletId, amount)
}

func (m *MockRepo) CreateTransaction(_ *gorm.DB, transactions *models.Transaction) error {
	return m.MockCreateTransaction(transactions)
}

func (m *MockRepo) LockWalletRecordForUpdate(_ *gorm.DB, username, walletId string) (*models.Wallet, error) {
	return m.MockLockWalletRecordForUpdate(username, walletId)
}
