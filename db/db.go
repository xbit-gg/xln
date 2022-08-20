package db

import (
	"errors"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	Repo models.Repository
}

func ConnectDB(connectionString string) (*DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	opts := &gorm.Config{
		PrepareStmt: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	if strings.HasPrefix(connectionString, "postgres") {
		log.Info("Using Postgres database")
		db, err = gorm.Open(postgres.Open(connectionString), opts)
	} else {
		log.Info("Using SQLite database")
		if _, err := os.Stat(connectionString); errors.Is(err, os.ErrNotExist) {
			if myfile, err := os.Create(connectionString); err != nil {
				log.WithError(err).WithField("DatabaseConnectionString", connectionString).Fatal("Failed to create database file")
			} else {
				log.WithField("DatabaseConnectionString", connectionString).Info("Created database file")
				myfile.Close()
			}
		} else {
			log.WithField("DatabaseConnectionString", connectionString).Info("file exists")
		}
		db, err = gorm.Open(sqlite.Open(connectionString), opts)
	}
	if err != nil {
		log.WithError(err).Error("Failed to connect to DB")
		return nil, err
	}
	if err := migrate(db); err != nil {
		log.WithError(err).Error("Failed to migrate DB")
		return nil, err
	}

	log.Info("Connected to DB")

	return &DB{DB: db, Repo: models.NewRepository()}, err
}

func migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Withdraw{},
		&models.Transaction{},
		&models.PendingInvoice{},
		&models.PendingPayment{},
		&models.Auth{},
	)
	return err
}
