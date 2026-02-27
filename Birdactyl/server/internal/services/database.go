package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func generatePassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)
	if len(name) > 16 {
		name = name[:16]
	}
	return name
}

func connectToHost(host *models.DatabaseHost) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", host.Username, host.Password, host.Host, host.Port)
	return sql.Open("mysql", dsn)
}

func CreateServerDatabase(serverID uuid.UUID, hostID uuid.UUID, name string) (*models.ServerDatabase, error) {
	var host models.DatabaseHost
	if err := database.DB.First(&host, "id = ?", hostID).Error; err != nil {
		return nil, fmt.Errorf("database host not found")
	}

	if host.MaxDatabases > 0 {
		var count int64
		database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", hostID).Count(&count)
		if count >= int64(host.MaxDatabases) {
			return nil, fmt.Errorf("database host has reached maximum capacity")
		}
	}

	shortID := serverID.String()[:8]
	dbName := fmt.Sprintf("s%s_%s", shortID, sanitizeName(name))
	username := fmt.Sprintf("u%s_%s", shortID, sanitizeName(name))
	if len(username) > 32 {
		username = username[:32]
	}
	password := generatePassword(24)

	conn, err := connectToHost(&host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database host: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)); err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	if _, err := conn.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", username, password)); err != nil {
		conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	if _, err := conn.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", dbName, username)); err != nil {
		conn.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%'", username))
		conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		return nil, fmt.Errorf("failed to grant privileges: %v", err)
	}

	conn.Exec("FLUSH PRIVILEGES")

	serverDB := &models.ServerDatabase{
		ServerID:     serverID,
		HostID:       hostID,
		DatabaseName: dbName,
		Username:     username,
		Password:     password,
	}

	if err := database.DB.Create(serverDB).Error; err != nil {
		conn.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%'", username))
		conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		return nil, fmt.Errorf("failed to save database record: %v", err)
	}

	return serverDB, nil
}

func DeleteServerDatabase(dbID uuid.UUID) error {
	var serverDB models.ServerDatabase
	if err := database.DB.Preload("Host").First(&serverDB, "id = ?", dbID).Error; err != nil {
		return fmt.Errorf("database not found")
	}

	if serverDB.Host != nil {
		conn, err := connectToHost(serverDB.Host)
		if err == nil {
			defer conn.Close()
			conn.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%'", serverDB.Username))
			conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", serverDB.DatabaseName))
			conn.Exec("FLUSH PRIVILEGES")
		}
	}

	return database.DB.Delete(&serverDB).Error
}

func RotateDatabasePassword(dbID uuid.UUID) (*models.ServerDatabase, error) {
	var serverDB models.ServerDatabase
	if err := database.DB.Preload("Host").First(&serverDB, "id = ?", dbID).Error; err != nil {
		return nil, fmt.Errorf("database not found")
	}

	newPassword := generatePassword(24)

	conn, err := connectToHost(serverDB.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database host: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Exec(fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s'", serverDB.Username, newPassword)); err != nil {
		return nil, fmt.Errorf("failed to update password: %v", err)
	}

	conn.Exec("FLUSH PRIVILEGES")

	serverDB.Password = newPassword
	if err := database.DB.Save(&serverDB).Error; err != nil {
		return nil, fmt.Errorf("failed to save new password: %v", err)
	}

	return &serverDB, nil
}

func GetServerDatabases(serverID uuid.UUID) ([]models.ServerDatabase, error) {
	var databases []models.ServerDatabase
	err := database.DB.Preload("Host").Where("server_id = ?", serverID).Find(&databases).Error
	return databases, err
}

func DeleteAllServerDatabases(serverID uuid.UUID) error {
	var databases []models.ServerDatabase
	if err := database.DB.Preload("Host").Where("server_id = ?", serverID).Find(&databases).Error; err != nil {
		return err
	}

	for _, db := range databases {
		DeleteServerDatabase(db.ID)
	}

	return nil
}

func CreateDatabaseHost(name, host string, port int, username, password string, maxDatabases int) (*models.DatabaseHost, error) {
	dbHost := &models.DatabaseHost{
		Name:         name,
		Host:         host,
		Port:         port,
		Username:     username,
		Password:     password,
		MaxDatabases: maxDatabases,
	}

	if err := database.DB.Create(dbHost).Error; err != nil {
		return nil, err
	}

	return dbHost, nil
}
