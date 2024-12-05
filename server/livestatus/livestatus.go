package livestatus

import (
	"fmt"
	"io"
	"net"
)

// QueryLivestatus sends a query to the Livestatus tcp socket and returns the response
func QueryLivestatus(query string) (string, error) {
	conn, err := net.Dial("tcp", "localhost:6557")
	if err != nil {
		return "", fmt.Errorf("failed to connect to Livestatus: %v", err)
	}
	defer conn.Close()

	// Add a newline to the query
	query += "\n"

	_, err = conn.Write([]byte(query))
	if err != nil {
		return "", fmt.Errorf("failed to send query: %w", err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(response), nil
}
