package db

import (
	"fmt"
	"log"
	"os"

	"github.com/bazilio91/sferra-cloud/pkg/proto"

	"github.com/bazilio91/sferra-cloud/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Store struct {
	db *gorm.DB
}

func GetDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)
}

func InitDB(cfg *config.Config) error {
	dsn := GetDSN(cfg)
	return InitDBWithDSN(dsn)
}

func InitDBWithDSN(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	err = DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	if err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Enable logging in debug mode
	if os.Getenv("DEBUG") != "" {
		DB = DB.Debug()
		log.Println("Database debug mode enabled")
	}

	// Migrate the schema
	if err := migrateDB(DB); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func migrateDB(db *gorm.DB) error {
	// Create tables in order of dependencies
	models := []interface{}{
		&proto.ClientUserORM{},
		&proto.ClientORM{},
		&proto.DataRecognitionTaskORM{},
		&proto.Admin{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
