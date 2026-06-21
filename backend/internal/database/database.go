package database

import (
	"fmt"
	"log"

	"config-center/internal/config"
	"config-center/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.Tenant{},
		&models.Namespace{},
		&models.Group{},
		&models.ConfigItem{},
		&models.ConfigVersion{},
		&models.GrayRelease{},
		&models.ClientConnection{},
		&models.Metric{},
		&models.User{},
		&models.UserRole{},
		&models.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed")
	return nil
}
