package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-test/deep"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type WalletRepositorySuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	repository Repository
}

func (s *WalletRepositorySuite) BeforeTest(_, _ string) {
	log.SetLevel(log.DebugLevel)
	var (
		sqlDB *sql.DB
		err   error
	)

	sqlDB, s.mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Require().NoError(err)
	s.DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{}) // open gorm db
	s.Require().NoError(err)
	s.repository = NewRepository()
}

func (s *WalletRepositorySuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func TestWalletRepository(t *testing.T) {
	suite.Run(t, new(WalletRepositorySuite))
}

func (s *WalletRepositorySuite) TestGetWallet() {
	var (
		expectedID       = "test-id"
		expectedUsername = "testusername"
	)
	s.mock.ExpectQuery("SELECT * FROM `wallets` WHERE username = ? AND id = ? LIMIT 1").
		WithArgs(expectedUsername, expectedID).
		WillReturnRows(sqlmock.NewRows(nil)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(expectedID, expectedUsername))

	res, err := s.repository.GetWallet(s.DB, expectedUsername, expectedID)
	s.Require().NoError(err)
	s.Require().Nil(deep.Equal(Wallet{ID: expectedID, Username: expectedUsername}, *res))
}

func (s *WalletRepositorySuite) TestDecrementWalletBalance() {
	// TODO: Investigate the update method.
	// Current behaviour: sql mock does not accept the gorm args passed
	// even though they are correct (it also appears that the error message
	// is misleading and is actually unrelated to mismatched params).
	// Expected behaviour: no error occurs.
	s.T().Skip("Skipping decrement wallet balance test")
}

func (s *WalletRepositorySuite) TestIncrementWalletBalance() {
	var (
		walletId = "wallet-id"
		amt      = uint64(10)
		username = "testusername"
	)
	s.mock.ExpectBegin()
	s.mock.ExpectExec("UPDATE `wallets` SET `balance`=balance + ? WHERE username = ? AND id = ?").
		WithArgs(amt, username, walletId).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err := s.repository.IncrementWalletBalance(s.DB, username, walletId, amt)
	s.Require().NoError(err)
}

func (s *WalletRepositorySuite) TestDeleteWalletSuccessful() {
	var (
		id       = "test-id"
		username = "testusername"
	)
	s.mock.ExpectBegin()
	s.mock.ExpectExec("DELETE FROM `wallets` WHERE username = ? AND id = ? AND balance = 0").
		WithArgs(username, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err := s.repository.DeleteWalletZeroBalance(s.DB, username, id)
	s.Require().NoError(err)
}
