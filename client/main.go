package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const serverAddress = "localhost:8080"

type RequestFlight struct {
	ID            int
	Source        string
	Destination   string
	DepartureTime string
	SeattoBook    int
	duration      int
}

type Flight struct {
	ID               int
	Source           string
	Destination      string
	DepartureTime    string
	Airfare          float64
	SeatAvailability int
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to the server at", serverAddress)
	for {
		showMenu()
		handleUserChoice(conn)
	}
}

func showMenu() {
	fmt.Println("\nDistributed Flight Information System")
	fmt.Println("1. Query flights by source and destination")
	fmt.Println("2. Query flight details by flight ID")
	fmt.Println("3. Make a seat reservation")
	fmt.Println("4. Monitor seat availability updates")
	fmt.Println("5. Query points based on IP address")
	fmt.Println("6. Make a seat reservation with points")
	fmt.Println("7. Exit")
	fmt.Print("Enter your choice: ")
}

func handleUserChoice(conn *net.UDPConn) {
	var choice int
	fmt.Scanf("%d", &choice)
	switch choice {
	case 1:
		queryFlights(conn)
	case 2:
		queryFlightDetails(conn)
	case 3:
		makeSeatReservation(conn)
	case 4:
		monitorSeatAvailability(conn)
	case 5:
		queryPoints(conn) // New case for querying points
	case 6:
		makeSeatReservationWithPoints(conn)
	case 7:
		fmt.Println("Exiting...")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice, please try again.")
	}
}

func queryFlights(conn *net.UDPConn) {
	var source, destination string
	fmt.Print("Enter source: ")
	fmt.Scan(&source)
	fmt.Print("Enter destination: ")
	fmt.Scan(&destination)

	request, _ := EncodeClientRequest(RequestFlight{Source: source, Destination: destination}, 1)
	conn.Write(request)

	receiveResponse(conn)
}

func queryFlightDetails(conn *net.UDPConn) {
	var flightID int
	fmt.Print("Enter flight ID: ")
	fmt.Scan(&flightID)

	request, _ := EncodeClientRequest(RequestFlight{ID: flightID}, 2)
	conn.Write(request)

	receiveResponse(conn)
}

func makeSeatReservation(conn *net.UDPConn) {
	var flightID, seats int
	fmt.Print("Enter flight ID: ")
	fmt.Scan(&flightID)
	fmt.Print("Enter number of seats to reserve: ")
	fmt.Scan(&seats)

	request, _ := EncodeClientRequest(RequestFlight{ID: flightID, SeattoBook: seats}, 3)
	conn.Write(request)

	receiveResponse(conn)
}

func makeSeatReservationWithPoints(conn *net.UDPConn) {
	var flightID, seats int
	fmt.Print("Enter flight ID: ")
	fmt.Scan(&flightID)
	fmt.Print("Enter number of seats to reserve: ")
	fmt.Scan(&seats)

	request, _ := EncodeClientRequest(RequestFlight{ID: flightID, SeattoBook: seats}, 6)
	conn.Write(request)

	receiveResponse(conn)
}

func monitorSeatAvailability(conn *net.UDPConn) {
	var flightID, duration int
	fmt.Print("Enter flight ID to monitor: ")
	fmt.Scan(&flightID)
	fmt.Print("Enter monitor duration (in seconds): ")
	fmt.Scan(&duration)

	request, _ := EncodeClientRequest(RequestFlight{ID: flightID, duration: duration}, 4)
	conn.Write(request)

	monitorEndTime := time.Now().Add(time.Duration(duration) * time.Second)
	for time.Now().Before(monitorEndTime) {
		receiveResponse(conn)
	}
}

func queryPoints(conn *net.UDPConn) {
	request, _ := EncodeClientRequest(RequestFlight{}, 5)
	conn.Write(request)

	receiveResponse(conn)
}

func receiveResponse(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from server:", err)
		return
	}

	statuscode, opcode, flights, message, err := decodeServerResponse(buffer[:n])
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	fmt.Printf("Status Code: %d, Opcode: %d\n", statuscode, opcode)
	for _, flight := range flights {
		fmt.Printf("Flight ID: %d, Source: %s, Destination: %s, Departure Time: %s, Airfare: %.2f, Seats Available: %d\n",
			flight.ID, flight.Source, flight.Destination, flight.DepartureTime, flight.Airfare, flight.SeatAvailability)
	}
	fmt.Println("Message:", message)
}

func decodeServerResponse(data []byte) (statuscode, opcode int, flights []Flight, message string, err error) {
	buffer := bytes.NewBuffer(data)

	// Read the statuscode and opcode
	statusByte, _ := buffer.ReadByte()
	statuscode = int(statusByte)
	opByte, _ := buffer.ReadByte()
	opcode = int(opByte)

	// Read number of flights
	flightCount, err := buffer.ReadByte()
	if err != nil {
		return
	}

	// Read flights
	for i := 0; i < int(flightCount); i++ {
		var flight Flight
		flight, err = decodeFlight(buffer)
		if err != nil {
			return
		}
		flights = append(flights, flight)
	}

	// Read the message
	message, err = readString(buffer)
	return
}

func decodeFlight(buffer *bytes.Buffer) (Flight, error) {
	var flight Flight
	// Decode each field of Flight
	idStr, err := readString(buffer)
	if err != nil {
		return flight, err
	}
	flight.ID, _ = strconv.Atoi(idStr)

	flight.Source, err = readString(buffer)
	if err != nil {
		return flight, err
	}

	flight.Destination, err = readString(buffer)
	if err != nil {
		return flight, err
	}

	flight.DepartureTime, err = readString(buffer)
	if err != nil {
		return flight, err
	}

	airfareStr, err := readString(buffer)
	if err != nil {
		return flight, err
	}
	flight.Airfare, _ = strconv.ParseFloat(airfareStr, 64)

	seatsStr, err := readString(buffer)
	if err != nil {
		return flight, err
	}
	flight.SeatAvailability, _ = strconv.Atoi(seatsStr)

	return flight, nil
}

func writeString(buffer *bytes.Buffer, str string) {
	length := byte(len(str))
	buffer.WriteByte(length)
	buffer.WriteString(str)
}

func readString(buffer *bytes.Buffer) (string, error) {
	length, err := buffer.ReadByte()
	if err != nil {
		return "", err
	}

	str := make([]byte, length)
	_, err = buffer.Read(str)
	return string(str), err
}

func EncodeClientRequest(flight RequestFlight, opcode byte) ([]byte, error) {
	buffer := new(bytes.Buffer)

	// Write opcode as a single byte
	if err := binary.Write(buffer, binary.BigEndian, opcode); err != nil {
		return nil, err
	}

	// Encode flight details
	if err := encodeFlight(buffer, flight); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func encodeString(buffer *bytes.Buffer, str string) error {
	length := byte(len(str))
	if err := binary.Write(buffer, binary.BigEndian, length); err != nil {
		return err
	}
	if _, err := buffer.Write([]byte(str)); err != nil {
		return err
	}
	return nil
}

func encodeFlight(buffer *bytes.Buffer, flight RequestFlight) error {
	idStr := fmt.Sprintf("%d", flight.ID)
	if err := encodeString(buffer, idStr); err != nil {
		return err
	}
	if err := encodeString(buffer, flight.Source); err != nil {
		return err
	}
	if err := encodeString(buffer, flight.Destination); err != nil {
		return err
	}
	if err := encodeString(buffer, flight.DepartureTime); err != nil {
		return err
	}
	seatStr := fmt.Sprintf("%d", flight.SeattoBook)
	if err := encodeString(buffer, seatStr); err != nil {
		return err
	}
	durationStr := fmt.Sprintf("%d", flight.duration)
	if err := encodeString(buffer, durationStr); err != nil {
		return err
	}
	return nil
}
