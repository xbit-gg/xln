package xlntests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestEndToEndUsingPostgres(t *testing.T) {
	if os.Getenv(cfg.EnvVarNameTest) != cfg.EnvValTestPostgresE2E {
		t.Skip("Skipping integration tests")
	}
	postgresSuite := new(tlsClientSuite)
	postgresSuite.config = cfg.DefaultConfig()
	postgresSuite.config.Serving.Tls.EnableTls = true
	postgresSuite.config.DatabaseConnectionString = "postgresql://postgres:hodl@localhost:5432/xln_test"
	postgresSuite.config.LogLevel = log.DebugLevel

	_, err := SetupPostgres(postgresSuite.config.DatabaseConnectionString)
	require.Nil(t, err, "failed to setup postgres")

	suite.Run(t, postgresSuite)
}

type tlsClientSuite struct {
	suite.Suite
	client   xlnrpc.XlnAdminClient
	grpcConn *grpc.ClientConn
	config   *cfg.Config
}

func (s *tlsClientSuite) SetupSuite() {
	StartServer(s.config)
	var err error
	s.client, s.grpcConn, err = CreateHealthyAdminClient(s.config)
	s.Require().Nil(err)
}

func (s *tlsClientSuite) TearDownSuite() {
	s.grpcConn.Close()
}

func (s *tlsClientSuite) TestGRPCListUsers() {
	header := metadata.New(map[string]string{auth.AdminApiKeyHeader: s.config.XLNApiKey})
	ctx := metadata.NewOutgoingContext(context.Background(), header)
	res, err := s.client.ListUsers(ctx, &xlnrpc.ListUsersRequest{})
	s.Require().Nil(err, "should not error when listing users")
	s.Require().NotNil(res, "response should not be empty")
}

func (s *tlsClientSuite) TestHTTPSListUsers() {
	// disable checking if certificate is valid
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://localhost:%v/admin/users", s.config.Serving.RestPort), nil)
	s.Require().Nilf(err, "should not error when making request. Instead got: %v", err)
	req.Header = http.Header{
		auth.AdminApiKeyHeader: []string{s.config.XLNApiKey},
	}
	res, err := client.Do(req)
	s.Require().Nilf(err, "should not error when making request. Instead got: %v", err)
	s.Require().NotNil(res, "response should not be empty")
}
