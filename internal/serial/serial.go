package serial

import (
	"log"

	"go.bug.st/serial"
)

// GetPorts trả về danh sách các cổng COM / tty hiện có trên máy
func GetPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Println("Lỗi khi quét cổng Serial:", err)
		return nil, err
	}

	if len(ports) == 0 {
		log.Println("Không tìm thấy cổng Serial nào!")
	} else {
		log.Printf("Đã quét thấy %d cổng: %v\n", len(ports), ports)
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
		log.Printf("Lỗi mở cổng %s: %v\n", portName, err)
		return nil, err
	}

	log.Printf("Đã mở cổng %s với Baudrate %d\n", portName, baudRate)
	return port, nil
}
