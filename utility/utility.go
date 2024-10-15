package utility

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/Guesstrain/airline/models"
)

func DeserializeFlight(data []byte) (int, models.RequestFlight, string, error) {
	var flight models.RequestFlight
	buffer := bytes.NewBuffer(data)

	// Read the opcode (1 byte)
	opcode, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	fmt.Println("Opcode:", opcode) // You can handle the opcode as needed.
	opcodeInt := int(opcode)
	// Read ID
	idLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	idBytes := make([]byte, idLen)
	if _, err := buffer.Read(idBytes); err != nil {
		return -1, flight, "", err
	}
	flight.ID, err = strconv.Atoi(string(idBytes))
	if err != nil {
		return -1, flight, "", err
	}

	// Read Source
	sourceLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	sourceBytes := make([]byte, sourceLen)
	if _, err := buffer.Read(sourceBytes); err != nil {
		return -1, flight, "", err
	}
	flight.Source = string(sourceBytes)

	// Read Destination
	destinationLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	destinationBytes := make([]byte, destinationLen)
	if _, err := buffer.Read(destinationBytes); err != nil {
		return -1, flight, "", err
	}
	flight.Destination = string(destinationBytes)

	// Read DepartureTime
	departureTimeLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	departureTimeBytes := make([]byte, departureTimeLen)
	if _, err := buffer.Read(departureTimeBytes); err != nil {
		return -1, flight, "", err
	}
	flight.DepartureTime = string(departureTimeBytes)

	// Read SeattoBook
	SeattoBookLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	SeattoBookBytes := make([]byte, SeattoBookLen)
	if _, err := buffer.Read(SeattoBookBytes); err != nil {
		return -1, flight, "", err
	}
	flight.SeattoBook, err = strconv.Atoi(string(SeattoBookBytes))
	if err != nil {
		return -1, flight, "", err
	}

	//Read Duration
	durationLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	durationBytes := make([]byte, durationLen)
	if _, err := buffer.Read(durationBytes); err != nil {
		return -1, flight, "", err
	}
	durationInt, err := strconv.Atoi(string(durationBytes))
	if err != nil {
		return -1, flight, "", err
	}
	flight.Duration = time.Duration(durationInt) * time.Second

	//Read RequestID
	// Read RequestID
	requestIDLen, err := buffer.ReadByte()
	if err != nil {
		return -1, flight, "", err
	}
	requestIDBytes := make([]byte, requestIDLen)
	if _, err := buffer.Read(requestIDBytes); err != nil {
		return -1, flight, "", err
	}
	requestID := string(requestIDBytes)
	fmt.Println("flight:", flight)
	fmt.Println("requestID:", requestID)
	return opcodeInt, flight, requestID, nil
}

func SerializeFlights(flights []models.Flight, opcode, statuscode byte, message string) ([]byte, error) {
	buffer := new(bytes.Buffer)

	// Pack the opcode and statuscode as single bytes
	if err := binary.Write(buffer, binary.BigEndian, statuscode); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, opcode); err != nil {
		return nil, err
	}

	// Write the number of flights as a single byte (assuming less than 256 flights)
	flightCount := byte(len(flights))
	if err := binary.Write(buffer, binary.BigEndian, flightCount); err != nil {
		return nil, err
	}

	// Encode each FlightInfo in the list
	for _, flight := range flights {
		if err := encodeFlight(buffer, flight); err != nil {
			return nil, err
		}
	}

	// Encode the final message string
	if err := encodeString(buffer, message); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Helper function to encode an individual FlightInfo struct
func encodeFlight(buffer *bytes.Buffer, flight models.Flight) error {
	// Encode ID
	idStr := fmt.Sprintf("%d", flight.ID)
	if err := encodeString(buffer, idStr); err != nil {
		return err
	}

	// Encode Source
	if err := encodeString(buffer, flight.Source); err != nil {
		return err
	}

	// Encode Destination
	if err := encodeString(buffer, flight.Destination); err != nil {
		return err
	}

	// Encode DepartureTime
	if err := encodeString(buffer, flight.DepartureTime); err != nil {
		return err
	}

	// Encode Airfare as a string
	airfareStr := fmt.Sprintf("%.2f", flight.Airfare)
	if err := encodeString(buffer, airfareStr); err != nil {
		return err
	}

	// Encode SeatAvailability as a string
	seatStr := fmt.Sprintf("%d", flight.SeatAvailability)
	if err := encodeString(buffer, seatStr); err != nil {
		return err
	}

	return nil
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
