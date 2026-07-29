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
		log.Println("Error upgrading to WebSocket:", err)
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
			log.Println("Client disconnected:", err)
			break
		}

		// LOG DEBUG TẤT CẢ GÓI TIN NHẬN TỪ WEB
		log.Printf("DEBUG - Received from Web: %+v\n", msg)

		switch msg.Cmd {
		case "list_ports":
			ports, _ := serial.GetPorts()
			sendMsg(WsMessage{Evt: "ports", Data: ports})

		case "connect":
			payload, ok := msg.Data.(map[string]interface{})
			if !ok {
				sendMsg(WsMessage{Evt: "error", Data: "Invalid configuration data"})
				continue
			}
			portName := payload["port"].(string)
			baud := int(payload["baud"].(float64))

			port, err := serial.OpenPort(portName, baud)
			if err != nil {
				sendMsg(WsMessage{Evt: "error", Data: "Error opening port: " + err.Error()})
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
					log.Println("Error: TX data is not a string")
					sendMsg(WsMessage{Evt: "error", Data: "Error formatting TX data"})
					continue
				}

				n, err := activePort.Write([]byte(text))
				if err != nil {
					log.Println("Error writing to Serial:", err)
					sendMsg(WsMessage{Evt: "error", Data: "Error sending data"})
				} else {
					log.Printf("TX SUCCESS: Sent %d bytes: %q\n", n, text)
				}
			} else {
				sendMsg(WsMessage{Evt: "error", Data: "Not connected to COM port"})
			}
		}
	}
}
