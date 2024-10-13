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

type RequestFlight struct {
	ID            int
	Source        string
	Destination   string
	DepartureTime string
	SeattoBook    int
	Duration      time.Duration
}

type ClientInfo struct {
	ClientAddr *net.UDPAddr
	Expiry     time.Time
}

type ClientPoints struct {
	ClientAddr string  `gorm:"primaryKey;type:varchar(255)"` // Use string to store the UDP address
	Points     float64 `gorm:"type:double"`                  // Store points as a double
}
