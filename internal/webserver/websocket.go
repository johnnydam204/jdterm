package webserver

import (
	"log"
	"net/http"
	"sync"

	"jdterm/internal/serial"

	goSerial "go.bug.st/serial"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WsMessage struct {
	Cmd  string      `json:"cmd,omitempty"`
	Evt  string      `json:"evt,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Lỗi nâng cấp WebSocket:", err)
		return
	}
	defer conn.Close()

	var activePort goSerial.Port
	var stopRead chan struct{}
	var writeMu sync.Mutex

	// Hàm gửi JSON an toàn qua WebSocket
	sendMsg := func(msg WsMessage) {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.WriteJSON(msg)
	}

	defer func() {
		if activePort != nil {
			activePort.Close()
		}
	}()

	for {
		var msg WsMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Client ngắt kết nối:", err)
			break
		}

		// LOG DEBUG TẤT CẢ GÓI TIN NHẬN TỪ WEB
		log.Printf("DEBUG - Nhận từ Web: %+v\n", msg)

		switch msg.Cmd {
		case "list_ports":
			ports, _ := serial.GetPorts()
			sendMsg(WsMessage{Evt: "ports", Data: ports})

		case "connect":
			payload, ok := msg.Data.(map[string]interface{})
			if !ok {
				sendMsg(WsMessage{Evt: "error", Data: "Dữ liệu cấu hình không hợp lệ"})
				continue
			}
			portName := payload["port"].(string)
			baud := int(payload["baud"].(float64))

			port, err := serial.OpenPort(portName, baud)
			if err != nil {
				sendMsg(WsMessage{Evt: "error", Data: "Lỗi mở cổng: " + err.Error()})
				continue
			}

			activePort = port
			stopRead = make(chan struct{})

			// Goroutine hứng RX từ MCU
			go func(p goSerial.Port, stop chan struct{}) {
				buf := make([]byte, 1024)
				for {
					select {
					case <-stop:
						return
					default:
						n, err := p.Read(buf)
						if err != nil {
							sendMsg(WsMessage{Evt: "disconnected"})
							return
						}
						if n > 0 {
							// Bắn RX nguyên bản lên Web
							sendMsg(WsMessage{Evt: "rx", Data: string(buf[:n])})
						}
					}
				}
			}(activePort, stopRead)

			sendMsg(WsMessage{Evt: "connected"})

		case "disconnect":
			if activePort != nil {
				close(stopRead)
				activePort.Close()
				activePort = nil
				sendMsg(WsMessage{Evt: "disconnected"})
			}

		case "tx":
			if activePort != nil {
				text, ok := msg.Data.(string)
				if !ok {
					log.Println("Lỗi: Dữ liệu TX không phải là chuỗi")
					sendMsg(WsMessage{Evt: "error", Data: "Lỗi định dạng TX"})
					continue
				}

				n, err := activePort.Write([]byte(text))
				if err != nil {
					log.Println("Lỗi ghi Serial:", err)
					sendMsg(WsMessage{Evt: "error", Data: "Lỗi gửi dữ liệu"})
				} else {
					log.Printf("TX SUCCESS: Đã gửi %d bytes: %q\n", n, text)
				}
			} else {
				sendMsg(WsMessage{Evt: "error", Data: "Chưa kết nối COM"})
			}
		}
	}
}
