package user

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type userManagerSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	mgr  Manager
	repo *MockRepo
}

func (s *userManagerSuite) SetupSuite() {
	var (
		sqlDB *sql.DB
		err   error
	)

	sqlDB, s.mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Require().NoError(err)
	mockDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{}) // open gorm db
	s.repo = &MockRepo{}
	s.mgr = NewManager(&db.DB{DB: mockDB, Repo: s.repo})
	s.Require().NoError(err)
}

func (s *userManagerSuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.repo = &MockRepo{}
}

func TestUserManager(t *testing.T) {
	suite.Run(t, new(userManagerSuite))
}

func (s *userManagerSuite) TestCreateUser() {
	s.Run("returns correctly", func() {
		username := "testusername"

		s.repo.MockCreateUser = func(user *models.User) error {
			s.Require().Equal(user.Username, username)
			user.Username = username
			return nil
		}
		s.repo.MockCreateWallet = func(wallet *models.Wallet) error {
			s.Require().Equal(wallet.ID, username)
			s.Require().Equal(wallet.Username, username)
			return nil
		}

		// mock start of db transaction
		s.mock.ExpectBegin()
		s.mock.ExpectCommit().WillReturnError(nil)

		user, err := s.mgr.CreateUser(username)
		s.Require().NoError(err, "valid and successful create user call should not error")
		s.Require().Equal(user.Username, username, "user should be correctly returned from create user call")
		s.Require().Equal(user.Wallets[0].ID, username, "user should be correctly returned from create user call")
	})

	s.Run("errors when default wallet creation fails", func() {
		username := "testusername"

		s.repo.MockCreateUser = func(user *models.User) error {
			s.Require().Equal(user.Username, username)
			user.Username = username
			return nil
		}
		expectedErr := errors.New("create wallet db call failed")
		s.repo.MockCreateWallet = func(wallet *models.Wallet) error {
			s.Require().Equal(wallet.ID, username)
			s.Require().Equal(wallet.Username, username)
			return expectedErr
		}

		// mock start of db transaction
		s.mock.ExpectBegin()

		user, err := s.mgr.CreateUser(username)
		s.Require().Error(err, "valid but unsuccessful create user call should have errored")
		s.Require().Equal(err.Error(), expectedErr.Error())
		s.Require().Nil(user, "user should be nil when create user errors")
	})
}

// Mocks
type MockRepo struct {
	models.Repository

	MockCreateUser   func(user *models.User) error
	MockCreateWallet func(wallet *models.Wallet) error
}

func (m *MockRepo) CreateUser(_ *gorm.DB, user *models.User) error {
	return m.MockCreateUser(user)
}

func (m *MockRepo) CreateWallet(_ *gorm.DB, wallet *models.Wallet) error {
	return m.MockCreateWallet(wallet)
}
