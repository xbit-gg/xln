package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserRespositorySuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	repository Repository
}

func (s *UserRespositorySuite) BeforeTest(_, _ string) {
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

	s.repository = NewRepository()
}

func (s *UserRespositorySuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRespositorySuite))
}

func (s *UserRespositorySuite) TestCreateUser() {
	s.Run("creates new user", func() {
		username := "testusername"
		s.mock.ExpectBegin()
		s.mock.ExpectExec("INSERT INTO `users` (`username`,`created_at`,`updated_at`,`deleted_at`,`api_key`,`link_key`,`link_label`) VALUES (?,?,?,?,?,?,?)").
			WithArgs(username, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=", nil, nil).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		user := &User{Username: username}
		err := s.repository.CreateUser(s.DB, user)
		s.Require().NoError(err)
	})
}
