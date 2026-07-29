package serial

import (
	"log"

	"go.bug.st/serial"
)

// GetPorts trả về danh sách các cổng COM / tty hiện có trên máy
func GetPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Println("Error when scanning Serial ports:", err)
		return nil, err
	}

	if len(ports) == 0 {
		log.Println("No Serial ports found!")
	} else {
		log.Printf("Found %d ports: %v\n", len(ports), ports)
	}

	return ports, nil
}

// OpenPort mở kết nối đến một cổng COM với Baudrate được chỉ định
func OpenPort(portName string, baudRate int) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Printf("Error opening port %s: %v\n", portName, err)
		return nil, err
	}

	log.Printf("Port opened: %s with Baudrate %d\n", portName, baudRate)
	return port, nil
}
