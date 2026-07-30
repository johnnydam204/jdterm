document.addEventListener('DOMContentLoaded', () => {
    const consoleOutput = document.getElementById('console-output');
    const comPortSelect = document.getElementById('com-port');
    const baudRateSelect = document.getElementById('baud-rate');
    const btnConnect = document.getElementById('btn-connect');
    const btnSend = document.getElementById('btn-send');
    const inputField = document.getElementById('serial-input');
    const statusBar = document.getElementById('status-bar');

    let ws = null;
    let isSerialConnected = false;

    function logToConsole(message, type = 'system') {
        const time = new Date().toLocaleTimeString('en-US', { hour12: false });
        const span = document.createElement('span');
        span.textContent = `[${time}] ${message}\n`;
        
        if (type === 'error') span.style.color = '#ff5555';
        else if (type === 'system') span.style.color = '#aaaaaa';
        else if (type === 'tx') span.style.color = '#55ffff';
        
        consoleOutput.appendChild(span);
        consoleOutput.parentElement.scrollTop = consoleOutput.parentElement.scrollHeight;
    }

    function initWebSocket() {
        const wsUrl = `ws://${window.location.host}/ws`;
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            logToConsole('Connected to Backend Core.', 'system');
            ws.send(JSON.stringify({ cmd: 'list_ports' }));
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);

                // Xử lý phân quyền ngay khi kết nối
                if (msg.evt === 'role') {
                    const isHost = msg.data.isHost;
                    if (!isHost) {
                        // Nếu là máy khách (Viewer), tiến hành vô hiệu hóa control
                        comPortSelect.disabled = true;
                        baudRateSelect.disabled = true;
                        btnConnect.disabled = true;
                        btnSend.disabled = true;
                        inputField.disabled = true;
                        
                        btnConnect.style.opacity = '0.5';
                        btnSend.style.opacity = '0.5';
                        inputField.placeholder = "Viewer Mode - Input Disabled";
                        
                        statusBar.textContent = "Status: Connected as Viewer (Read-only)";
                        logToConsole("You are connected as a Viewer (Read-only mode), controls are disabled.", "system");
                    } else {
                        statusBar.textContent = "Status: Connected as Host (Master)";
                    }
                }     

                if (msg.evt === 'ports') {
                    const ports = msg.data || [];
                    comPortSelect.innerHTML = '';
                    if (ports.length === 0) {
                        const opt = document.createElement('option');
                        opt.textContent = 'No Port Found';
                        comPortSelect.appendChild(opt);
                    } else {
                        ports.forEach(port => {
                            const opt = document.createElement('option');
                            opt.value = port;
                            opt.textContent = port;
                            comPortSelect.appendChild(opt);
                        });
                    }
                } 
                else if (msg.evt === 'connected') {
                    isSerialConnected = true;
                    btnConnect.textContent = 'Disconnect';
                    btnConnect.className = 'btn disconnect';
                    comPortSelect.disabled = true;
                    baudRateSelect.disabled = true;
                    logToConsole(`Serial Port is opened successfully: ${comPortSelect.value} @ ${baudRateSelect.value} bps.`, 'system');
                    statusBar.textContent = `Status: Connected to ${comPortSelect.value} @ ${baudRateSelect.value} bps`;
                } 
                else if (msg.evt === 'disconnected') {
                    if (isSerialConnected) { // Kiểm tra cờ trạng thái
                        isSerialConnected = false;
                        btnConnect.textContent = 'Connect';
                        btnConnect.className = 'btn connect';
                        comPortSelect.disabled = false;
                        baudRateSelect.disabled = false;
                        logToConsole('Serial Port is closed.', 'system');
                        statusBar.textContent = 'Status: Port Closed';
                    }
                } 
                else if (msg.evt === 'error') {
                    logToConsole(`Error: ${msg.data}`, 'error');
                }
                else if (msg.evt === 'rx') {
                    const consoleOutput = document.getElementById('console-output');
                    const rxSpan = document.createElement('span');
                    rxSpan.style.color = '#4af626'; // Xanh lá đặc trưng Terminal
                    
                    // Sử dụng TimeStamp đồng bộ từ Server truyền xuống
                    const timeTag = msg.timestamp ? `[${msg.timestamp}] ` : '';
                    rxSpan.textContent = `${timeTag}${msg.data}`;
                    
                    consoleOutput.appendChild(rxSpan);
                    consoleOutput.parentElement.scrollTop = consoleOutput.parentElement.scrollHeight;
                }
            } catch (err) {
                console.error("Error parsing JSON:", err);
            }
        };

        ws.onclose = () => {
            logToConsole('Connection to Backend lost!', 'error');
            statusBar.textContent = 'Status: Backend Disconnected';
            isSerialConnected = false;
        };
    }

    btnConnect.addEventListener('click', () => {
        if (!ws) return;
        
        if (!isSerialConnected) {
            const port = comPortSelect.value;
            const baud = parseInt(baudRateSelect.value);
            
            if (!port || port === "No Port Found") return;
            
            logToConsole(`Connecting to ${port} @ ${baud} bps...`, 'system');
            ws.send(JSON.stringify({ 
                cmd: 'connect', 
                data: { port: port, baud: baud } 
            }));
        } else {
            ws.send(JSON.stringify({ cmd: 'disconnect' }));
        }
    });

    // Lắng nghe sự kiện click nút Send
    btnSend.addEventListener('click', () => {
        if (!ws || !isSerialConnected) {
            logToConsole('Not connected to a serial port!', 'error');
            return;
        }

        const text = inputField.value;
        if (text) {
            logToConsole(`TX: ${text}`, 'tx');
            
            // Gắn \r\n (CRLF) phổ biến cho hầu hết tập lệnh AT / CLI của MCU
            ws.send(JSON.stringify({ 
                cmd: 'tx', 
                data: text + '\r\n' 
            }));
            
            inputField.value = ''; // Xóa ô nhập sau khi gửi
        }
    });

    // Hỗ trợ ấn phím Enter trong ô Input để gửi
    inputField.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') btnSend.click();
    });

    initWebSocket();
});