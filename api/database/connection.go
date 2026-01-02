package database

import (
	"gluon-api/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("gluon.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db

	err = db.AutoMigrate(
		&models.User{},

		&models.UserRegistrationRequest{},

		&models.IPPool{},
		&models.IPAllocation{},
		&models.LinkAllocation{},

		&models.NodeEnrollmentRequest{},
		&models.Node{},
		&models.WireGuardInterface{},
		&models.NodePeer{},

		&models.NodeConfig{},

		&models.APIKey{},

		&models.WireGuardProfile{},
		&models.OSPFProfile{},

		&models.AuditLog{},
		&models.Event{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
