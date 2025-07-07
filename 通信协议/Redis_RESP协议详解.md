# Redis RESP 协议详解

## 概述

RESP（Redis Serialization Protocol）是 Redis 使用的通信协议。它是一个简单、高效、人类可读的协议，用于客户端和服务器之间的数据交换。

## 协议特点

- **简单性**：协议规则简单，易于实现
- **高效性**：解析速度快，网络开销小
- **可读性**：人类可以直接阅读和理解
- **类型安全**：支持多种数据类型

## 数据类型

RESP 协议支持 5 种数据类型，每种类型都有特定的前缀标识：

### 1. 简单字符串 (Simple Strings) - `+`
```
+OK\r\n
+PONG\r\n
```

### 2. 错误 (Errors) - `-`
```
-Error message\r\n
-ERR unknown command 'helloworld'\r\n
```

### 3. 整数 (Integers) - `:`
```
:0\r\n
:1000\r\n
:-1\r\n
```

### 4. 批量字符串 (Bulk Strings) - `$`
```
$6\r\nfoobar\r\n
$0\r\n\r\n
$-1\r\n  (表示NULL)
```

### 5. 数组 (Arrays) - `*`
```
*0\r\n  (空数组)
*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n  (包含两个元素的数组)
*-1\r\n  (NULL数组)
```

## 请求格式

客户端发送给服务器的命令总是以数组形式表示：

```
*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n
```

对应的 Redis 命令：`SET mykey myvalue`

解析：
- `*3\r\n`：数组包含 3 个元素
- `$3\r\nSET\r\n`：第一个元素是长度为 3 的字符串 "SET"
- `$5\r\nmykey\r\n`：第二个元素是长度为 5 的字符串 "mykey"
- `$7\r\nmyvalue\r\n`：第三个元素是长度为 7 的字符串 "myvalue"

## 响应格式

### 成功响应示例

**SET 命令响应**：
```
+OK\r\n
```

**GET 命令响应**：
```
$7\r\nmyvalue\r\n
```

**LRANGE 命令响应**：
```
*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
```

### 错误响应示例

```
-ERR wrong number of arguments for 'set' command\r\n
```

## Go 语言实现示例

### 简单的 RESP 解析器

```go
package main

import (
    "bufio"
    "fmt"
    "strconv"
    "strings"
)

type RESPParser struct {
    reader *bufio.Reader
}

func NewRESPParser(reader *bufio.Reader) *RESPParser {
    return &RESPParser{reader: reader}
}

func (p *RESPParser) Read() (interface{}, error) {
    line, err := p.reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    
    line = strings.TrimRight(line, "\r\n")
    
    switch line[0] {
    case '+':
        return line[1:], nil
    case '-':
        return fmt.Errorf("error: %s", line[1:]), nil
    case ':':
        return strconv.Atoi(line[1:])
    case '$':
        return p.readBulkString(line)
    case '*':
        return p.readArray(line)
    default:
        return nil, fmt.Errorf("unknown type: %c", line[0])
    }
}

func (p *RESPParser) readBulkString(line string) (string, error) {
    length, err := strconv.Atoi(line[1:])
    if err != nil {
        return "", err
    }
    
    if length == -1 {
        return "", nil // NULL
    }
    
    buffer := make([]byte, length+2) // +2 for \r\n
    _, err = p.reader.Read(buffer)
    if err != nil {
        return "", err
    }
    
    return string(buffer[:length]), nil
}

func (p *RESPParser) readArray(line string) ([]interface{}, error) {
    length, err := strconv.Atoi(line[1:])
    if err != nil {
        return nil, err
    }
    
    if length == -1 {
        return nil, nil // NULL array
    }
    
    array := make([]interface{}, length)
    for i := 0; i < length; i++ {
        array[i], err = p.Read()
        if err != nil {
            return nil, err
        }
    }
    
    return array, nil
}
```

### 简单的 RESP 编码器

```go
func EncodeSimpleString(s string) string {
    return fmt.Sprintf("+%s\r\n", s)
}

func EncodeError(err string) string {
    return fmt.Sprintf("-%s\r\n", err)
}

func EncodeInteger(i int) string {
    return fmt.Sprintf(":%d\r\n", i)
}

func EncodeBulkString(s string) string {
    if s == "" {
        return "$-1\r\n" // NULL
    }
    return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func EncodeArray(arr []string) string {
    if arr == nil {
        return "*-1\r\n" // NULL array
    }
    
    result := fmt.Sprintf("*%d\r\n", len(arr))
    for _, item := range arr {
        result += EncodeBulkString(item)
    }
    return result
}
```

## 实际应用场景

### 1. Redis 客户端实现
```go
func (client *RedisClient) Set(key, value string) error {
    command := EncodeArray([]string{"SET", key, value})
    _, err := client.conn.Write([]byte(command))
    if err != nil {
        return err
    }
    
    response, err := client.parser.Read()
    if err != nil {
        return err
    }
    
    if response == "OK" {
        return nil
    }
    return fmt.Errorf("unexpected response: %v", response)
}
```

### 2. 管道操作
```go
func (client *RedisClient) Pipeline(commands [][]string) ([]interface{}, error) {
    // 发送所有命令
    for _, cmd := range commands {
        encoded := EncodeArray(cmd)
        client.conn.Write([]byte(encoded))
    }
    
    // 读取所有响应
    responses := make([]interface{}, len(commands))
    for i := range commands {
        resp, err := client.parser.Read()
        if err != nil {
            return nil, err
        }
        responses[i] = resp
    }
    
    return responses, nil
}
```

## 协议优势

1. **简单易懂**：协议规则简单，容易理解和实现
2. **高性能**：解析速度快，适合高并发场景
3. **类型安全**：明确的类型标识，避免歧义
4. **可扩展**：易于添加新的数据类型
5. **调试友好**：人类可读，便于调试和监控

## 注意事项

1. **换行符**：必须使用 `\r\n` 作为行结束符
2. **长度前缀**：批量字符串和数组必须包含长度信息
3. **NULL 值**：用 -1 表示 NULL 值
4. **错误处理**：客户端必须正确处理错误响应
5. **连接管理**：长连接复用，避免频繁建立连接

RESP 协议的设计哲学体现了 Redis 追求简单高效的理念，是学习网络协议设计的优秀范例。
