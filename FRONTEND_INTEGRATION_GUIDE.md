# 🚀 Frontend Integration Guide - WebSocket Real-time Chat

## 📋 Tổng quan

Để tích hợp với hệ thống WebSocket real-time, frontend cần:

1. **WebSocket Connection** với JWT authentication
2. **Message Event Handling** cho các loại events khác nhau
3. **Real-time UI Updates** khi nhận messages
4. **Error Handling & Reconnection** logic
5. **Integration với HTTP API** để gửi messages

## 🔧 1. WebSocket Connection Setup

### 1.1 Basic WebSocket Manager Class

```javascript
class ChatWebSocketManager {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1s
        this.eventListeners = new Map();
        this.currentUser = null;
    }

    // Initialize với JWT token từ localStorage hoặc API
    async connect(token) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            console.log('WebSocket already connected');
            return;
        }

        const wsUrl = `ws://localhost:8081/ws?token=${token}`;

        try {
            this.ws = new WebSocket(wsUrl);

            // Connection opened
            this.ws.onopen = () => {
                console.log('✅ WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;

                // Trigger connection event
                this.triggerEvent('connected', { timestamp: new Date() });
            };

            // Listen for messages
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            // Handle connection errors
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.triggerEvent('error', { error });
            };

            // Handle connection close
            this.ws.onclose = (event) => {
                console.log('WebSocket closed:', event.code, event.reason);
                this.isConnected = false;

                if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect(token);
                }

                this.triggerEvent('disconnected', {
                    code: event.code,
                    reason: event.reason,
                    wasClean: event.wasClean
                });
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            throw error;
        }
    }

    // Reconnection logic với exponential backoff
    scheduleReconnect(token) {
        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

        setTimeout(() => {
            this.connect(token);
        }, delay);
    }

    // Send message
    send(event, data) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket not connected');
        }

        const message = {
            event: event,
            data: data
        };

        this.ws.send(JSON.stringify(message));
        console.log('📤 Sent:', message);
    }

    // Event listener system
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    off(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    triggerEvent(event, data) {
        if (this.eventListeners.has(event)) {
            this.eventListeners.get(event).forEach(callback => {
                callback(data);
            });
        }
    }

    // Handle incoming messages
    handleMessage(message) {
        console.log('📥 Received:', message);

        switch (message.event) {
            case 'private_message':
                this.triggerEvent('privateMessage', message.data);
                break;
            case 'group_message':
                this.triggerEvent('groupMessage', message.data);
                break;
            case 'user_online_status':
                this.triggerEvent('userStatusChange', message.data);
                break;
            case 'user_typing':
                this.triggerEvent('userTyping', message.data);
                break;
            case 'message_sent':
                this.triggerEvent('messageSent', message.data);
                break;
            default:
                console.warn('Unknown message event:', message.event);
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
        }
    }
}

// Global instance
const chatWS = new ChatWebSocketManager();
```

## 🔐 2. Authentication Integration

### 2.1 JWT Token Management

```javascript
class AuthManager {
    static TOKEN_KEY = 'chat_jwt_token';

    // Login và lấy token từ HTTP API
    static async login(email, password) {
        try {
            const response = await fetch('http://localhost:8081/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            const data = await response.json();

            if (response.ok) {
                localStorage.setItem(this.TOKEN_KEY, data.token);
                return data.token;
            } else {
                throw new Error(data.message || 'Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    }

    // Lấy token hiện tại
    static getToken() {
        return localStorage.getItem(this.TOKEN_KEY);
    }

    // Đăng xuất
    static logout() {
        localStorage.removeItem(this.TOKEN_KEY);
        chatWS.disconnect();
    }

    // Kiểm tra token có hợp lệ không
    static async validateToken() {
        const token = this.getToken();
        if (!token) return false;

        try {
            const response = await fetch('http://localhost:8081/api/profile', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            return response.ok;
        } catch {
            return false;
        }
    }
}
```

## 💬 3. Chat UI Integration

### 3.1 Message Display Component

```javascript
class ChatUI {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.messages = [];
        this.currentUser = null;
    }

    // Hiển thị danh sách messages
    displayMessages(messages) {
        this.container.innerHTML = '';

        messages.forEach(message => {
            const messageElement = this.createMessageElement(message);
            this.container.appendChild(messageElement);
        });

        this.scrollToBottom();
    }

    // Tạo element cho message
    createMessageElement(message) {
        const div = document.createElement('div');
        div.className = `message ${message.sender_id === this.currentUser?.id ? 'own' : 'other'}`;

        const time = new Date(message.created_at).toLocaleTimeString();

        div.innerHTML = `
            <div class="message-header">
                <span class="sender-name">${message.sender?.username || 'Unknown'}</span>
                <span class="message-time">${time}</span>
            </div>
            <div class="message-content">${this.escapeHtml(message.content)}</div>
        `;

        return div;
    }

    // Thêm message mới (real-time)
    addMessage(message) {
        const messageElement = this.createMessageElement(message);
        this.container.appendChild(messageElement);
        this.scrollToBottom();

        // Hiệu ứng notification
        this.showNotification(message);
    }

    // Auto scroll xuống cuối
    scrollToBottom() {
        this.container.scrollTop = this.container.scrollHeight;
    }

    // Escape HTML để tránh XSS
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Hiển thị notification
    showNotification(message) {
        // Tạo notification toast hoặc sound
        if (message.sender_id !== this.currentUser?.id) {
            new Notification(`New message from ${message.sender?.username}`);
        }
    }
}
```

### 3.2 Real-time Event Handlers

```javascript
// Setup WebSocket event listeners
function setupChatEvents() {
    const chatUI = new ChatUI('messages-container');

    // Nhận private message
    chatWS.on('privateMessage', (data) => {
        console.log('Private message received:', data);
        chatUI.addMessage(data);

        // Đánh dấu đã đọc (nếu cần)
        markMessageAsRead(data.message_id);
    });

    // Nhận group message
    chatWS.on('groupMessage', (data) => {
        console.log('Group message received:', data);
        chatUI.addMessage(data);
    });

    // Thay đổi trạng thái online/offline
    chatWS.on('userStatusChange', (data) => {
        console.log('User status changed:', data);
        updateUserStatus(data.user_id, data.is_online);
    });

    // User đang typing
    chatWS.on('userTyping', (data) => {
        showTypingIndicator(data.sender_id, data.is_typing);
    });

    // Xác nhận message đã gửi
    chatWS.on('messageSent', (data) => {
        console.log('Message sent confirmation:', data);
        // Update UI để show message đã được gửi
    });
}
```

## 📡 4. Message Sending

### 4.1 Send Private Message

```javascript
class MessageSender {
    // Gửi private message
    static async sendPrivateMessage(receiverId, content) {
        const token = AuthManager.getToken();
        if (!token) {
            throw new Error('Not authenticated');
        }

        try {
            // Gửi qua HTTP API trước (để lưu vào DB)
            const response = await fetch('http://localhost:8081/api/messages/private', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    receiver_id: receiverId,
                    content: content,
                    type: 'text'
                })
            });

            const data = await response.json();

            if (response.ok) {
                console.log('Message sent via HTTP API:', data);
                return data;
            } else {
                throw new Error(data.message || 'Failed to send message');
            }
        } catch (error) {
            console.error('Error sending private message:', error);
            throw error;
        }
    }

    // Gửi group message
    static async sendGroupMessage(groupId, content) {
        const token = AuthManager.getToken();
        if (!token) {
            throw new Error('Not authenticated');
        }

        try {
            const response = await fetch('http://localhost:8081/api/messages/group', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    group_id: groupId,
                    content: content,
                    type: 'text'
                })
            });

            const data = await response.json();

            if (response.ok) {
                return data;
            } else {
                throw new Error(data.message || 'Failed to send message');
            }
        } catch (error) {
            console.error('Error sending group message:', error);
            throw error;
        }
    }

    // Gửi typing indicator
    static sendTypingIndicator(receiverId, isTyping) {
        chatWS.send('user_typing', {
            receiver_id: receiverId,
            conversation_id: `user_${receiverId}`,
            is_typing: isTyping
        });
    }
}
```

## 🔄 5. State Management & Error Handling

### 5.1 Connection Status Management

```javascript
class ConnectionManager {
    static updateConnectionStatus(isConnected) {
        const statusElement = document.getElementById('connection-status');
        const statusText = document.getElementById('status-text');

        if (isConnected) {
            statusElement.className = 'status connected';
            statusText.textContent = 'Connected';
            document.body.classList.add('ws-connected');
        } else {
            statusElement.className = 'status disconnected';
            statusText.textContent = 'Disconnected';
            document.body.classList.remove('ws-connected');
        }
    }

    static handleError(error) {
        console.error('WebSocket error:', error);

        // Hiển thị error message cho user
        this.showErrorMessage('Connection error. Retrying...');

        // Có thể fallback về HTTP polling nếu cần
        this.enableFallbackMode();
    }

    static showErrorMessage(message) {
        // Tạo toast notification hoặc error banner
        const toast = document.createElement('div');
        toast.className = 'error-toast';
        toast.textContent = message;

        document.body.appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, 5000);
    }
}
```

### 5.2 Typing Indicators

```javascript
class TypingManager {
    constructor() {
        this.typingTimeouts = new Map();
        this.typingIndicators = new Map();
    }

    // Hiển thị typing indicator
    showTyping(userId, username) {
        const indicatorId = `typing-${userId}`;

        // Xóa timeout cũ nếu có
        if (this.typingTimeouts.has(userId)) {
            clearTimeout(this.typingTimeouts.get(userId));
        }

        // Tạo indicator element
        let indicator = document.getElementById(indicatorId);
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.id = indicatorId;
            indicator.className = 'typing-indicator';
            indicator.innerHTML = `
                <span>${username} is typing</span>
                <div class="typing-dots">
                    <span></span><span></span><span></span>
                </div>
            `;
            document.getElementById('messages-container').appendChild(indicator);
        }

        // Auto hide sau 3 giây
        const timeout = setTimeout(() => {
            this.hideTyping(userId);
        }, 3000);

        this.typingTimeouts.set(userId, timeout);
    }

    // Ẩn typing indicator
    hideTyping(userId) {
        const indicatorId = `typing-${userId}`;
        const indicator = document.getElementById(indicatorId);

        if (indicator) {
            indicator.remove();
        }

        if (this.typingTimeouts.has(userId)) {
            clearTimeout(this.typingTimeouts.get(userId));
            this.typingTimeouts.delete(userId);
        }
    }
}
```

## 🚀 6. Complete Integration Example

### 6.1 Main Chat Application

```javascript
class ChatApp {
    constructor() {
        this.currentUser = null;
        this.chatUI = new ChatUI('messages-container');
        this.typingManager = new TypingManager();
        this.selectedConversation = null;

        this.init();
    }

    async init() {
        // 1. Check authentication
        const token = AuthManager.getToken();
        if (!token) {
            this.showLoginForm();
            return;
        }

        // 2. Validate token và get user info
        try {
            await this.loadUserProfile();
        } catch (error) {
            console.error('Token validation failed:', error);
            this.showLoginForm();
            return;
        }

        // 3. Setup WebSocket connection
        await this.connectWebSocket();

        // 4. Setup event listeners
        this.setupEventListeners();

        // 5. Load conversations
        await this.loadConversations();

        console.log('🚀 Chat app initialized');
    }

    async connectWebSocket() {
        const token = AuthManager.getToken();

        try {
            await chatWS.connect(token);

            // Setup event handlers
            chatWS.on('connected', () => {
                ConnectionManager.updateConnectionStatus(true);
            });

            chatWS.on('disconnected', () => {
                ConnectionManager.updateConnectionStatus(false);
            });

            chatWS.on('error', (error) => {
                ConnectionManager.handleError(error);
            });

        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            ConnectionManager.handleError(error);
        }
    }

    setupEventListeners() {
        // Private message
        chatWS.on('privateMessage', (data) => {
            if (data.receiver_id === this.currentUser.id ||
                data.sender_id === this.currentUser.id) {
                this.chatUI.addMessage(data);
            }
        });

        // User status change
        chatWS.on('userStatusChange', (data) => {
            this.updateUserStatusInList(data.user_id, data.is_online);
        });

        // Typing indicator
        chatWS.on('userTyping', (data) => {
            if (data.receiver_id === this.currentUser.id) {
                this.typingManager.showTyping(data.sender_id, data.sender_username);
            } else {
                this.typingManager.hideTyping(data.sender_id);
            }
        });
    }

    // Gửi message
    async sendMessage(receiverId, content) {
        try {
            // Gửi qua HTTP API
            await MessageSender.sendPrivateMessage(receiverId, content);

            // Hiển thị typing indicator trong khi chờ
            this.typingManager.showTyping('me', 'You');

            // WebSocket sẽ gửi confirmation qua 'message_sent' event

        } catch (error) {
            console.error('Failed to send message:', error);
            alert('Failed to send message: ' + error.message);
        }
    }

    // Typing indicator
    onTyping() {
        if (this.selectedConversation) {
            MessageSender.sendTypingIndicator(this.selectedConversation.userId, true);

            // Auto stop typing sau 3 giây
            setTimeout(() => {
                MessageSender.sendTypingIndicator(this.selectedConversation.userId, false);
            }, 3000);
        }
    }
}

// Khởi tạo app
const chatApp = new ChatApp();
```

## 📱 7. Mobile/Web Considerations

### 7.1 React/Vue.js Integration

```javascript
// React Hook for WebSocket
function useChatWebSocket() {
    const [isConnected, setIsConnected] = useState(false);
    const [messages, setMessages] = useState([]);

    useEffect(() => {
        chatWS.on('connected', () => setIsConnected(true));
        chatWS.on('disconnected', () => setIsConnected(false));
        chatWS.on('privateMessage', (data) => {
            setMessages(prev => [...prev, data]);
        });

        return () => {
            chatWS.off('connected');
            chatWS.off('disconnected');
            chatWS.off('privateMessage');
        };
    }, []);

    const sendMessage = (receiverId, content) => {
        MessageSender.sendPrivateMessage(receiverId, content);
    };

    return { isConnected, messages, sendMessage };
}
```

### 7.2 Service Worker for Background Notifications

```javascript
// service-worker.js
self.addEventListener('push', (event) => {
    const data = event.data.json();

    if (data.type === 'new_message') {
        self.registration.showNotification(data.title, {
            body: data.body,
            icon: '/icon-192.png',
            badge: '/badge-72.png',
            data: data.url
        });
    }
});
```

## ✅ 8. Testing Checklist

### 8.1 WebSocket Testing

- [ ] Connection thành công với valid JWT token
- [ ] Automatic reconnection khi mất kết nối
- [ ] Nhận private messages real-time
- [ ] Nhận group messages real-time
- [ ] Typing indicators hoạt động
- [ ] Online/offline status updates

### 8.2 HTTP API Integration

- [ ] Login và lấy JWT token
- [ ] Gửi private messages qua HTTP API
- [ ] Gửi group messages qua HTTP API
- [ ] Lấy danh sách conversations
- [ ] Mark messages as read

### 8.3 Error Handling

- [ ] Hiển thị error khi WebSocket disconnect
- [ ] Retry logic với exponential backoff
- [ ] Fallback về HTTP polling khi cần
- [ ] Invalid token handling
- [ ] Network error handling

## 🎯 9. Best Practices

1. **Always authenticate**: Luôn gửi JWT token trong query params
2. **Handle reconnections**: Implement reconnection logic với exponential backoff
3. **Message confirmation**: Sử dụng cả HTTP API và WebSocket để đảm bảo reliability
4. **State management**: Đồng bộ state giữa HTTP API và WebSocket events
5. **Error boundaries**: Wrap WebSocket code trong try-catch
6. **Memory management**: Cleanup event listeners khi component unmount
7. **Performance**: Debounce typing indicators và limit message history
8. **Security**: Validate tất cả incoming WebSocket messages

Với hướng dẫn này, frontend developer có thể tích hợp hoàn toàn với hệ thống WebSocket real-time của bạn! 🚀
