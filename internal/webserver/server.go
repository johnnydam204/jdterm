package webserver

import (
	"embed"
	"log"
	"net/http"
)

// Start khởi chạy HTTP Server và WebSocket
func Start(embeddedFiles embed.FS, port string) {
	mux := http.NewServeMux()

	// 1. Route cho WebSocket API
	mux.HandleFunc("/ws", HandleWebSocket)

	// 2. Route phục vụ file tĩnh (Frontend)
	mux.Handle("/", http.FileServer(http.FS(embeddedFiles)))

	log.Printf("Server đang chạy tại: http://localhost:%s", port)

	// Khởi chạy server với bộ định tuyến (mux)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Lỗi khởi chạy server: %v", err)
	}
}
