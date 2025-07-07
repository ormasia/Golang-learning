# WebSocket 协议详解

## 概述

WebSocket 是一种在单个 TCP 连接上进行全双工通信的协议。它使得客户端和服务器之间的数据交换变得更加简单，允许服务端主动向客户端推送数据。在 WebSocket API 中，浏览器和服务器只需要完成一次握手，两者之间就直接可以创建持久性的连接，并进行双向数据传输。

## 协议特点

- **全双工通信**：客户端和服务器可以同时发送数据
- **较少的控制开销**：连接建立后，数据传输时只需很少的协议开销
- **更强的实时性**：服务器可以主动推送数据给客户端
- **更好的二进制支持**：可以发送文本和二进制数据
- **支持扩展**：可以实现自定义的子协议

## 协议握手过程

WebSocket 握手基于 HTTP 协议，使用 HTTP Upgrade 机制将连接从 HTTP 升级为 WebSocket。

### 客户端握手请求

```http
GET /chat HTTP/1.1
Host: server.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Protocol: chat, superchat
Sec-WebSocket-Version: 13
Origin: http://example.com
```

### 服务器握手响应

```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
Sec-WebSocket-Protocol: chat
```

### 握手过程详解

1. **客户端发起请求**：
   - `Upgrade: websocket`：请求升级到 WebSocket 协议
   - `Connection: Upgrade`：表示希望升级连接
   - `Sec-WebSocket-Key`：客户端生成的随机字符串
   - `Sec-WebSocket-Version`：WebSocket 协议版本（当前为 13）

2. **服务器响应**：
   - `101 Switching Protocols`：表示协议切换成功
   - `Sec-WebSocket-Accept`：根据客户端的 Key 计算得出的值

### Sec-WebSocket-Accept 计算方法

```go
func calculateWebSocketAccept(key string) string {
    const websocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
    h := sha1.New()
    h.Write([]byte(key + websocketMagicString))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
```

## 数据帧格式

WebSocket 使用帧（Frame）来传输数据，每个帧包含以下结构：

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - - +-------------------------------+
|                               |Masking-key, if MASK set to 1  |
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - - +
:                     Payload Data continued ...                :
+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
|                     Payload Data continued ...                |
+---------------------------------------------------------------+
```

### 帧字段说明

- **FIN (1 bit)**：表示这是消息的最后一个分片
- **RSV1-3 (3 bits)**：保留位，必须为 0
- **Opcode (4 bits)**：操作码，定义帧类型
- **MASK (1 bit)**：是否使用掩码（客户端发送的帧必须设置为 1）
- **Payload Length (7 bits, 7+16 bits, 7+64 bits)**：负载数据长度
- **Masking-key (4 bytes)**：掩码密钥（如果 MASK 为 1）
- **Payload Data**：负载数据

### 操作码定义

- `0x0`：继续帧（Continuation Frame）
- `0x1`：文本帧（Text Frame）
- `0x2`：二进制帧（Binary Frame）
- `0x8`：连接关闭帧（Close Frame）
- `0x9`：Ping 帧
- `0xA`：Pong 帧

## Go 语言实现示例

### 基本的 WebSocket 服务器

```go
package main

import (
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "net/http"
    "strings"
)

const (
    websocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

type WebSocketServer struct {
    clients map[*WebSocketConn]bool
}

type WebSocketConn struct {
    conn   net.Conn
    server *WebSocketServer
}

func NewWebSocketServer() *WebSocketServer {
    return &WebSocketServer{
        clients: make(map[*WebSocketConn]bool),
    }
}

func (s *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // 验证 WebSocket 握手
    if !s.checkHeaders(r) {
        http.Error(w, "Bad Request", 400)
        return
    }
    
    // 计算 Sec-WebSocket-Accept
    key := r.Header.Get("Sec-WebSocket-Key")
    accept := s.calculateAccept(key)
    
    // 获取底层连接
    hj, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Server doesn't support hijacking", 500)
        return
    }
    
    conn, _, err := hj.Hijack()
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    // 发送握手响应
    response := fmt.Sprintf(
        "HTTP/1.1 101 Switching Protocols\r\n"+
            "Upgrade: websocket\r\n"+
            "Connection: Upgrade\r\n"+
            "Sec-WebSocket-Accept: %s\r\n\r\n",
        accept)
    
    conn.Write([]byte(response))
    
    // 创建 WebSocket 连接
    wsConn := &WebSocketConn{
        conn:   conn,
        server: s,
    }
    
    s.clients[wsConn] = true
    go wsConn.handleMessages()
}

func (s *WebSocketServer) checkHeaders(r *http.Request) bool {
    return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
        strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") &&
        r.Header.Get("Sec-WebSocket-Version") == "13" &&
        r.Header.Get("Sec-WebSocket-Key") != ""
}

func (s *WebSocketServer) calculateAccept(key string) string {
    h := sha1.New()
    h.Write([]byte(key + websocketMagicString))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
```

### WebSocket 帧处理

```go
type Frame struct {
    Fin    bool
    Opcode uint8
    Masked bool
    Mask   [4]byte
    Length uint64
    Data   []byte
}

func (c *WebSocketConn) readFrame() (*Frame, error) {
    // 读取前两个字节
    header := make([]byte, 2)
    _, err := c.conn.Read(header)
    if err != nil {
        return nil, err
    }
    
    frame := &Frame{}
    
    // 解析第一个字节
    frame.Fin = (header[0] & 0x80) != 0
    frame.Opcode = header[0] & 0x0F
    
    // 解析第二个字节
    frame.Masked = (header[1] & 0x80) != 0
    length := uint64(header[1] & 0x7F)
    
    // 读取扩展长度
    if length == 126 {
        extLen := make([]byte, 2)
        _, err := c.conn.Read(extLen)
        if err != nil {
            return nil, err
        }
        length = uint64(extLen[0])<<8 | uint64(extLen[1])
    } else if length == 127 {
        extLen := make([]byte, 8)
        _, err := c.conn.Read(extLen)
        if err != nil {
            return nil, err
        }
        for i := 0; i < 8; i++ {
            length = length<<8 | uint64(extLen[i])
        }
    }
    frame.Length = length
    
    // 读取掩码
    if frame.Masked {
        _, err := c.conn.Read(frame.Mask[:])
        if err != nil {
            return nil, err
        }
    }
    
    // 读取负载数据
    if length > 0 {
        frame.Data = make([]byte, length)
        _, err := c.conn.Read(frame.Data)
        if err != nil {
            return nil, err
        }
        
        // 如果有掩码，解除掩码
        if frame.Masked {
            for i := uint64(0); i < length; i++ {
                frame.Data[i] ^= frame.Mask[i%4]
            }
        }
    }
    
    return frame, nil
}

func (c *WebSocketConn) writeFrame(opcode uint8, data []byte) error {
    length := len(data)
    
    // 计算帧头长度
    headerLen := 2
    if length >= 65536 {
        headerLen += 8
    } else if length >= 126 {
        headerLen += 2
    }
    
    header := make([]byte, headerLen)
    
    // 设置第一个字节（FIN + Opcode）
    header[0] = 0x80 | opcode
    
    // 设置长度字段
    if length >= 65536 {
        header[1] = 127
        for i := 0; i < 8; i++ {
            header[9-i] = byte(length >> uint(i*8))
        }
    } else if length >= 126 {
        header[1] = 126
        header[2] = byte(length >> 8)
        header[3] = byte(length)
    } else {
        header[1] = byte(length)
    }
    
    // 发送帧头
    _, err := c.conn.Write(header)
    if err != nil {
        return err
    }
    
    // 发送数据
    if length > 0 {
        _, err = c.conn.Write(data)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 消息处理

```go
func (c *WebSocketConn) handleMessages() {
    defer func() {
        delete(c.server.clients, c)
        c.conn.Close()
    }()
    
    for {
        frame, err := c.readFrame()
        if err != nil {
            fmt.Printf("Error reading frame: %v\n", err)
            break
        }
        
        switch frame.Opcode {
        case 0x1: // 文本消息
            message := string(frame.Data)
            fmt.Printf("Received text message: %s\n", message)
            c.broadcast(frame.Opcode, frame.Data)
            
        case 0x2: // 二进制消息
            fmt.Printf("Received binary message: %d bytes\n", len(frame.Data))
            c.broadcast(frame.Opcode, frame.Data)
            
        case 0x8: // 关闭连接
            fmt.Println("Connection close requested")
            c.writeFrame(0x8, []byte{})
            return
            
        case 0x9: // Ping
            fmt.Println("Received ping")
            c.writeFrame(0xA, frame.Data) // 回复 Pong
            
        case 0xA: // Pong
            fmt.Println("Received pong")
        }
    }
}

func (c *WebSocketConn) broadcast(opcode uint8, data []byte) {
    for client := range c.server.clients {
        if client != c {
            err := client.writeFrame(opcode, data)
            if err != nil {
                fmt.Printf("Error broadcasting to client: %v\n", err)
                delete(c.server.clients, client)
                client.conn.Close()
            }
        }
    }
}
```

### 完整的聊天服务器示例

```go
func main() {
    server := NewWebSocketServer()
    
    http.HandleFunc("/ws", server.HandleConnection)
    http.HandleFunc("/", serveHTML)
    
    fmt.Println("WebSocket server starting on :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        panic(err)
    }
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
    html := `
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Chat</title>
</head>
<body>
    <div id="messages"></div>
    <input type="text" id="messageInput" placeholder="输入消息...">
    <button onclick="sendMessage()">发送</button>
    
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        const messages = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        
        ws.onmessage = function(event) {
            const div = document.createElement('div');
            div.textContent = event.data;
            messages.appendChild(div);
        };
        
        function sendMessage() {
            const message = messageInput.value;
            if (message) {
                ws.send(message);
                messageInput.value = '';
            }
        }
        
        messageInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>
    `
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

## 使用 Gorilla WebSocket 库

在实际项目中，建议使用成熟的 WebSocket 库，如 Gorilla WebSocket：

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许所有来源
    },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()
    
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        
        log.Printf("Received: %s", message)
        
        err = conn.WriteMessage(messageType, message)
        if err != nil {
            log.Println("Write error:", err)
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 心跳机制

```go
func (c *WebSocketConn) startHeartbeat() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := c.writeFrame(0x9, []byte("ping")); err != nil {
                fmt.Printf("Ping error: %v\n", err)
                return
            }
        }
    }
}
```

## 实际应用场景

### 1. 实时聊天
```go
type ChatRoom struct {
    clients   map[*websocket.Conn]bool
    broadcast chan []byte
    register  chan *websocket.Conn
    unregister chan *websocket.Conn
}

func (room *ChatRoom) run() {
    for {
        select {
        case client := <-room.register:
            room.clients[client] = true
            
        case client := <-room.unregister:
            if _, ok := room.clients[client]; ok {
                delete(room.clients, client)
                client.Close()
            }
            
        case message := <-room.broadcast:
            for client := range room.clients {
                if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
                    delete(room.clients, client)
                    client.Close()
                }
            }
        }
    }
}
```

### 2. 实时数据推送
```go
func (s *StockServer) pushStockData() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stockData := s.getLatestStockData()
        jsonData, _ := json.Marshal(stockData)
        
        for client := range s.clients {
            client.WriteMessage(websocket.TextMessage, jsonData)
        }
    }
}
```

### 3. 游戏状态同步
```go
type GameServer struct {
    players map[string]*Player
    rooms   map[string]*GameRoom
}

func (gs *GameServer) syncGameState(roomID string) {
    room := gs.rooms[roomID]
    gameState := room.getGameState()
    stateData, _ := json.Marshal(gameState)
    
    for _, player := range room.players {
        player.conn.WriteMessage(websocket.TextMessage, stateData)
    }
}
```

## 安全考虑

### 1. 跨站点 WebSocket 劫持（CSWSH）
```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://yourdomain.com"
    },
}
```

### 2. 认证和授权
```go
func authenticate(r *http.Request) (*User, error) {
    token := r.Header.Get("Authorization")
    if token == "" {
        return nil, errors.New("missing authorization token")
    }
    
    return validateToken(token)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    user, err := authenticate(r)
    if err != nil {
        http.Error(w, "Unauthorized", 401)
        return
    }
    
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    // 处理已认证的连接
    handleAuthenticatedConnection(conn, user)
}
```

### 3. 消息限速
```go
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.Mutex
}

func (rl *RateLimiter) Allow(clientID string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    requests := rl.requests[clientID]
    
    // 清理过期的请求记录
    var validRequests []time.Time
    for _, reqTime := range requests {
        if now.Sub(reqTime) < time.Minute {
            validRequests = append(validRequests, reqTime)
        }
    }
    
    if len(validRequests) >= 60 { // 每分钟最多60个请求
        return false
    }
    
    validRequests = append(validRequests, now)
    rl.requests[clientID] = validRequests
    return true
}
```

## 协议优势

1. **实时性**：全双工通信，服务器可主动推送
2. **效率**：协议开销小，适合频繁通信
3. **兼容性**：基于 HTTP 握手，易于部署
4. **扩展性**：支持自定义子协议
5. **二进制支持**：可传输任意格式数据

## 注意事项

1. **连接管理**：合理处理连接的建立和断开
2. **错误处理**：网络异常、协议错误的处理
3. **资源限制**：防止连接数过多耗尽资源
4. **安全防护**：认证、授权、防止攻击
5. **监控告警**：连接数、消息量、错误率监控

WebSocket 协议为实时Web应用提供了强大的基础，是构建现代实时应用的重要技术。
