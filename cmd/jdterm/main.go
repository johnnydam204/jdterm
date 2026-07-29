package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jdterm/internal/webserver"
	"jdterm/web"

	"github.com/pkg/browser"
)

// Hàm tự động tìm một cổng TCP còn trống trên máy tính
func getFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}

func main() {
	// Tự động tìm port trống để tránh xung đột
	port, err := getFreePort()
	if err != nil {
		port = "8080" // Fallback mặc định nếu lỗi
	}

	url := fmt.Sprintf("http://127.0.0.1:%s", port)

	// Khởi chạy HTTP Server & WebSocket trong Goroutine
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", webserver.HandleWebSocket)
		mux.Handle("/", http.FileServer(http.FS(web.FS)))

		log.Printf("JDTerm Server đang chạy tại: %s\n", url)
		if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
			log.Fatalf("Lỗi server: %v", err)
		}
	}()

	// Đợi server kịp khởi động 1 chút rồi tự động mở trình duyệt (App Mode / Default Browser)
	go func() {
		time.Sleep(300 * time.Millisecond)
		log.Printf("Đang mở giao diện JDTerm...")
		err := browser.OpenURL(url)
		if err != nil {
			log.Printf("Không thể tự động mở trình duyệt. Hãy truy cập thủ công: %s\n", url)
		}
	}()

	// Giữ chương trình chạy cho đến khi người dùng tắt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("JDTerm đã tắt.")
}
