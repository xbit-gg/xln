package withdraw

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	golnurl "github.com/fiatjaf/go-lnurl"
	"github.com/stretchr/testify/suite"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestWithdrawManager(t *testing.T) {
	suite.Run(t, new(withdrawManagerSuite))
}

type withdrawManagerSuite struct {
	suite.Suite
	mgr      Manager
	mockRepo mockRepo
	mock     sqlmock.Sqlmock
}

func (s *withdrawManagerSuite) SetupSuite() {
	var (
		err   error
		sqlDB *sql.DB
	)
	sqlDB, s.mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Require().NoError(err)
	sDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	s.mockRepo = mockRepo{}
	s.mgr = NewManager(
		cfg.DefaultConfig().Serving.Hostname,
		&db.DB{DB: sDB, Repo: &s.mockRepo},
	)
}

func (s *withdrawManagerSuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.mockRepo = mockRepo{}
}

func (s *withdrawManagerSuite) TestCreateLNURLWSucceedsWhenRepoSucceeds() {
	var (
		actualLNURL string
		actualErr   error
	)
	testInputs := []*models.Withdraw{{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}, {
		WalletID: "test-walletid-1",
		Username: "test-username-1",
		K1:       "test-k1-1",
	}}

	for _, expectedWithdraw := range testInputs {
		s.mockRepo.mockGetWallet = func(tx *gorm.DB, username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{
				ID:       expectedWithdraw.WalletID,
				Username: expectedWithdraw.Username,
				Locked:   false,
			}, nil
		}

		s.mockRepo.mockCreateWithdraw = func(tx *gorm.DB, actualWithdraw *models.Withdraw) (*models.Withdraw, error) {
			s.Require().Equal(expectedWithdraw.CreatedAt, actualWithdraw.CreatedAt)
			s.Require().Equal(expectedWithdraw.Description, actualWithdraw.Description)
			s.Require().Equal(expectedWithdraw.Expiry, actualWithdraw.Expiry)
			s.Require().Empty(actualWithdraw.K1)
			s.Require().Equal(expectedWithdraw.MaxMsat, actualWithdraw.MaxMsat)
			s.Require().Equal(expectedWithdraw.MaxUse, actualWithdraw.MaxUse)
			s.Require().Equal(expectedWithdraw.MinMsat, actualWithdraw.MinMsat)
			s.Require().Equal(expectedWithdraw.UpdatedAt, actualWithdraw.UpdatedAt)
			s.Require().Equal(expectedWithdraw.Username, actualWithdraw.Username)
			s.Require().Equal(expectedWithdraw.Uses, actualWithdraw.Uses)
			s.Require().Equal(expectedWithdraw.WalletID, actualWithdraw.WalletID)
			return expectedWithdraw, nil
		}

		actualLNURL, actualErr = s.mgr.CreateLNURLW(
			expectedWithdraw.Username,
			expectedWithdraw.WalletID,
			expectedWithdraw.Description,
			expectedWithdraw.MinMsat,
			expectedWithdraw.MaxMsat,
			expectedWithdraw.MaxUse,
			expectedWithdraw.Expiry,
		)
		s.Require().Nil(actualErr)
		decodedLNURL, decodeErr := golnurl.LNURLDecode(actualLNURL)
		s.Require().Nil(decodeErr, "should not error when decoding lnurl")
		s.Require().Equal(
			fmt.Sprintf("https://localhost:5551/lnurl/withdraw/request?k1=%s", expectedWithdraw.K1),
			decodedLNURL)
	}
}

func (s *withdrawManagerSuite) TestCreateLNURLWErrorsWhenWalletLocked() {
	var (
		actualLNURL string
		actualErr   error
	)

	testInput := models.Withdraw{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}

	s.mockRepo.mockGetWallet = func(tx *gorm.DB, username, walletId string) (*models.Wallet, error) {
		return &models.Wallet{
			ID:       testInput.WalletID,
			Username: testInput.Username,
			Locked:   true,
		}, nil
	}

	actualLNURL, actualErr = s.mgr.CreateLNURLW(
		testInput.Username,
		testInput.WalletID,
		testInput.Description,
		testInput.MinMsat,
		testInput.MaxMsat,
		testInput.MaxUse,
		testInput.Expiry,
	)

	s.Require().Equal(models.ErrLockedWallet, actualErr)
	s.Require().Empty(actualLNURL)
}

func (s *withdrawManagerSuite) TestCreateLNURLWErrorsWhenRepoErrors() {
	var (
		actualLNURL string
		actualErr   error
	)
	testInputs := []*models.Withdraw{{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}, {
		WalletID: "test-walletid-1",
		Username: "test-username-1",
		K1:       "test-k1-1",
	}}

	expectedErr := errors.New("test-error")

	for _, expectedWithdraw := range testInputs {
		s.mockRepo.mockGetWallet = func(tx *gorm.DB, username, walletId string) (*models.Wallet, error) {
			return &models.Wallet{
				ID:       expectedWithdraw.WalletID,
				Username: expectedWithdraw.Username,
				Locked:   false,
			}, nil
		}
		s.mockRepo.mockCreateWithdraw = func(tx *gorm.DB, actualWithdraw *models.Withdraw) (*models.Withdraw, error) {
			s.Require().Equal(expectedWithdraw.CreatedAt, actualWithdraw.CreatedAt)
			s.Require().Equal(expectedWithdraw.Description, actualWithdraw.Description)
			s.Require().Equal(expectedWithdraw.Expiry, actualWithdraw.Expiry)
			s.Require().Empty(actualWithdraw.K1)
			s.Require().Equal(expectedWithdraw.MaxMsat, actualWithdraw.MaxMsat)
			s.Require().Equal(expectedWithdraw.MaxUse, actualWithdraw.MaxUse)
			s.Require().Equal(expectedWithdraw.MinMsat, actualWithdraw.MinMsat)
			s.Require().Equal(expectedWithdraw.UpdatedAt, actualWithdraw.UpdatedAt)
			s.Require().Equal(expectedWithdraw.Username, actualWithdraw.Username)
			s.Require().Equal(expectedWithdraw.Uses, actualWithdraw.Uses)
			s.Require().Equal(expectedWithdraw.WalletID, actualWithdraw.WalletID)
			return expectedWithdraw, expectedErr
		}

		actualLNURL, actualErr = s.mgr.CreateLNURLW(
			expectedWithdraw.Username,
			expectedWithdraw.WalletID,
			expectedWithdraw.Description,
			expectedWithdraw.MinMsat,
			expectedWithdraw.MaxMsat,
			expectedWithdraw.MaxUse,
			expectedWithdraw.Expiry,
		)
		s.Require().Equal(expectedErr, actualErr)
		s.Require().Empty(actualLNURL)
	}
}

func (s *withdrawManagerSuite) TestGetLNURLWSucceedsWhenRepoSucceeds() {
	var (
		actualLNURL string
		actualErr   error
	)

	expectedWithdraw := &models.Withdraw{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}

	s.mockRepo.mockGetWalletWithdraw = func(tx *gorm.DB, actualUsername, actualWalletId, actualK1 string, actualIncludeExpired bool) (*models.Withdraw, error) {
		s.Require().False(actualIncludeExpired)
		s.Require().Equal(actualUsername, expectedWithdraw.Username)
		s.Require().Equal(actualWalletId, expectedWithdraw.WalletID)
		s.Require().Equal(actualK1, expectedWithdraw.K1)
		return expectedWithdraw, nil
	}

	actualLNURL, actualErr = s.mgr.GetLNURLW(
		expectedWithdraw.Username,
		expectedWithdraw.WalletID,
		expectedWithdraw.K1,
	)
	s.Require().Nil(actualErr)
	decodedLNURL, decodeErr := golnurl.LNURLDecode(actualLNURL)
	s.Require().Nil(decodeErr, "should not error when decoding lnurl")
	s.Require().Equal(
		fmt.Sprintf("https://localhost:5551/lnurl/withdraw/request?k1=%s", expectedWithdraw.K1),
		decodedLNURL)
}

func (s *withdrawManagerSuite) TestGetLNURLWErrorsWhenRepoErrors() {
	var (
		actualLNURL string
		actualErr   error
	)

	expectedUsername := "test-username"
	expectedWalletId := "test-walletId"
	expectedK1 := "test-k1"
	expectedErr := errors.New("test-error")

	s.mockRepo.mockGetWalletWithdraw = func(tx *gorm.DB, actualUsername, actualWalletId, actualK1 string, actualIncludeExpired bool) (*models.Withdraw, error) {
		s.Require().False(actualIncludeExpired)
		s.Require().Equal(actualUsername, expectedUsername)
		s.Require().Equal(actualWalletId, expectedWalletId)
		s.Require().Equal(actualK1, expectedK1)
		return &models.Withdraw{}, expectedErr
	}

	actualLNURL, actualErr = s.mgr.GetLNURLW(expectedUsername, expectedWalletId, expectedK1)
	s.Require().Equal(expectedErr, actualErr)
	s.Require().Empty(actualLNURL)
}

func (s *withdrawManagerSuite) TestGetWithdrawRequestSucceedsPermissionlesslyWhenRepoSucceeds() {
	expectedK1 := "test-k1"
	expectedWithdraw := &models.Withdraw{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}

	s.mockRepo.mockGetWithdraw = func(tx *gorm.DB, actualK1 string, actualIncludeExpired bool) (*models.Withdraw, error) {
		s.Require().False(actualIncludeExpired)
		s.Require().Equal(actualK1, expectedK1)
		return expectedWithdraw, nil
	}

	actualWithdraw, actualCallback, actualErr := s.mgr.GetWithdrawRequest(expectedK1)
	s.Require().Nil(actualErr)
	s.Require().Equal("https://localhost:5551/lnurl/withdraw/pay", actualCallback)
	s.Require().Equal(expectedWithdraw.CreatedAt, actualWithdraw.CreatedAt)
	s.Require().Equal(expectedWithdraw.Description, actualWithdraw.Description)
	s.Require().Equal(expectedWithdraw.Expiry, actualWithdraw.Expiry)
	s.Require().Equal(expectedWithdraw.K1, actualWithdraw.K1)
	s.Require().Equal(expectedWithdraw.MaxMsat, actualWithdraw.MaxMsat)
	s.Require().Equal(expectedWithdraw.MaxUse, actualWithdraw.MaxUse)
	s.Require().Equal(expectedWithdraw.MinMsat, actualWithdraw.MinMsat)
	s.Require().Equal(expectedWithdraw.UpdatedAt, actualWithdraw.UpdatedAt)
	s.Require().Equal(expectedWithdraw.Username, actualWithdraw.Username)
	s.Require().Equal(expectedWithdraw.Uses, actualWithdraw.Uses)
	s.Require().Equal(expectedWithdraw.WalletID, actualWithdraw.WalletID)
}

func (s *withdrawManagerSuite) TestGetWithdrawRequestErrorsWhenRepoErrors() {
	expectedK1 := "test-k1"
	expectedErr := errors.New("test-error")
	expectedWithdraw := &models.Withdraw{
		WalletID:    "test-walletid",
		Username:    "test-username",
		Description: "test-description",
		MinMsat:     1,
		MaxMsat:     2,
		MaxUse:      10000,
		Expiry:      time.Now(),
		K1:          "test-k1",
	}

	s.mockRepo.mockGetWithdraw = func(tx *gorm.DB, actualK1 string, actualIncludeExpired bool) (*models.Withdraw, error) {
		s.Require().False(actualIncludeExpired)
		s.Require().Equal(actualK1, expectedK1)
		return expectedWithdraw, expectedErr
	}

	actualWithdraw, actualCallback, actualErr := s.mgr.GetWithdrawRequest(expectedK1)
	s.Require().Equal(expectedErr, actualErr)
	s.Require().Nil(actualWithdraw)
	s.Require().Empty(actualCallback)
}

type mockRepo struct {
	models.Repository

	mockCreateWithdraw    func(tx *gorm.DB, withdraw *models.Withdraw) (*models.Withdraw, error)
	mockGetWalletWithdraw func(tx *gorm.DB, username, walletId, k1 string, includeExpired bool) (*models.Withdraw, error)
	mockGetWithdraw       func(tx *gorm.DB, k1 string, includeExpired bool) (*models.Withdraw, error)
	mockGetWallet         func(tx *gorm.DB, username, walletId string) (*models.Wallet, error)
}

func (m *mockRepo) CreateWithdraw(tx *gorm.DB, withdraw *models.Withdraw) (*models.Withdraw, error) {
	return m.mockCreateWithdraw(tx, withdraw)
}

func (m *mockRepo) GetWalletWithdraw(tx *gorm.DB, username, walletId, k1 string, includeExpired bool) (*models.Withdraw, error) {
	return m.mockGetWalletWithdraw(tx, username, walletId, k1, includeExpired)
}

func (m *mockRepo) GetWithdraw(tx *gorm.DB, k1 string, includeExpired bool) (*models.Withdraw, error) {
	return m.mockGetWithdraw(tx, k1, includeExpired)
}

func (m *mockRepo) GetWallet(tx *gorm.DB, username, walletId string) (*models.Wallet, error) {
	return m.mockGetWallet(tx, username, walletId)
}
