package main

import (
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
		handleRequest(conn, db)
	}
}

func handleRequest(conn *net.UDPConn, db *gorm.DB) {
	flightService := &service.FlightServiceImpl{DB: db}
	pointsService := &service.PointsServiceImpl{DB: db}
	buffer := make([]byte, 1024)
	_, clientAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error receiving:", err)
		return
	}
	requestType, flight, err := utility.DeserializeFlight(buffer)
	if err != nil {
		fmt.Println("Error DeserializeFlight:", err)
		return
	}
	fmt.Println("flight info is", flight)

	switch requestType {
	case 1: // Query flights by source and destination
		respondQueryFlights(conn, clientAddr, flightService, flight.Source, flight.Destination)
		fmt.Println(clientAddr, "Query flights by source and destination")

	case 2: // Query flight details by flight ID
		respondFlightDetails(conn, clientAddr, flightService, flight.ID)
		fmt.Println(clientAddr, "Query flight details by flight ID")

	case 3: // Make a seat reservation
		respondSeatReservation(conn, clientAddr, flightService, pointsService, flight.ID, flight.SeattoBook)
		fmt.Println(clientAddr, "Make a seat reservation")

	case 4: // Monitor seat availability
		registerForMonitoring(conn, clientAddr, flight.ID, flight.Duration)
		fmt.Println(clientAddr, "Monitor seat availability")

	case 5: // Query points based on client address
		respondQueryPoints(conn, clientAddr, pointsService)
		fmt.Println(clientAddr, "Queried points")

	case 6: // Make a seat reservation with points
		respondUsingPoints(conn, clientAddr, flightService, pointsService, flight.ID, flight.SeattoBook)
		fmt.Println(clientAddr, "Make a seat reservation with points")
	}
}

func respondQueryFlights(conn *net.UDPConn, clientAddr *net.UDPAddr, service service.FlightService, source, destination string) {
	flights, err := service.QueryFlights(source, destination)
	if err != nil {
		response, _ := utility.SerializeFlights(flights, 1, 1, "Error querying flights")
		conn.WriteToUDP(response, clientAddr)
		return
	}
	if len(flights) == 0 {
		response, _ := utility.SerializeFlights(flights, 1, 0, "No flights found")
		conn.WriteToUDP(response, clientAddr)
		return
	}

	var response []byte
	response, _ = utility.SerializeFlights(flights, 1, 0, "Success")
	conn.WriteToUDP(response, clientAddr)
}

func respondFlightDetails(conn *net.UDPConn, clientAddr *net.UDPAddr, service service.FlightService, flightID int) {
	flight, err := service.GetFlightDetails(flightID)
	flights := []models.Flight{*flight}
	if err != nil {
		response, _ := utility.SerializeFlights(flights, 2, 1, "No flights found")
		conn.WriteToUDP(response, clientAddr)
		return
	}
	response, _ := utility.SerializeFlights(flights, 2, 0, "Success")
	conn.WriteToUDP(response, clientAddr)
}

func respondUsingPoints(conn *net.UDPConn, clientAddr *net.UDPAddr, flightService service.FlightService, pointsService service.PointsService, flightID, seats int) {
	flight, err := flightService.GetFlightDetails(flightID)
	if err != nil {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, err.Error())
		conn.WriteToUDP(response, clientAddr)
		return
	}
	clientPoints, _ := pointsService.QueryPoints(clientAddr.String())
	if clientPoints.Points < (flight.Airfare * float64(seats)) {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, "Not Enough Points")
		conn.WriteToUDP(response, clientAddr)
		return
	}
	*flight, err = flightService.ReserveSeats(flightID, seats)
	if err != nil {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, err.Error())
		conn.WriteToUDP(response, clientAddr)
		return
	}
	clientPoints.Points = clientPoints.Points - flight.Airfare*float64(seats)
	fmt.Println("clientPoints.Points: ", clientPoints.Points)
	fmt.Println("flight.Airfare: ", flight.Airfare)
	_, err = pointsService.UpdatePoints(clientAddr.String(), clientPoints.Points)
	if err != nil {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, err.Error())
		conn.WriteToUDP(response, clientAddr)
		return
	}

	response, _ := utility.SerializeFlights([]models.Flight{}, 3, 0, "Reservation using points successful")
	conn.WriteToUDP(response, clientAddr)
	notifyMonitors(conn, flightID, flight.SeatAvailability)
}

func respondSeatReservation(conn *net.UDPConn, clientAddr *net.UDPAddr, flightService service.FlightService, pointsService service.PointsService, flightID, seats int) {
	flight, err := flightService.ReserveSeats(flightID, seats)
	if err != nil {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, err.Error())
		conn.WriteToUDP(response, clientAddr)
		return
	}
	clientPoints, _ := pointsService.QueryPoints(clientAddr.String())
	clientPoints.Points = clientPoints.Points + flight.Airfare*float64(seats)
	_, err = pointsService.UpdatePoints(clientAddr.String(), clientPoints.Points)
	if err != nil {
		response, _ := utility.SerializeFlights([]models.Flight{}, 3, 1, err.Error())
		conn.WriteToUDP(response, clientAddr)
		return
	}

	response, _ := utility.SerializeFlights([]models.Flight{}, 3, 0, "Reservation successful")
	conn.WriteToUDP(response, clientAddr)
	notifyMonitors(conn, flightID, flight.SeatAvailability)
}

func registerForMonitoring(conn *net.UDPConn, clientAddr *net.UDPAddr, flightID int, duration time.Duration) {
	clientInfo := &models.ClientInfo{clientAddr, time.Now().Add(duration)}
	fmt.Println("New register for monitoring: ", clientInfo)
	monitors[flightID] = append(monitors[flightID], clientInfo)
}

func respondQueryPoints(conn *net.UDPConn, clientAddr *net.UDPAddr, pointsService service.PointsService) {
	// Query points using the client address as a string
	points, err := pointsService.QueryPoints(clientAddr.String())
	if err != nil {
		fmt.Println("Error querying points:", err)
		response, _ := utility.SerializeFlights([]models.Flight{}, 5, 1, err.Error())
		conn.WriteToUDP([]byte(response), clientAddr)
		return
	}

	// Format the response message
	response, _ := utility.SerializeFlights([]models.Flight{}, 5, 0, fmt.Sprintf("%.2f", points.Points))
	conn.WriteToUDP([]byte(response), clientAddr)
}

func notifyMonitors(conn *net.UDPConn, flightID int, seats int) {
	clients := monitors[flightID]
	for i, client := range clients {
		if time.Now().Before(client.Expiry) {
			message := fmt.Sprintf("Flight %d seat update: %d", flightID, seats)
			response, _ := utility.SerializeFlights([]models.Flight{}, 4, 0, message)
			conn.WriteToUDP(response, client.ClientAddr)
		} else {
			clients = append(clients[:i], clients[i+1:]...)
		}
	}
	monitors[flightID] = clients
}
