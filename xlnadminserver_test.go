package xln

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/resources/user"
	"github.com/xbit-gg/xln/xlnrpc"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestGetInfoReturnsInfoCorrectly(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	// mock version to validate correct version is used
	originalVersion := x.xln.Version
	// restore original mocked function's behaviour
	defer func() { x.xln.Version = originalVersion }()
	expectedVersion := "test-version"
	x.xln.Version = expectedVersion
	userManager.MockListUsers = func() ([]*models.User, error) {
		return []*models.User{{Username: "TestUsername"}}, nil
	}

	request := xlnrpc.GetAdminInfoRequest{}
	actualInfo, err := x.GetInfo(ctx, &request)
	var exepectedUsers uint64 = 1
	if err != nil {
		t.Error("DB fetch did not return an error but the request method unexpectedly returned an error")
	}
	if exepectedUsers != actualInfo.Users {
		t.Error("Username returned from the DB does not match the username in the response")
	}
	if actualInfo.Version != expectedVersion {
		t.Error("Response returned an unexpected version")
	}
}

func TestGetInfoReturnsErrorWhenDBFails(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	expectedError := errors.New("test-error")
	userManager.MockListUsers = func() ([]*models.User, error) {
		return nil, expectedError
	}
	request := xlnrpc.GetAdminInfoRequest{}
	actualInfo, actualErr := x.GetInfo(ctx, &request)
	if actualErr == nil {
		t.Error("DB fetch errored which should have made the request method error")
	}
	if actualInfo != nil {
		t.Error("Response should be empty because DB request failed")
	}
	if !strings.Contains(actualErr.Error(), expectedError.Error()) {
		t.Error("Request error should contain DB error message")
	}
}

func TestGetInfoFailsWhenUnauthorized(t *testing.T) {
	authService, _, ctx, x := setupMocks()
	authService.MockValidateAdminCredentials = func() error { return errors.New("test-error") }
	request := xlnrpc.GetAdminInfoRequest{}
	info, err := x.GetInfo(ctx, &request)
	if err == nil {
		t.Error("Unauthorized request should have errored")
	}
	if info != nil {
		t.Error("Unauthorized request should not return data")
	}
}

func TestCreateUserDbWriteCorrect(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	expectedUsername := "testusername"
	var actualUsername string
	userManager.MockCreateUser = func(username string) (*models.User, error) {
		actualUsername = username
		return &models.User{Username: username}, nil
	}
	request := xlnrpc.CreateUserRequest{Username: expectedUsername}
	x.CreateUser(ctx, &request)

	t.Log(expectedUsername)
	t.Log(actualUsername)
	if actualUsername != expectedUsername {
		t.Error("The username in the request does not match the username passed to the db write manager")
	}
}

func TestCreateUserReturnsUserCorrectly(t *testing.T) {
	authService, userManager, ctx, x := setupMocks()
	expectedUsername := "testusername"
	expectedApiKey := "apikey"
	authService.MockValidateAdminCredentials = func() error {
		return nil
	}
	userManager.MockCreateUser = func(_ string) (*models.User, error) {
		return &models.User{Username: expectedUsername, ApiKey: expectedApiKey}, nil
	}
	request := xlnrpc.CreateUserRequest{Username: expectedUsername}
	actualUser, err := x.CreateUser(ctx, &request)
	if err != nil {
		t.Error("DB fetch did not return an error but the request method unexpectedly returned an error")
	}
	if expectedUsername != actualUser.Username {
		t.Error("Username returned from the DB manager does not match the username in the response")
	}
	if expectedApiKey != actualUser.ApiKey {
		t.Error("User apikey returned from the DB does not match the apikey in the response")
	}
}

func TestCreateUserErrorsWhenInvalidName(t *testing.T) {
	authService, _, ctx, x := setupMocks()
	authService.MockValidateAdminCredentials = func() error {
		return nil
	}

	request := xlnrpc.CreateUserRequest{Username: "invalid-username-*!"}
	user, actualErr := x.CreateUser(ctx, &request)
	if actualErr == nil {
		t.Error("Invalid request should have returned an error")
	}
	if user != nil {
		t.Error("Response should be empty because invalid input was provided")
	}
}

func TestCreateUserReturnsErrorWhenDBFails(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	expectedUsername := "testusername"
	expectedError := errors.New("test-error")
	userManager.MockCreateUser = func(_ string) (*models.User, error) {
		return nil, expectedError
	}
	request := xlnrpc.CreateUserRequest{Username: expectedUsername}
	user, actualErr := x.CreateUser(ctx, &request)
	if actualErr == nil {
		t.Error("DB fetch errored which should have made the request method error")
	}
	if user != nil {
		t.Error("Response should be empty because DB request failed")
	}
	if !strings.Contains(actualErr.Error(), expectedError.Error()) {
		t.Error("Request error should contain DB error message")
	}
}

func TestCreateUserFailsWhenUnauthorized(t *testing.T) {
	authService, _, ctx, x := setupMocks()
	authService.MockValidateAdminCredentials = func() error { return errors.New("test-error") }
	request := xlnrpc.CreateUserRequest{}
	user, err := x.CreateUser(ctx, &request)
	if err == nil {
		t.Error("Unauthorized request should have errored")
	}
	if user != nil {
		t.Error("Unauthorized request should not return data")
	}
}

func TestListUsersReturnsUsersCorrectly(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	expectedUsername := "testusername"
	userManager.MockListUsers = func() ([]*models.User, error) {
		return []*models.User{{Username: expectedUsername}}, nil
	}
	request := xlnrpc.ListUsersRequest{}
	actualUser, err := x.ListUsers(ctx, &request)
	if err != nil {
		t.Error("DB fetch did not return an error but the request method unexpectedly returned an error")
	}
	if expectedUsername != actualUser.Usernames[0] {
		t.Error("username returned from the DB does not match the username in the response")
	}
	if len(actualUser.Usernames) != 1 {
		t.Error("DB returned 1 user but ListUsers response returned more than 1 user")
	}
}

func TestListUsersReturnsErrorWhenDBFails(t *testing.T) {
	_, userManager, ctx, x := setupMocks()
	expectedError := errors.New("test-error")
	userManager.MockListUsers = func() ([]*models.User, error) {
		return nil, expectedError
	}
	request := xlnrpc.ListUsersRequest{}
	users, actualErr := x.ListUsers(ctx, &request)
	if actualErr == nil {
		t.Error("DB fetch errored which should have made the request method error")
	}
	if users != nil {
		t.Error("Response should be empty because DB request failed")
	}
	if !strings.Contains(actualErr.Error(), expectedError.Error()) {
		t.Error("Request error should contain DB error message")
	}
}

func TestListUsersFailsWhenUnauthorized(t *testing.T) {
	authService, _, ctx, x := setupMocks()
	authService.MockValidateAdminCredentials = func() error { return errors.New("test-error") }
	request := xlnrpc.ListUsersRequest{}
	users, err := x.ListUsers(ctx, &request)
	if err == nil {
		t.Error("Unauthorized request should have errored")
	}
	if users != nil {
		t.Error("Unauthorized request should not return data")
	}
}

// Mock DB
type mockUserManager struct {
	user.Manager

	MockCreateUser func(username string) (*models.User, error)
	MockGetUser    func(username string) (*models.User, error)
	MockListUsers  func() ([]*models.User, error)
}

func (m *mockUserManager) CreateUser(username string) (*models.User, error) {
	return m.MockCreateUser(username)
}

func (m *mockUserManager) GetUser(username string) (*models.User, error) {
	return m.MockGetUser(username)
}

func (m *mockUserManager) ListUsers() ([]*models.User, error) {
	return m.MockListUsers()
}

// Mock credential authentication
type mockAuthService struct {
	auth.Service

	MockValidateAdminCredentials func() error
}

func (m *mockAuthService) ValidateAdminCredentials(_ context.Context, _ string) error {
	return m.MockValidateAdminCredentials()
}

func (m *mockAuthService) ValidateUserCredentials(_ context.Context, _ string) (string, error) {
	return "testusername", nil
}

func (m *mockAuthService) ValidateWalletCredentials(_ context.Context, _ string, _ string) (string, error) {
	return "testusername", nil
}

func setupMocks() (*mockAuthService, *mockUserManager, context.Context, xlnAdminServer) {
	authService := &mockAuthService{}
	userManager := &mockUserManager{}
	authService.MockValidateAdminCredentials = func() error { return nil }
	xln := XLN{Version: version, AuthService: authService, Users: userManager}
	return authService, userManager, context.Background(), xlnAdminServer{xln: &xln}
}
