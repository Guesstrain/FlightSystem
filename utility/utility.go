package utility

import (
	"fmt"
	"strings"

	"github.com/Guesstrain/airline/models"
)

func EncodeFlightInfo(flight *models.Flight) []byte {
	// Simplified encoding for demonstration
	return []byte(fmt.Sprintf("%d|%s|%s|%v|%f|%d", flight.ID, flight.Source, flight.Destination, flight.DepartureTime, flight.Airfare, flight.SeatAvailability))
}

func DecodeStrings(data []byte) (string, string) {
	// Simplified decoding for demonstration
	parts := strings.SplitN(string(data), "|", 2)
	return parts[0], parts[1]
}
