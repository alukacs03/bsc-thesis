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
	err = db.AutoMigrate(&models.User{}, &models.UserRegistrationRequest{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
