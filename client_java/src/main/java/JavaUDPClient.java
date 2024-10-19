/**
 * Java UDP Client for Distributed Flight Information System.
 * This client allows users to interact with a Go server to perform various operations related to flights,
 * such as querying flight information, reserving seats, and monitoring seat availability.
 * It sends UDP requests to the server and receives responses, simulating a distributed system.
 */
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Random;
import java.util.Scanner;
import java.util.UUID;

public class JavaUDPClient {
    private static final String SERVER_ADDRESS = "192.168.1.53"; // Server address
    private static final int SERVER_PORT = 8080; // Server port number
    private static final int TIMEOUT_MS = 5000; // Timeout for retry mechanism
    private static final Random RANDOM = new Random();

    public static void main(String[] args) {
        try (DatagramSocket socket = new DatagramSocket()) { // Create a UDP socket for communication
            Scanner scanner = new Scanner(System.in);
            InetAddress serverAddress = InetAddress.getByName(SERVER_ADDRESS);

            while (true) {
                showMenu(); // Display user menu
                int choice = scanner.nextInt();
                scanner.nextLine(); // Read newline character to avoid affecting subsequent input

                byte[] requestData = null;
                String requestID = UUID.randomUUID().toString(); // Generate a unique request ID
                switch (choice) {
                    case 1:
                        requestData = createQueryFlightsRequest(scanner, requestID); // Create flight query request
                        break;
                    case 2:
                        requestData = createFlightDetailsRequest(scanner, requestID); // Create flight details query request
                        break;
                    case 3:
                        requestData = createSeatReservationRequest(scanner, requestID); // Create seat reservation request
                        break;
                    case 4:
                        requestData = createMonitorSeatAvailabilityRequest(scanner, requestID); // Create monitor seat availability request
                        sendMonitoringRequest(socket, requestData, serverAddress); // Send monitoring request without retries
                        break;
                    case 5:
                        requestData = createQueryPointsRequest(requestID); // Create query points request
                        break;
                    case 6:
                        requestData = createSeatReservationWithPointsRequest(scanner, requestID); // Create seat reservation with points request
                        break;
                    case 7:
                        System.out.println("Exiting..."); // Exit the program
                        return;
                    default:
                        System.out.println("Invalid choice, please try again."); // Invalid option prompt
                        continue;
                }

                if (requestData != null && choice != 4) { // Apply retry mechanism to services other than monitoring
                    sendRequestWithRetries(socket, requestData, serverAddress);
                }
            }
        } catch (Exception e) {
            e.printStackTrace(); // Catch and print exceptions
        }
    }

    // Display user menu
    private static void showMenu() {
        System.out.println("\nDistributed Flight Information System");
        System.out.println("1. Query flights by source and destination");
        System.out.println("2. Query flight details by flight ID");
        System.out.println("3. Make a seat reservation");
        System.out.println("4. Monitor seat availability updates");
        System.out.println("5. Query points based on IP address");
        System.out.println("6. Make a seat reservation with points");
        System.out.println("7. Exit");
        System.out.print("Enter your choice: ");
    }

    // Create flight query request
    private static byte[] createQueryFlightsRequest(Scanner scanner, String requestID) {
        try {
            System.out.print("Enter source: ");
            String source = scanner.nextLine();
            System.out.print("Enter destination: ");
            String destination = scanner.nextLine();

            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 1); // Opcode 1 represents flight query
            encodeFlight(buffer, new RequestFlight(0, source, destination, "", 0, 0));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Create flight details query request
    private static byte[] createFlightDetailsRequest(Scanner scanner, String requestID) {
        try {
            System.out.print("Enter flight ID: ");
            int flightID = scanner.nextInt();

            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 2); // Opcode 2 represents flight details query
            encodeFlight(buffer, new RequestFlight(flightID, "", "", "", 0, 0));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Create seat reservation request
    private static byte[] createSeatReservationRequest(Scanner scanner, String requestID) {
        try {
            System.out.print("Enter flight ID: ");
            int flightID = scanner.nextInt();
            System.out.print("Enter number of seats to reserve: ");
            int seats = scanner.nextInt();

            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 3); // Opcode 3 represents seat reservation
            encodeFlight(buffer, new RequestFlight(flightID, "", "", "", seats, 0));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Create monitor seat availability request
    private static byte[] createMonitorSeatAvailabilityRequest(Scanner scanner, String requestID) {
        try {
            System.out.print("Enter flight ID to monitor: ");
            int flightID = scanner.nextInt();
            System.out.print("Enter monitor duration (in seconds): ");
            int duration = scanner.nextInt();

            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 4); // Opcode 4 represents monitor seat availability
            encodeFlight(buffer, new RequestFlight(flightID, "", "", "", 0, duration));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Create query points request
    private static byte[] createQueryPointsRequest(String requestID) {
        try {
            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 5); // Opcode 5 represents query points
            encodeFlight(buffer, new RequestFlight(0, "", "", "", 0, 0));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Create seat reservation with points request
    private static byte[] createSeatReservationWithPointsRequest(Scanner scanner, String requestID) {
        try {
            System.out.print("Enter flight ID: ");
            int flightID = scanner.nextInt();
            System.out.print("Enter number of seats to reserve: ");
            int seats = scanner.nextInt();

            ByteBuffer buffer = ByteBuffer.allocate(1024);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian
            buffer.put((byte) 6); // Opcode 6 represents seat reservation with points
            encodeFlight(buffer, new RequestFlight(flightID, "", "", "", seats, 0));
            return trimBuffer(buffer);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Encode flight details into buffer
    private static void encodeFlight(ByteBuffer buffer, RequestFlight flight) {
        encodeString(buffer, String.valueOf(flight.ID));
        encodeString(buffer, flight.Source == null ? "" : flight.Source);
        encodeString(buffer, flight.Destination == null ? "" : flight.Destination);
        encodeString(buffer, flight.DepartureTime == null ? "" : flight.DepartureTime);
        encodeString(buffer, String.valueOf(flight.SeattoBook));
        encodeString(buffer, String.valueOf(flight.duration));
    }

    // Encode string into buffer
    private static void encodeString(ByteBuffer buffer, String str) {
        byte[] strBytes = str.getBytes();
        buffer.put((byte) strBytes.length); // Add string length first
        buffer.put(strBytes); // Then add string content
    }

    // Decode string from buffer
    private static String decodeString(ByteBuffer buffer) {
        try {
            int length = buffer.get(); // Get string length
            byte[] strBytes = new byte[length];
            buffer.get(strBytes); // Read string content
            return new String(strBytes); // Return decoded string
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    // Handle server response
    private static void handleResponse(DatagramPacket responsePacket) {
        try {
            byte[] data = responsePacket.getData();
            ByteBuffer buffer = ByteBuffer.wrap(data);
            buffer.order(ByteOrder.BIG_ENDIAN); // Set byte order to big-endian

            int statusCode = buffer.get(); // Get status code
            int opcode = buffer.get(); // Get opcode

            System.out.printf("Status Code: %d, Opcode: %d\n", statusCode, opcode);

            while (buffer.remaining() > 0) {
                String responseMessage = decodeString(buffer); // Decode string from response
                if (responseMessage != null && !responseMessage.isEmpty()) {
                    System.out.println(responseMessage); // Print response message
                }
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Send request with retry mechanism (for services other than monitoring)
    private static void sendRequestWithRetries(DatagramSocket socket, byte[] requestData, InetAddress serverAddress) {
        int attempts = 0;
        boolean responseReceived = false;
        while (attempts < 3 && !responseReceived) {
            try {
                if (RANDOM.nextInt(100) < 20) { // Simulate 20% chance of packet loss
                    System.out.println("Simulating request loss, not sending packet.");
                } else {
                    DatagramPacket requestPacket = new DatagramPacket(requestData, requestData.length, serverAddress, SERVER_PORT);
                    socket.send(requestPacket);
                }

                socket.setSoTimeout(TIMEOUT_MS); // Set timeout for response
                byte[] responseData = new byte[1024];
                DatagramPacket responsePacket = new DatagramPacket(responseData, responseData.length);
                socket.receive(responsePacket);
                handleResponse(responsePacket);
                responseReceived = true;
            } catch (Exception e) {
                attempts++;
                System.out.printf("No response received, retrying... (attempt %d)\n", attempts);
            }
        }
        if (!responseReceived) {
            System.out.println("Failed to receive response after maximum retries.");
        }
    }

    // Send monitoring request without retry mechanism (used for monitoring seat availability)
    private static void sendMonitoringRequest(DatagramSocket socket, byte[] requestData, InetAddress serverAddress) {
        try {
            DatagramPacket requestPacket = new DatagramPacket(requestData, requestData.length, serverAddress, SERVER_PORT);
            socket.send(requestPacket);

            long startTime = System.currentTimeMillis();
            int monitorDuration = ByteBuffer.wrap(requestData).order(ByteOrder.BIG_ENDIAN).getInt(requestData.length - 4); // Extract monitoring duration from the request data
            boolean responseReceived = false;

            while (System.currentTimeMillis() - startTime < monitorDuration * 1000L) {
                byte[] responseData = new byte[1024];
                DatagramPacket responsePacket = new DatagramPacket(responseData, responseData.length);
                try {
                    socket.setSoTimeout(1000); // Wait for response for 1 second before continuing
                    socket.receive(responsePacket);
                    handleResponse(responsePacket);
                    responseReceived = true;
                } catch (Exception e) {
                    // No response received during this interval, continue waiting
                }
            }

            if (!responseReceived) {
                System.out.println("No response received during this period.");
            }

            System.out.println("Monitoring duration ended. You can now choose another service.");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Trim the ByteBuffer to the actual data length
    private static byte[] trimBuffer(ByteBuffer buffer) {
        byte[] data = new byte[buffer.position()];
        buffer.flip();
        buffer.get(data);
        return data;
    }

    // Flight request structure
    static class RequestFlight {
        int ID;
        String Source;
        String Destination;
        String DepartureTime;
        int SeattoBook;
        int duration;

        RequestFlight(int ID, String Source, String Destination, String DepartureTime, int SeattoBook, int duration) {
            this.ID = ID;
            this.Source = Source;
            this.Destination = Destination;
            this.DepartureTime = DepartureTime;
            this.SeattoBook = SeattoBook;
            this.duration = duration;
        }
    }
}