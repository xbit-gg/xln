package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TransactionRepositorySuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	repository Repository
}

func (s *TransactionRepositorySuite) BeforeTest(_, _ string) {
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

func (s *TransactionRepositorySuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func TestTransactionRepository(t *testing.T) {
	suite.Run(t, new(TransactionRepositorySuite))
}

func (s *WalletRepositorySuite) TestCreateTransactions() {
	var (
		walletId   = "wallet-id"
		toWalletId = "to-wallet-id"
		amt        = uint64(10)
		username   = "testusername"
		toUsername = "to-testusername"
	)
	// add to transaction list
	s.mock.ExpectBegin()
	s.mock.ExpectExec("INSERT INTO `transactions` "+
		"(`id`,`created_at`,`updated_at`,`deleted_at`,`from_id`,`from_username`,"+
		"`to_id`,`to_username`,`label`,`amount`,`fees_paid`,`invoice_id`) "+
		"VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, walletId, username, toWalletId, toUsername, nil, amt, 0, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	transactions := []Transaction{{
		FromID:       &walletId,
		FromUsername: &username,
		ToID:         &toWalletId,
		ToUsername:   &toUsername,
		Amount:       amt,
		FeesPaid:     0,
	}}
	err := s.repository.CreateTransactions(s.DB, &transactions)
	s.Require().NoError(err)
}
