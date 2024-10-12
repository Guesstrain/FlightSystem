package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Guesstrain/airline/models"
	"github.com/Guesstrain/airline/service"
	"github.com/Guesstrain/airline/utility"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var monitors = map[int][]*models.ClientInfo{}

func main() {
	dsn := "root:password@tcp(127.0.0.1:3306)/airline?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL database:", err)
	}

	flightService := &service.FlightServiceImpl{DB: db}

	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Server listening on port 8080")

	for {
		handleRequest(conn, flightService)
	}
}

func handleRequest(conn *net.UDPConn, service service.FlightService) {
	buffer := make([]byte, 1024)
	_, clientAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error receiving:", err)
		return
	}

	requestType := buffer[0]

	switch requestType {
	case 1: // Query flights by source and destination
		source, destination := utility.DecodeStrings(buffer[1:])
		respondQueryFlights(conn, clientAddr, service, source, destination)
		fmt.Println(clientAddr, "Query flights by source and destination")

	case 2: // Query flight details by flight ID
		flightID := int(binary.BigEndian.Uint32(buffer[1:5]))
		respondFlightDetails(conn, clientAddr, service, flightID)
		fmt.Println(clientAddr, "Query flight details by flight ID")

	case 3: // Make a seat reservation
		flightID := int(binary.BigEndian.Uint32(buffer[1:5]))
		seats := int(binary.BigEndian.Uint32(buffer[5:9]))
		respondSeatReservation(conn, clientAddr, service, flightID, seats)
		fmt.Println(clientAddr, "Make a seat reservation")

	case 4: // Monitor seat availability
		flightID := int(binary.BigEndian.Uint32(buffer[1:5]))
		duration := time.Duration(binary.BigEndian.Uint32(buffer[5:9])) * time.Second
		registerForMonitoring(conn, clientAddr, flightID, duration)
		fmt.Println(clientAddr, "Monitor seat availability")
	}
}

func respondQueryFlights(conn *net.UDPConn, clientAddr *net.UDPAddr, service service.FlightService, source, destination string) {
	flights, err := service.QueryFlights(source, destination)
	if err != nil {
		conn.WriteToUDP([]byte("Error querying flights"), clientAddr)
		return
	}
	if len(flights) == 0 {
		conn.WriteToUDP([]byte("No flights found"), clientAddr)
		return
	}

	var response []byte
	for _, flight := range flights {
		response = append(response, utility.EncodeFlightInfo(&flight)...)
	}
	conn.WriteToUDP(response, clientAddr)
}

func respondFlightDetails(conn *net.UDPConn, clientAddr *net.UDPAddr, service service.FlightService, flightID int) {
	flight, err := service.GetFlightDetails(flightID)
	if err != nil {
		conn.WriteToUDP([]byte("Flight not found"), clientAddr)
		return
	}
	conn.WriteToUDP(utility.EncodeFlightInfo(flight), clientAddr)
}

func respondSeatReservation(conn *net.UDPConn, clientAddr *net.UDPAddr, service service.FlightService, flightID, seats int) {
	availableSeats, err := service.ReserveSeats(flightID, seats)
	if err != nil {
		conn.WriteToUDP([]byte(err.Error()), clientAddr)
		return
	}
	conn.WriteToUDP([]byte("Reservation successful"), clientAddr)
	notifyMonitors(conn, flightID, availableSeats)
}

func registerForMonitoring(conn *net.UDPConn, clientAddr *net.UDPAddr, flightID int, duration time.Duration) {
	clientInfo := &models.ClientInfo{clientAddr, time.Now().Add(duration)}
	monitors[flightID] = append(monitors[flightID], clientInfo)
}

func notifyMonitors(conn *net.UDPConn, flightID int, seats int) {
	clients := monitors[flightID]
	for i, client := range clients {
		if time.Now().Before(client.Expiry) {
			message := fmt.Sprintf("Flight %d seat update: %d", flightID, seats)
			conn.WriteToUDP([]byte(message), client.ClientAddr)
		} else {
			clients = append(clients[:i], clients[i+1:]...)
		}
	}
	monitors[flightID] = clients
}
