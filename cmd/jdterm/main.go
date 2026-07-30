package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jdterm/internal/webserver"
	"jdterm/web"

	"github.com/pkg/browser"
)

func main() {
	// Tự động tìm port trống để tránh xung đột
	port, err := webserver.GetFreePort()
	if err != nil {
		port = "8080" // Fallback mặc định nếu lỗi
	}

	// Địa chỉ đầy đủ để khởi chạy server
	// Địa chỉ localhost với port được chọn, không truy cập từ bên ngoài
	// address := fmt.Sprintf("127.0.0.1:%s", port)
	// Hoặc
	// address := fmt.Sprintf("localhost:%s", port)

	// Địa chỉ localhost với port được chọn, có thể truy cập từ bên ngoài (nếu cần)
	address := fmt.Sprintf(":%s", port)

	// Địa chỉ URL để mở trình duyệt
	// url := fmt.Sprintf("http://%s", address)
	url := fmt.Sprintf("http://127.0.0.1:%s", port)

	// Gọi hàm Start từ package webserver thay vì viết lại
	go func() {
		if err := webserver.Start(web.FS, address); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Đợi server kịp khởi động 1 chút rồi tự động mở trình duyệt (App Mode / Default Browser)
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Printf("Opening JDTerm interface...")
		err := browser.OpenURL(url)
		if err != nil {
			log.Printf("Cannot automatically open browser. Please access manually: %s\n", url)
		}
	}()

	// In thông báo rõ ràng lên Terminal để dễ theo dõi
	log.Printf("-> Host Access: http://127.0.0.1:%s", port)
	log.Printf("-> Client Access: http://<HOST_PUBLIC_IP>:%s", port)

	// Giữ chương trình chạy cho đến khi người dùng tắt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("JDTerm has been shut down.")
}
