# STOMP 协议详解

## 概述

STOMP（Simple Text Oriented Messaging Protocol）是一个简单的文本导向消息协议，用于在客户端和消息代理（Message Broker）之间进行通信。它是一个轻量级、易于实现的协议，广泛用于消息队列系统。

## 协议特点

- **简单性**：基于文本的协议，易于理解和调试
- **语言无关**：可以在任何编程语言中实现
- **框架互操作**：不同的消息队列系统都支持 STOMP
- **连接导向**：基于 TCP 连接的可靠传输

## 协议版本

- **STOMP 1.0**：最初版本
- **STOMP 1.1**：添加了心跳机制
- **STOMP 1.2**：当前主流版本，增强了错误处理

## 帧结构

STOMP 协议基于帧（Frame）进行通信，每个帧包含：

```
COMMAND
header1:value1
header2:value2

Body^@
```

### 帧组成部分

1. **命令行**：指定帧的类型
2. **头部**：键值对形式的元数据
3. **空行**：分隔头部和正文
4. **正文**：消息内容
5. **NULL字符**：帧结束标识符（`^@` 表示 NULL 字符）

## 客户端帧类型

### 1. CONNECT / STOMP
建立连接到服务器：

```
CONNECT
accept-version:1.2
host:stomp.github.org
login:username
passcode:password

^@
```

**STOMP 1.2 版本示例**：
```
STOMP
accept-version:1.2
host:stomp.github.org
heart-beat:10000,10000

^@
```

### 2. SEND
发送消息到目的地：

```
SEND
destination:/queue/test
content-type:text/plain
content-length:12

Hello World!^@
```

### 3. SUBSCRIBE
订阅目的地：

```
SUBSCRIBE
id:sub-0
destination:/queue/test
ack:client

^@
```

### 4. UNSUBSCRIBE
取消订阅：

```
UNSUBSCRIBE
id:sub-0

^@
```

### 5. ACK
确认消息：

```
ACK
id:message-12345

^@
```

### 6. NACK
拒绝消息：

```
NACK
id:message-12345

^@
```

### 7. BEGIN
开始事务：

```
BEGIN
transaction:tx1

^@
```

### 8. COMMIT
提交事务：

```
COMMIT
transaction:tx1

^@
```

### 9. ABORT
回滚事务：

```
ABORT
transaction:tx1

^@
```

### 10. DISCONNECT
断开连接：

```
DISCONNECT
receipt:77

^@
```

## 服务器帧类型

### 1. CONNECTED
连接确认：

```
CONNECTED
version:1.2
heart-beat:10000,10000
server:ActiveMQ/5.8.0

^@
```

### 2. MESSAGE
发送消息给客户端：

```
MESSAGE
destination:/queue/test
message-id:12345
subscription:sub-0
content-type:text/plain
content-length:12

Hello World!^@
```

### 3. RECEIPT
收据确认：

```
RECEIPT
receipt-id:77

^@
```

### 4. ERROR
错误信息：

```
ERROR
content-type:text/plain
content-length:42
message:Malformed frame received

The message:
-----
MESSAGE
destined:/queue/test
^@
```

## Go 语言实现示例

### STOMP 客户端基本结构

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "time"
)

type STOMPClient struct {
    conn   net.Conn
    reader *bufio.Reader
}

type STOMPFrame struct {
    Command string
    Headers map[string]string
    Body    []byte
}

func NewSTOMPClient(address string) (*STOMPClient, error) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return nil, err
    }
    
    return &STOMPClient{
        conn:   conn,
        reader: bufio.NewReader(conn),
    }, nil
}

func (c *STOMPClient) Connect(host, login, passcode string) error {
    frame := &STOMPFrame{
        Command: "CONNECT",
        Headers: map[string]string{
            "accept-version": "1.2",
            "host":          host,
            "login":         login,
            "passcode":      passcode,
            "heart-beat":    "10000,10000",
        },
    }
    
    return c.SendFrame(frame)
}

func (c *STOMPClient) Subscribe(destination, id string) error {
    frame := &STOMPFrame{
        Command: "SUBSCRIBE",
        Headers: map[string]string{
            "id":          id,
            "destination": destination,
            "ack":         "client",
        },
    }
    
    return c.SendFrame(frame)
}

func (c *STOMPClient) Send(destination string, body []byte) error {
    frame := &STOMPFrame{
        Command: "SEND",
        Headers: map[string]string{
            "destination":    destination,
            "content-type":   "text/plain",
            "content-length": fmt.Sprintf("%d", len(body)),
        },
        Body: body,
    }
    
    return c.SendFrame(frame)
}

func (c *STOMPClient) SendFrame(frame *STOMPFrame) error {
    frameData := c.EncodeFrame(frame)
    _, err := c.conn.Write(frameData)
    return err
}

func (c *STOMPClient) EncodeFrame(frame *STOMPFrame) []byte {
    var result strings.Builder
    
    // 命令行
    result.WriteString(frame.Command)
    result.WriteString("\n")
    
    // 头部
    for key, value := range frame.Headers {
        result.WriteString(fmt.Sprintf("%s:%s\n", key, value))
    }
    
    // 空行
    result.WriteString("\n")
    
    // 正文
    if len(frame.Body) > 0 {
        result.Write(frame.Body)
    }
    
    // NULL 字符
    result.WriteByte(0)
    
    return []byte(result.String())
}

func (c *STOMPClient) ReadFrame() (*STOMPFrame, error) {
    // 读取命令行
    command, err := c.reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    command = strings.TrimRight(command, "\n\r")
    
    // 读取头部
    headers := make(map[string]string)
    for {
        line, err := c.reader.ReadString('\n')
        if err != nil {
            return nil, err
        }
        
        line = strings.TrimRight(line, "\n\r")
        if line == "" {
            break // 空行表示头部结束
        }
        
        parts := strings.SplitN(line, ":", 2)
        if len(parts) == 2 {
            headers[parts[0]] = parts[1]
        }
    }
    
    // 读取正文
    var body []byte
    if contentLength, ok := headers["content-length"]; ok {
        length := parseInt(contentLength)
        body = make([]byte, length)
        _, err = c.reader.Read(body)
        if err != nil {
            return nil, err
        }
        // 读取 NULL 字符
        c.reader.ReadByte()
    } else {
        // 读取直到 NULL 字符
        for {
            b, err := c.reader.ReadByte()
            if err != nil {
                return nil, err
            }
            if b == 0 {
                break
            }
            body = append(body, b)
        }
    }
    
    return &STOMPFrame{
        Command: command,
        Headers: headers,
        Body:    body,
    }, nil
}

func parseInt(s string) int {
    result := 0
    for _, char := range s {
        if char >= '0' && char <= '9' {
            result = result*10 + int(char-'0')
        }
    }
    return result
}
```

### 消息消费者示例

```go
func (c *STOMPClient) StartConsumer(destination string) {
    // 订阅目的地
    c.Subscribe(destination, "sub-0")
    
    for {
        frame, err := c.ReadFrame()
        if err != nil {
            fmt.Printf("Error reading frame: %v\n", err)
            break
        }
        
        switch frame.Command {
        case "MESSAGE":
            fmt.Printf("Received message: %s\n", string(frame.Body))
            
            // 发送 ACK
            if messageId, ok := frame.Headers["message-id"]; ok {
                ackFrame := &STOMPFrame{
                    Command: "ACK",
                    Headers: map[string]string{
                        "id": messageId,
                    },
                }
                c.SendFrame(ackFrame)
            }
            
        case "ERROR":
            fmt.Printf("Error: %s\n", string(frame.Body))
            
        case "RECEIPT":
            fmt.Printf("Receipt: %s\n", frame.Headers["receipt-id"])
        }
    }
}
```

### 心跳机制实现

```go
func (c *STOMPClient) startHeartbeat(clientHeartbeat, serverHeartbeat int) {
    if clientHeartbeat > 0 {
        go func() {
            ticker := time.NewTicker(time.Duration(clientHeartbeat) * time.Millisecond)
            defer ticker.Stop()
            
            for range ticker.C {
                // 发送心跳帧（空行 + NULL）
                c.conn.Write([]byte("\n\x00"))
            }
        }()
    }
    
    if serverHeartbeat > 0 {
        // 监听服务器心跳
        go func() {
            timeout := time.Duration(serverHeartbeat*2) * time.Millisecond
            timer := time.NewTimer(timeout)
            
            for {
                timer.Reset(timeout)
                select {
                case <-timer.C:
                    fmt.Println("Server heartbeat timeout")
                    c.conn.Close()
                    return
                }
            }
        }()
    }
}
```

## 实际应用示例

### 完整的生产者-消费者示例

```go
func main() {
    // 创建生产者
    producer, err := NewSTOMPClient("localhost:61613")
    if err != nil {
        panic(err)
    }
    defer producer.conn.Close()
    
    // 连接
    err = producer.Connect("localhost", "admin", "admin")
    if err != nil {
        panic(err)
    }
    
    // 发送消息
    for i := 0; i < 10; i++ {
        message := fmt.Sprintf("Message %d", i)
        err = producer.Send("/queue/test", []byte(message))
        if err != nil {
            fmt.Printf("Error sending message: %v\n", err)
        }
        time.Sleep(1 * time.Second)
    }
    
    // 创建消费者
    consumer, err := NewSTOMPClient("localhost:61613")
    if err != nil {
        panic(err)
    }
    defer consumer.conn.Close()
    
    // 连接并开始消费
    err = consumer.Connect("localhost", "admin", "admin")
    if err != nil {
        panic(err)
    }
    
    consumer.StartConsumer("/queue/test")
}
```

## 错误处理和重连机制

```go
func (c *STOMPClient) ConnectWithRetry(host, login, passcode string, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = c.Connect(host, login, passcode)
        if err == nil {
            return nil
        }
        
        fmt.Printf("Connection attempt %d failed: %v\n", i+1, err)
        time.Sleep(time.Duration(i+1) * time.Second)
        
        // 重新建立 TCP 连接
        c.conn.Close()
        c.conn, err = net.Dial("tcp", c.address)
        if err != nil {
            continue
        }
        c.reader = bufio.NewReader(c.conn)
    }
    
    return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}
```

## 常见消息代理支持

### ActiveMQ
```go
client, _ := NewSTOMPClient("localhost:61613")
client.Connect("localhost", "admin", "admin")
```

### RabbitMQ
```go
client, _ := NewSTOMPClient("localhost:61613")
client.Connect("/", "guest", "guest")
```

### Apache Apollo
```go
client, _ := NewSTOMPClient("localhost:61613")
client.Connect("apollo", "admin", "password")
```

## 协议优势

1. **简单易懂**：文本协议，便于调试和理解
2. **互操作性**：广泛支持，不同语言和框架都有实现
3. **可靠性**：基于 TCP 的可靠传输
4. **事务支持**：支持消息事务
5. **确认机制**：支持消息确认，保证消息处理

## 注意事项

1. **转义字符**：头部值需要转义特殊字符
2. **内容长度**：建议总是包含 content-length 头部
3. **心跳机制**：在长连接场景下启用心跳
4. **错误处理**：正确处理 ERROR 帧
5. **连接管理**：实现重连机制

STOMP 协议为消息队列系统提供了标准化的通信方式，是构建分布式消息系统的重要工具。
