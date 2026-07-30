package webserver

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
)

// Hàm tự động tìm một cổng TCP còn trống trên máy tính
func GetFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}

// Start khởi chạy HTTP Server và WebSocket với port linh hoạt
func Start(embeddedFiles embed.FS, address string) error {
	mux := http.NewServeMux()

	// 1. Route cho WebSocket API
	mux.HandleFunc("/ws", HandleWebSocket)

	// 2. Route phục vụ file tĩnh (Frontend)
	mux.Handle("/", http.FileServer(http.FS(embeddedFiles)))

	log.Printf("JDTerm Server is running at: http://%s\n", address)

	// Khởi chạy server
	return http.ListenAndServe(address, mux)
}
