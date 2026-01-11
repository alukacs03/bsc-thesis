package database

import (
	"gluon-api/models"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	dbPath := resolveDBPath()
	log.Printf("Using database path: %s", dbPath)
	gormLog := gormlogger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             2 * time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}
	DB = db

	
	
	_ = db.Exec("PRAGMA journal_mode = WAL;").Error
	_ = db.Exec("PRAGMA synchronous = NORMAL;").Error
	_ = db.Exec("PRAGMA busy_timeout = 5000;").Error

	if sqlDB, err := db.DB(); err == nil {
		
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(0)
	}

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
		&models.NodeSSHAuthorizedKey{},
		&models.NodeCommand{},

		&models.APIKey{},

		&models.WireGuardProfile{},
		&models.OSPFProfile{},

		&models.KubernetesCluster{},
		&models.DeploymentSettings{},

		&models.AuditLog{},
		&models.Event{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func resolveDBPath() string {
	if v := strings.TrimSpace(os.Getenv("GLUON_DB_PATH")); v != "" {
		return v
	}
	if err := os.MkdirAll("/var/lib/gluon", 0755); err == nil {
		return "/var/lib/gluon/gluon.db"
	}
	return "gluon.db"
}
