package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Driver string

func ILike(column, value string) string {
	if Driver == "postgres" {
		return fmt.Sprintf("%s ILIKE ?", column)
	}
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column)
}

func ILikeValue(value string) string {
	return "%" + strings.ReplaceAll(strings.ReplaceAll(value, "%", "\\%"), "_", "\\_") + "%"
}

func Connect(cfg *config.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector
	Driver = cfg.Driver

	switch cfg.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Name)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	if cfg.Driver != "sqlite" {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.IPRegistration{},
		&models.Node{},
		&models.Package{},
		&models.Server{},
		&models.ActivityLog{},
		&models.IPBan{},
		&models.Setting{},
		&models.Subuser{},
		&models.DatabaseHost{},
		&models.ServerDatabase{},
		&models.Schedule{},
		&models.APIKey{},
	); err != nil {
		return err
	}

	log.Printf("Database connected successfully (driver: %s)", cfg.Driver)
	return nil
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
