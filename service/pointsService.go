package service

import (
	"errors"

	"github.com/Guesstrain/airline/models"
	"gorm.io/gorm"
)

type PointsService interface {
	QueryPoints(clientAddr string) (*models.ClientPoints, error)
	UpdatePoints(clientAddr string, points float64) (float64, error)
}

type PointsServiceImpl struct {
	DB *gorm.DB
}

func (p *PointsServiceImpl) QueryPoints(clientAddr string) (*models.ClientPoints, error) {
	var clientPoints models.ClientPoints
	if err := p.DB.First(&clientPoints, "client_addr = ?", clientAddr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("No points record found for this IP address")
		}
		return nil, err
	}
	return &clientPoints, nil
}

func (p *PointsServiceImpl) UpdatePoints(clientAddr string, points float64) (float64, error) {
	var clientPoints models.ClientPoints

	// Try to find existing points record
	if err := p.DB.First(&clientPoints, "client_addr = ?", clientAddr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record does not exist, create a new one
			clientPoints = models.ClientPoints{
				ClientAddr: clientAddr,
				Points:     points,
			}
			if err := p.DB.Create(&clientPoints).Error; err != nil {
				return 0, err
			}
			return points, nil
		}
		return 0, err
	}

	// Record exists, update points
	clientPoints.Points = points
	if err := p.DB.Save(&clientPoints).Error; err != nil {
		return 0, err
	}
	return clientPoints.Points, nil
}
