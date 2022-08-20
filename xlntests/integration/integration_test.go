package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/xlnrpc"
	"github.com/xbit-gg/xln/xlntests"
	"google.golang.org/grpc"
)

func TestSqliteIntegration(t *testing.T) {
	if os.Getenv(cfg.EnvVarNameTest) != cfg.EnvValTestSqliteIntegration {
		t.Skip("Skipping sqlite integration tests...")
	}
	sqliteSuite := new(integrationSuite)
	// set config
	sqliteSuite.config = cfg.DefaultConfig()
	sqliteSuite.config.Serving.Tls.EnableTls = false
	sqliteSuite.config.DatabaseConnectionString = fmt.Sprintf("%s/xln_test_integration.db", os.Getenv("HOME"))
	sqliteSuite.config.LogLevel = log.DebugLevel

	// clear previous test db to drop all tables
	sqlite, err := xlntests.SetupSqlite(sqliteSuite.config.DatabaseConnectionString)
	require.Nil(t, err, "connection to database should not error")
	sqliteSuite.db = sqlite

	suite.Run(t, sqliteSuite)
}

func TestPostgresIntegration(t *testing.T) {
	if os.Getenv(cfg.EnvVarNameTest) != cfg.EnvValTestPostgresIntegration {
		t.Skip("Skipping postgres integration tests...")
	}
	postgresSuite := new(integrationSuite)
	postgresSuite.config = cfg.DefaultConfig()
	postgresSuite.config.Serving.Tls.EnableTls = false
	postgresSuite.config.DatabaseConnectionString = "postgresql://postgres:hodl@localhost:5432/xln_test"
	postgresSuite.config.LogLevel = log.DebugLevel

	postgres, err := xlntests.SetupPostgres(postgresSuite.config.DatabaseConnectionString)
	require.Nil(t, err, "failed to setup postgres")
	postgresSuite.db = postgres

	suite.Run(t, postgresSuite)
}

type integrationSuite struct {
	suite.Suite
	client        xlnrpc.XlnClient
	adminClient   xlnrpc.XlnAdminClient
	grpcConn      *grpc.ClientConn
	adminGrpcConn *grpc.ClientConn
	config        *cfg.Config
	db            *db.DB
}

func (s *integrationSuite) SetupSuite() {
	// start server
	xlntests.StartServer(s.config)

	// set up test connection and client
	var err error
	s.client, s.grpcConn, err = xlntests.CreateHealthyClient(s.config)
	s.Require().Nil(err)
	s.adminClient, s.adminGrpcConn, err = xlntests.CreateHealthyAdminClient(s.config)
	s.Require().Nil(err)
}

func (s *integrationSuite) TearDownSuite() {
	s.grpcConn.Close()
	s.adminGrpcConn.Close()
}

func (s *integrationSuite) BeforeTest(_, _ string) {
	// need a WHERE clause in order to trigger global updates
	// https://gorm.io/docs/delete.html#Block-Global-Delete
	s.db.Unscoped().Where("1 = 1").Delete(&models.Auth{}, nil)
	s.db.Unscoped().Where("1 = 1").Delete(&models.Invoice{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.PendingInvoice{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.PendingPayment{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.Transaction{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.User{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.Wallet{})
	s.db.Unscoped().Where("1 = 1").Delete(&models.Withdraw{})
}

func (s *integrationSuite) createUser(ctx context.Context, username string) (*xlnrpc.CreateUserResponse, error) {
	res, err := s.adminClient.CreateUser(ctx, &xlnrpc.CreateUserRequest{Username: username})
	s.Require().Nil(err, "should not error when create users", err)
	s.Require().NotNil(res, "create user response should not be empty")
	return res, err
}

func (s *integrationSuite) createWallet(ctx context.Context, walletId string) (*xlnrpc.CreateWalletResponse, error) {
	res, err := s.client.CreateWallet(ctx, &xlnrpc.CreateWalletRequest{WalletId: walletId})
	s.Require().Nil(err, "should not error when creating wallet", err)
	s.Require().NotNil(res, "create wallet response should not be empty")
	s.Require().NotEmpty(res.ApiKey, "new wallet should have api key")
	s.Require().Equal(walletId, res.WalletId)
	s.Require().Equal(walletId, res.WalletName, "by default wallet name should be equal to wallet id")
	return res, err
}

func (s *integrationSuite) createWalletWithName(ctx context.Context, walletId, walletName string) (*xlnrpc.CreateWalletResponse, error) {
	res, err := s.client.CreateWallet(ctx, &xlnrpc.CreateWalletRequest{WalletId: walletId, WalletName: walletName})
	s.Require().Nil(err, "should not error when creating wallet", err)
	s.Require().NotNil(res, "create wallet response should not be empty")
	s.Require().NotEmpty(res.ApiKey, "new wallet should have api key")
	s.Require().Equal(walletId, res.WalletId)
	s.Require().Equal(walletName, res.WalletName)
	return res, err
}

func (s *integrationSuite) deleteWallet(ctx context.Context, walletId string) (*xlnrpc.DeleteWalletResponse, error) {
	return s.client.DeleteWallet(ctx, &xlnrpc.DeleteWalletRequest{WalletId: walletId})
}

func (s *integrationSuite) listWallets(ctx context.Context, data bool) (*xlnrpc.ListWalletsResponse, error) {
	return s.client.ListWallets(ctx, &xlnrpc.ListWalletsRequest{Data: data})
}

func (s *integrationSuite) getWallet(ctx context.Context, walletId string) (*xlnrpc.GetWalletResponse, error) {
	return s.client.GetWallet(ctx, &xlnrpc.GetWalletRequest{WalletId: walletId})
}

func (s *integrationSuite) softDeleteUser(ctx context.Context, username string) {
	res, err := s.adminClient.DeleteUser(ctx, &xlnrpc.DeleteUserRequest{Username: username})
	s.Require().Nil(err, "should not error when deleting existing user", err)
	s.Require().NotNil(res, "delete user response should not be empty")
	getWalletRes, err := s.client.GetWallet(ctx, &xlnrpc.GetWalletRequest{WalletId: username})
	s.Require().Nil(getWalletRes, "default wallet should be deleted")
	s.Require().NotNil(err, "should error because nonexistent wallet was requested")
	listWalRes, err := s.listWallets(ctx, false)
	s.Require().NotNil(err, "should error when listing wallets")
	s.Require().Nil(listWalRes, "list wallet results should not return on deleted user")
}
