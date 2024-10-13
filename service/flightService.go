package service

import (
	"errors"

	"github.com/Guesstrain/airline/models"
	"gorm.io/gorm"
)

type FlightService interface {
	QueryFlights(source, destination string) ([]models.Flight, error)
	GetFlightDetails(flightID int) (*models.Flight, error)
	ReserveSeats(flightID, seats int) (models.Flight, error)
}

type FlightServiceImpl struct {
	DB *gorm.DB
}

// QueryFlights returns flights based on source and destination.
func (f *FlightServiceImpl) QueryFlights(source, destination string) ([]models.Flight, error) {
	var flights []models.Flight
	if err := f.DB.Where("source = ? AND destination = ?", source, destination).Find(&flights).Error; err != nil {
		return nil, err
	}
	return flights, nil
}

// GetFlightDetails returns flight details by flight ID.
func (f *FlightServiceImpl) GetFlightDetails(flightID int) (*models.Flight, error) {
	var flight models.Flight
	if err := f.DB.First(&flight, flightID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Flight not found")
		}
		return nil, err
	}
	return &flight, nil
}

// ReserveSeats reserves a specified number of seats for a flight.
func (f *FlightServiceImpl) ReserveSeats(flightID, seats int) (models.Flight, error) {
	var flight models.Flight
	if err := f.DB.First(&flight, flightID).Error; err != nil {
		return models.Flight{}, errors.New("Flight not found")
	}
	if flight.SeatAvailability < seats {
		return models.Flight{}, errors.New("Insufficient seats available")
	}
	flight.SeatAvailability -= seats
	return flight, f.DB.Save(&flight).Error
}
