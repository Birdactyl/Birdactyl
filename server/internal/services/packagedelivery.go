package services

import (
	"errors"
	"fmt"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
)

var (
	ErrPackageNotFound  = errors.New("package not found")
	ErrPackageNameTaken = errors.New("package name already exists")
)

func CreatePackage(pkg *models.Package) error {
	var existing models.Package
	if err := database.DB.Where("name = ? AND version = ?", pkg.Name, pkg.Version).First(&existing).Error; err == nil {
		return ErrPackageNameTaken
	}
	return database.DB.Create(pkg).Error
}

func GetPackages() ([]models.Package, error) {
	var packages []models.Package
	err := database.DB.Order("name ASC, version DESC").Find(&packages).Error
	return packages, err
}

func GetPackageByID(id uuid.UUID) (*models.Package, error) {
	var pkg models.Package
	if err := database.DB.Where("id = ?", id).First(&pkg).Error; err != nil {
		return nil, ErrPackageNotFound
	}
	return &pkg, nil
}

func UpdatePackage(id uuid.UUID, updates map[string]interface{}) (*models.Package, error) {
	var pkg models.Package
	if err := database.DB.Where("id = ?", id).First(&pkg).Error; err != nil {
		return nil, ErrPackageNotFound
	}
	if err := database.DB.Model(&pkg).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

func DeletePackage(id uuid.UUID) error {
	// package lobotomy
	var count int64
	database.DB.Model(&models.Server{}).Where("package_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("cannot delete package: %d server(s) are using it", count)
	}

	result := database.DB.Where("id = ?", id).Delete(&models.Package{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPackageNotFound
	}
	return nil
}
