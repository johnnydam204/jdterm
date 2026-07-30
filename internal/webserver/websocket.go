package webserver

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	Cmd       string      `json:"cmd,omitempty"`
	Evt       string      `json:"evt,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"` // Thêm trường chứa thời gian
}

// --- QUẢN LÝ DANH SÁCH CLIENT (BROADCAST HUB) ---
type Client struct {
	conn   *websocket.Conn
	isHost bool
}

var (
	clients   = make(map[*Client]bool)
	clientsMu sync.Mutex
)

// Hàm broadcast gửi tin nhắn đến tất cả các client đang kết nối (Cả Host lẫn Viewer)
func BroadcastMessage(msg WsMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		client.conn.WriteJSON(msg)
	}
}

// ------------------------------------------------

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading WebSocket:", err)
		return
	}
	defer conn.Close()

	clientIP := r.RemoteAddr // Thường có dạng "127.0.0.1:xxxxx" hoặc "[::1]:xxxxx"

	// Kiểm tra toàn diện các trường hợp IP của Localhost (Host)
	isHost := false
	if len(clientIP) >= 9 && clientIP[:9] == "127.0.0.1" {
		isHost = true
	} else if len(clientIP) >= 4 && clientIP[:4] == "[::1" {
		isHost = true
	} else if r.Host != "" && (r.Host[:9] == "127.0.0.1" || r.Host[:9] == "localhost") {
		isHost = true
	}

	client := &Client{conn: conn, isHost: isHost}

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
	}()

	log.Printf("Client connected: %s (IsHost: %v)\n", clientIP, isHost)

	// Gửi quyền hạn riêng cho client mới kết nối
	conn.WriteJSON(WsMessage{
		Evt:  "role",
		Data: map[string]any{"isHost": isHost},
	})

	var activePort goSerial.Port
	var stopRead chan struct{}

	for {
		var msg WsMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		switch msg.Cmd {
		case "list_ports":
			ports, _ := serial.GetPorts()
			conn.WriteJSON(WsMessage{Evt: "ports", Data: ports})

		case "connect":
			if !isHost {
				conn.WriteJSON(WsMessage{Evt: "error", Data: "Permission denied! Host-only command."})
				continue
			}
			payload, _ := msg.Data.(map[string]any)
			portName := payload["port"].(string)
			baud := int(payload["baud"].(float64))

			port, err := serial.OpenPort(portName, baud)
			if err != nil {
				conn.WriteJSON(WsMessage{Evt: "error", Data: "Error opening port: " + err.Error()})
				continue
			}

			activePort = port
			stopRead = make(chan struct{})

			// Goroutine đọc RX từ phần cứng với cơ chế tách dòng theo \n hoặc \r\n
			go func(p goSerial.Port, stop chan struct{}) {
				buf := make([]byte, 1024)
				var lineBuffer strings.Builder // Bộ đệm tạm tích lũy ký tự

				for {
					select {
					case <-stop:
						return
					default:
						n, err := p.Read(buf)
						if err != nil {
							BroadcastMessage(WsMessage{Evt: "disconnected"})
							return
						}
						if n > 0 {
							// Đưa dữ liệu đọc được vào bộ đệm
							lineBuffer.Write(buf[:n])
							content := lineBuffer.String()

							// Kiểm tra nếu có chứa ký tự xuống dòng (\n hoặc \r\n)
							if strings.Contains(content, "\n") {
								// Tách lấy các dòng hoàn chỉnh
								lines := strings.Split(content, "\n")

								// Giữ lại phần dư chưa đủ dòng (nếu có) trong buffer
								lineBuffer.Reset()
								lineBuffer.WriteString(lines[len(lines)-1])

								// Lấy chuỗi thời gian hiện tại (VD: "16:25:30.123")
								nowStr := time.Now().Format("15:04:05.000")

								// Phát tán từng dòng hoàn chỉnh kèm TimeStamp đến toàn bộ Host & Viewer
								for i := 0; i < len(lines)-1; i++ {
									cleanLine := strings.TrimRight(lines[i], "\r")
									BroadcastMessage(WsMessage{
										Evt:       "rx",
										Data:      cleanLine + "\n",
										Timestamp: nowStr,
									})
								}
							}
						}
					}
				}
			}(activePort, stopRead)

			BroadcastMessage(WsMessage{Evt: "connected"})

		case "disconnect":
			// Chặn nếu không phải là Host
			if !isHost {
				conn.WriteJSON(WsMessage{Evt: "error", Data: "Permission denied! Host-only command."})
				continue
			}

			if activePort != nil {
				close(stopRead)
				activePort.Close()
				activePort = nil
				BroadcastMessage(WsMessage{Evt: "disconnected"})
			}

		case "tx":
			// Chặn nếu không phải là Host
			if !isHost {
				conn.WriteJSON(WsMessage{Evt: "error", Data: "Permission denied! Host-only command."})
				continue
			}

			if activePort != nil {
				text, ok := msg.Data.(string)
				if ok {
					activePort.Write([]byte(text))
					// Broadcast luôn chuỗi TX để các máy Viewer thấy được nội dung Host vừa gõ gửi đi
					BroadcastMessage(WsMessage{Evt: "tx_log", Data: "TX: " + text})
				}
			}
		}
	}
}
