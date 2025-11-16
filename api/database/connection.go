package database

import (
	"gluon-api/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("gluon.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db

	// Run migrations for all models
	// Order matters: parent tables before child tables (for foreign keys)
	err = db.AutoMigrate(
		// Core models
		&models.User{},

		// User enrollment
		&models.UserRegistrationRequest{},

		// IPAM
		&models.IPPool{},
		&models.IPAllocation{},

		// Node enrollment and management
		&models.NodeEnrollmentRequest{},
		&models.Node{},
		&models.NodePeer{},

		// Authentication
		&models.APIKey{},

		// Configuration profiles
		&models.WireGuardProfile{},
		&models.OSPFProfile{},

		// Logging
		&models.AuditLog{},
		&models.Event{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
