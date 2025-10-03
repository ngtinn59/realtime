# Hệ thống Real-time Chat trong ERP API

## Tổng quan

Dự án ERP API triển khai hệ thống chat thời gian thực sử dụng **WebSocket** kết hợp với **Redis** để quản lý trạng thái online/offline của người dùng và đồng bộ hóa tin nhắn giữa các client.

## ✅ Implementation Status

**✅ Đã hoàn thành:**
- WebSocket server với Hub architecture
- Redis integration cho trạng thái online/offline
- HTTP API endpoints cho chat
- **✅ Real-time message broadcasting** từ HTTP APIs đến WebSocket clients
- Client connection management với ping/pong
- Message queuing và error handling cơ bản

**⚠️ Có thể cải thiện thêm:**
- Group message chỉ broadcast tới group members cụ thể (hiện tại broadcast tới tất cả clients)
- Rate limiting cho message gửi
- Message persistence trong Redis để handle reconnection
- Advanced error handling và retry mechanism.

## Kiến trúc tổng thể

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
{{ ... }}
│   Client App    │    │   WebSocket     │    │     Redis       │
│                 │◄──►│   Server        │◄──►│   Cache         │
│ - Browser       │    │                 │    │                 │
│ - Mobile App    │    │ - Hub           │    │ - Online Status │
│ - Desktop App   │    │ - Clients       │    │ - Typing Status │
└─────────────────┘    │ - Message       │    │ - Pub/Sub       │
                       │   Routing       │    └─────────────────┘
┌─────────────────┐    └─────────────────┘
│   Database      │
│                 │
│ - PostgreSQL    │
│ - Messages      │
│ - Users         │
│ - Groups        │
└─────────────────┘
```

## Các thành phần chính

### 1. WebSocket Server
- **Thư viện**: `github.com/gorilla/websocket`
- **Endpoint**: `GET /ws?token={jwt_token}`
- **Chức năng**: Xử lý kết nối WebSocket từ client

### 2. Hub (Central Message Router)
- **Vị trí**: `internal/pkg/websocket/hub.go`
- **Chức năng**:
  - Quản lý các client kết nối
  - Định tuyến tin nhắn
  - Đồng bộ trạng thái online/offline

### 3. Redis Cache
- **Vị trí**: `internal/pkg/redis/client.go`
- **Chức năng**:
  - Lưu trữ trạng thái online/offline
  - Quản lý trạng thái typing
  - Hỗ trợ pub/sub cho scale

### 4. Client Connection
- **Chức năng**:
  - Xử lý giao tiếp với từng client
  - Ping/Pong để duy trì kết nối
  - Message queuing
  - Xử lý broadcast message từ HTTP API

## Cách hoạt động hiện tại

### 1. Client kết nối WebSocket
- Client gửi JWT token qua query parameter
- Server xác thực và tạo WebSocket connection
- Client được đăng ký với Hub và đánh dấu online

### 2. HTTP API gửi message
```go
// Khi gửi private message qua HTTP API
func (s *ChatService) SendPrivateMessage(senderID uint, req SendPrivateMessageRequest) {
    // Lưu message vào database
    message := models.PrivateMessage{...}
    db.Create(&message)
    
    // Broadcast real-time tới WebSocket clients
    messageData := map[string]interface{}{...}
    websocket.BroadcastPrivateMessage(senderID, req.ReceiverID, messageData)
}
```

### 2. Hub quản lý kết nối

```go
// Hub xử lý đăng ký client mới
func (h *Hub) registerClient(client *Client) {
    h.mu.Lock()
    h.Clients[client.UserID] = client
    h.mu.Unlock()

    // Đặt trạng thái online trong Redis
    redis.SetUserOnline(client.UserID)

    // Broadcast trạng thái online cho tất cả client khác
    h.broadcastUserStatus(client.UserID, true)
}
```

### 3. Gửi và nhận tin nhắn

**Client gửi tin nhắn:**
```javascript
// Client gửi message qua WebSocket
ws.send(JSON.stringify({
    event: "send_private_message",
    data: {
        receiver_id: 123,
        content: "Hello!",
        type: "text"
    }
}));
```

**Server xử lý:**
```go
// Client nhận và xử lý message
func (c *Client) ReadPump() {
    for {
        _, message, err := c.Conn.ReadMessage()
        var msg Message
        json.Unmarshal(message, &msg)

        // Gửi tới Hub để xử lý
        c.Hub.Broadcast <- BroadcastMessage{
            Message:  msg,
            SenderID: c.UserID,
        }
    }
}

// Hub xử lý và định tuyến
func (h *Hub) handlePrivateMessage(bm BroadcastMessage) {
    receiverID := bm.Message.Data["receiver_id"].(float64)

    // Gửi tới người nhận nếu online
    h.SendToUser(uint(receiverID), "private_message", bm.Message.Data)

    // Gửi lại cho người gửi để xác nhận
    h.SendToUser(bm.SenderID, "message_sent", bm.Message.Data)
}
```

## Cấu trúc Message

```go
type Message struct {
    Event string                 `json:"event"`
    Data  map[string]interface{} `json:"data"`
}
```

### Các loại Event

| Event | Mô tả | Data |
|-------|-------|------|
| `private_message` | Tin nhắn riêng tư | `receiver_id`, `content`, `type`, `file_id` |
| `group_message` | Tin nhắn nhóm | `group_id`, `content`, `type`, `file_id` |
| `user_typing` | Đang nhập | `conversation_id`, `receiver_id` |
| `user_online_status` | Trạng thái online | `user_id`, `is_online` |
| `message_sent` | Xác nhận gửi | Thông tin message |

## Redis Integration

### Trạng thái Online/Offline

```go
// Đặt user online với key không hết hạn
func SetUserOnline(userID uint) error {
    key := fmt.Sprintf("user:online:%d", userID)
    return Client.Set(ctx, key, "1", 0).Err()
}

// Kiểm tra user có online không
func IsUserOnline(userID uint) (bool, error) {
    key := fmt.Sprintf("user:online:%d", userID)
    result, err := Client.Exists(ctx, key).Result()
    return result > 0, err
}
```

### Trạng thái Typing

```go
// Đặt trạng thái typing với TTL 10 giây
func SetUserTyping(userID uint, conversationID string) error {
    key := fmt.Sprintf("typing:%s:%d", conversationID, userID)
    return Client.Set(ctx, key, "1", 10*time.Second).Err()
}
```

## Client-side Integration

### Kết nối WebSocket

```javascript
class ChatWebSocket {
    constructor(token) {
        this.token = token;
        this.ws = null;
        this.reconnectAttempts = 0;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?token=${this.token}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.attemptReconnect();
        };
    }

    sendMessage(event, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ event, data }));
        }
    }

    handleMessage(message) {
        switch(message.event) {
            case 'private_message':
                this.displayPrivateMessage(message.data);
                break;
            case 'user_online_status':
                this.updateUserStatus(message.data);
                break;
            case 'user_typing':
                this.showTypingIndicator(message.data);
                break;
        }
    }
}
```

### Authentication với JWT

```javascript
// Lấy token từ login
const token = localStorage.getItem('jwt_token');

// Khởi tạo WebSocket với token
const chatWS = new ChatWebSocket(token);
chatWS.connect();

// Gửi tin nhắn
chatWS.sendMessage('send_private_message', {
    receiver_id: 123,
    content: 'Hello!',
    type: 'text'
});
```

## Error Handling và Reconnection

### Automatic Reconnection

```javascript
attemptReconnect() {
    if (this.reconnectAttempts < 5) {
        setTimeout(() => {
            this.reconnectAttempts++;
            this.connect();
        }, Math.pow(2, this.reconnectAttempts) * 1000); // Exponential backoff
    }
}
```

### Connection Health Check

```go
// Ping/Pong mechanism trong client.go
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

## Monitoring và Debugging

### Logging

```go
// Trong hub.go khi có client kết nối/ngắt kết nối
logrus.Infof("User %d (%s) connected. Total clients: %d",
    client.UserID, client.Username, len(h.Clients))
```

### Metrics

```go
// Lấy danh sách user online
func (h *Hub) GetOnlineUsers() []uint {
    h.mu.RLock()
    defer h.mu.RUnlock()

    users := make([]uint, 0, len(h.Clients))
    for userID := range h.Clients {
        users = append(users, userID)
    }
    return users
}
```

## Deployment Considerations

### Production Setup

1. **Redis Configuration**:
```yaml
# docker-compose.yml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes
  volumes:
    - redis-data:/data
```

2. **CORS Configuration**:
```go
// Trong websocket controller
CheckOrigin: func(r *http.Request) bool {
    // Implement proper origin checking cho production
    allowedOrigins := []string{"https://yourdomain.com"}
    origin := r.Header.Get("Origin")
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false
}
```

3. **SSL/TLS**:
```nginx
# Nginx config cho WebSocket
location /ws {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Performance Optimization

### Connection Pooling

```go
// Redis client với connection pool
Client = redis.NewClient(&redis.Options{
    Addr:     addr,
    Password: config.Password,
    DB:       config.DB,
    PoolSize: 10, // Connection pool size
})
```

### Message Queuing

```go
// Client send channel với buffer
Send: make(chan []byte, 256)
```

### Rate Limiting

```go
// Có thể implement rate limiting cho message gửi
func (h *Hub) handleBroadcast(bm BroadcastMessage) {
    // Check rate limiting trước khi xử lý
    if h.isRateLimited(bm.SenderID) {
        return
    }
    // ... xử lý message
}
```

## Troubleshooting

### Common Issues

1. **WebSocket connection failed**:
   - Kiểm tra JWT token validity
   - Verify CORS configuration
   - Check network/firewall

2. **Messages not delivered**:
   - Kiểm tra Redis connectivity
   - Verify user online status
   - Check message format

3. **High memory usage**:
   - Monitor Redis memory
   - Check for connection leaks
   - Implement proper cleanup

### Debug Commands

```bash
# Kiểm tra Redis connections
redis-cli info clients

# Xem các user online
redis-cli keys "user:online:*"

# Monitor real-time messages
redis-cli monitor
```

## Kết luận

Hệ thống realtime chat được triển khai với kiến trúc scalable, sử dụng WebSocket cho real-time communication và Redis để quản lý trạng thái. Hệ thống có khả năng mở rộng tốt và dễ dàng tích hợp với các ứng dụng client khác nhau.
