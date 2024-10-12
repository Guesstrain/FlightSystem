package models

import (
	"net"
	"time"
)

type Flight struct {
	ID               int     `gorm:"primaryKey"`
	Source           string  `gorm:"size:100;not null"`
	Destination      string  `gorm:"size:100;not null"`
	DepartureTime    string  `gorm:"size:20;not null"` // You can customize this to your preferred time format.
	Airfare          float64 `gorm:"not null"`
	SeatAvailability int     `gorm:"not null"`
}

type ClientInfo struct {
	ClientAddr *net.UDPAddr
	Expiry     time.Time
}
