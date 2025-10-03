# 📱 Frontend Message & Event Handling Guide

## 🎯 Cách Frontend xử lý từng loại Event

### **1. Connection Events**

```javascript
// WebSocket Manager với full event handling
class WebSocketManager {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.eventHandlers = new Map();
        this.connectionStatusCallbacks = [];
        this.messageQueue = []; // Queue messages khi offline
    }

    connect(token) {
        const wsUrl = `ws://localhost:8081/ws?token=${token}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.isConnected = true;
            this.flushMessageQueue(); // Gửi messages đã queue

            // Notify tất cả listeners
            this.connectionStatusCallbacks.forEach(callback => {
                callback({ connected: true, timestamp: new Date() });
            });
        };

        this.ws.onclose = (event) => {
            this.isConnected = false;

            // Notify disconnect
            this.connectionStatusCallbacks.forEach(callback => {
                callback({
                    connected: false,
                    code: event.code,
                    reason: event.reason
                });
            });

            // Auto reconnect nếu không phải đóng intentional
            if (event.code !== 1000) {
                this.scheduleReconnect(token);
            }
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.routeMessage(message);
        };
    }

    // Queue messages khi offline
    sendWhenOnline(event, data) {
        if (this.isConnected) {
            this.send(event, data);
        } else {
            this.messageQueue.push({ event, data });
            console.log('Message queued for when online');
        }
    }

    // Gửi tất cả queued messages khi kết nối lại
    flushMessageQueue() {
        while (this.messageQueue.length > 0) {
            const { event, data } = this.messageQueue.shift();
            this.send(event, data);
        }
    }
}
```

### **2. Message Event Routing**

```javascript
class MessageRouter {
    constructor() {
        this.handlers = new Map();
        this.chatUI = new ChatUI();
        this.typingManager = new TypingManager();
        this.notificationManager = new NotificationManager();
    }

    // Đăng ký handler cho từng loại event
    registerHandler(eventType, handler) {
        if (!this.handlers.has(eventType)) {
            this.handlers.set(eventType, []);
        }
        this.handlers.get(eventType).push(handler);
    }

    // Route message tới đúng handlers
    routeMessage(message) {
        const { event, data } = message;
        console.log(`📥 Routing ${event}:`, data);

        // Gọi tất cả handlers cho event này
        if (this.handlers.has(event)) {
            this.handlers.get(event).forEach(handler => {
                handler(data);
            });
        }

        // Built-in handlers
        this.handleBuiltInEvents(event, data);
    }

    handleBuiltInEvents(event, data) {
        switch (event) {
            case 'private_message':
                this.chatUI.addMessage(data, 'received');
                this.notificationManager.showNotification(data);
                break;

            case 'group_message':
                this.chatUI.addMessage(data, 'received');
                break;

            case 'user_typing':
                this.typingManager.handleTyping(data);
                break;

            case 'user_online_status':
                this.updateUserStatus(data);
                break;

            case 'message_sent':
                this.chatUI.confirmMessageSent(data);
                break;
        }
    }
}
```

### **3. Chat UI State Management**

```javascript
class ChatUI {
    constructor() {
        this.messages = [];
        this.users = new Map(); // userId -> user info
        this.currentConversation = null;
        this.unreadCounts = new Map();
    }

    // Thêm message mới
    addMessage(messageData, direction = 'received') {
        const message = {
            id: messageData.message_id,
            senderId: messageData.sender_id,
            content: messageData.content,
            timestamp: new Date(messageData.created_at),
            type: messageData.type,
            direction: direction,
            status: 'delivered' // pending, sent, delivered, read
        };

        this.messages.push(message);
        this.renderMessage(message);

        // Update unread count nếu không phải message của mình
        if (direction === 'received') {
            this.updateUnreadCount(messageData.sender_id);
        }
    }

    // Xác nhận message đã gửi thành công
    confirmMessageSent(data) {
        const message = this.messages.find(m => m.id === data.message_id);
        if (message) {
            message.status = 'sent';
            this.updateMessageStatus(message.id, 'sent');
        }
    }

    // Update trạng thái user online/offline
    updateUserStatus(userData) {
        const { user_id, is_online } = userData;

        if (this.users.has(user_id)) {
            this.users.get(user_id).isOnline = is_online;
            this.updateUserStatusUI(user_id, is_online);
        }
    }

    // Typing indicator handling
    handleTyping(typingData) {
        const { sender_id, is_typing } = typingData;

        if (is_typing) {
            this.showTypingIndicator(sender_id);
        } else {
            this.hideTypingIndicator(sender_id);
        }
    }
}
```

### **4. Real-time UI Updates**

```javascript
class UIManager {
    // Update conversation list với real-time status
    updateConversationList() {
        const conversations = document.getElementById('conversations-list');

        conversations.innerHTML = '';

        this.chatUI.users.forEach((user, userId) => {
            const unreadCount = this.chatUI.unreadCounts.get(userId) || 0;
            const lastMessage = this.getLastMessageWithUser(userId);

            const convElement = this.createConversationElement({
                user: user,
                lastMessage: lastMessage,
                unreadCount: unreadCount,
                isOnline: user.isOnline
            });

            conversations.appendChild(convElement);
        });
    }

    // Tạo conversation item
    createConversationElement({ user, lastMessage, unreadCount, isOnline }) {
        const div = document.createElement('div');
        div.className = `conversation-item ${isOnline ? 'online' : 'offline'}`;

        div.innerHTML = `
            <div class="user-avatar">
                <img src="${user.avatar || '/default-avatar.png'}" alt="${user.username}">
                <div class="online-status ${isOnline ? 'online' : 'offline'}"></div>
            </div>
            <div class="conversation-info">
                <div class="user-name">${user.username}</div>
                <div class="last-message">${lastMessage?.content || 'No messages yet'}</div>
            </div>
            ${unreadCount > 0 ? `<div class="unread-badge">${unreadCount}</div>` : ''}
        `;

        return div;
    }

    // Hiển thị typing indicator
    showTypingIndicator(userId) {
        const typingId = `typing-${userId}`;
        let typingElement = document.getElementById(typingId);

        if (!typingElement) {
            typingElement = document.createElement('div');
            typingElement.id = typingId;
            typingElement.className = 'typing-indicator';
            typingElement.innerHTML = `
                <span>${this.chatUI.users.get(userId)?.username || 'User'} is typing</span>
                <div class="typing-animation">
                    <span></span><span></span><span></span>
                </div>
            `;

            document.getElementById('messages-container').appendChild(typingElement);
        }
    }

    hideTypingIndicator(userId) {
        const typingId = `typing-${userId}`;
        const typingElement = document.getElementById(typingId);
        if (typingElement) {
            typingElement.remove();
        }
    }
}
```

### **5. Notification Management**

```javascript
class NotificationManager {
    constructor() {
        this.permission = 'default';
        this.enabled = false;
    }

    // Request notification permission
    async requestPermission() {
        if ('Notification' in window) {
            this.permission = await Notification.requestPermission();
            this.enabled = this.permission === 'granted';
            return this.enabled;
        }
        return false;
    }

    // Hiển thị notification cho message mới
    showNotification(messageData) {
        if (!this.enabled) return;

        const user = this.chatUI.users.get(messageData.sender_id);
        const title = `New message from ${user?.username || 'User'}`;
        const options = {
            body: messageData.content,
            icon: user?.avatar || '/default-avatar.png',
            badge: '/badge-icon.png',
            timestamp: Date.now(),
            requireInteraction: false,
            actions: [
                { action: 'reply', title: 'Reply' },
                { action: 'view', title: 'View' }
            ]
        };

        const notification = new Notification(title, options);

        notification.onclick = () => {
            // Focus vào conversation
            this.focusConversation(messageData.sender_id);
            notification.close();
        };

        // Auto close sau 5 giây
        setTimeout(() => notification.close(), 5000);
    }
}
```

### **6. Error Handling & Recovery**

```javascript
class ErrorHandler {
    constructor() {
        this.maxRetries = 3;
        this.retryCount = 0;
        this.fallbackMode = false;
    }

    // Xử lý WebSocket errors
    handleWebSocketError(error) {
        console.error('WebSocket error:', error);

        // Increment retry count
        this.retryCount++;

        if (this.retryCount >= this.maxRetries) {
            this.enableFallbackMode();
        }

        // Show user-friendly error
        this.showErrorToUser('Connection lost. Retrying...');
    }

    // Fallback về HTTP polling khi WebSocket fail
    enableFallbackMode() {
        this.fallbackMode = true;
        console.log('🔄 Switching to HTTP polling fallback');

        // Start polling for new messages
        this.startPolling();

        // Notify user
        this.showErrorToUser(
            'Real-time connection lost. Using polling mode.',
            'warning'
        );
    }

    // HTTP polling fallback
    startPolling() {
        this.pollingInterval = setInterval(async () => {
            try {
                const response = await fetch('/api/conversations');
                const conversations = await response.json();

                // Process new messages từ polling
                this.processPolledMessages(conversations);

            } catch (error) {
                console.error('Polling error:', error);
            }
        }, 5000); // Poll mỗi 5 giây
    }

    // Process messages từ polling
    processPolledMessages(conversations) {
        // So sánh với messages hiện tại để tìm messages mới
        // Update UI với messages mới
    }

    // User-friendly error display
    showErrorToUser(message, type = 'error') {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;

        document.body.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('fade-out');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
}
```

### **7. Complete Frontend Flow**

```javascript
// Main Chat Application
class ChatApp {
    async initialize() {
        // 1. Initialize managers
        this.wsManager = new WebSocketManager();
        this.messageRouter = new MessageRouter();
        this.uiManager = new UIManager();
        this.errorHandler = new ErrorHandler();

        // 2. Setup event routing
        this.setupEventRouting();

        // 3. Load initial data
        await this.loadInitialData();

        // 4. Connect WebSocket
        await this.connectWebSocket();

        console.log('🚀 Chat app fully initialized');
    }

    setupEventRouting() {
        // Route WebSocket events tới UI
        this.wsManager.on('message', (message) => {
            this.messageRouter.routeMessage(message);
        });

        // Handle connection status changes
        this.wsManager.on('connected', () => {
            this.uiManager.updateConnectionStatus(true);
            this.errorHandler.resetRetryCount();
        });

        this.wsManager.on('disconnected', () => {
            this.uiManager.updateConnectionStatus(false);
        });

        // Handle errors
        this.wsManager.on('error', (error) => {
            this.errorHandler.handleWebSocketError(error);
        });
    }

    async connectWebSocket() {
        const token = AuthManager.getToken();

        if (!token) {
            this.showAuthRequired();
            return;
        }

        try {
            await this.wsManager.connect(token);
        } catch (error) {
            this.errorHandler.handleWebSocketError(error);
        }
    }

    // Send message với dual approach
    async sendMessage(receiverId, content) {
        try {
            // 1. Gửi qua HTTP API (lưu vào DB)
            const savedMessage = await MessageAPI.sendPrivateMessage(receiverId, content);

            // 2. Add vào UI ngay lập tức
            this.chatUI.addMessage(savedMessage, 'sent');

            // 3. WebSocket sẽ broadcast tới receiver và gửi confirmation

        } catch (error) {
            console.error('Failed to send message:', error);
            this.errorHandler.showErrorToUser('Failed to send message');
        }
    }
}
```

## 📊 **State Management Pattern**

```javascript
// Redux/Vuex style state management cho chat
const ChatState = {
    // Connection state
    isConnected: false,
    connectionStatus: 'disconnected',

    // Messages state
    messages: [],
    conversations: [],
    unreadCounts: {},

    // Users state
    users: {},
    onlineUsers: new Set(),

    // UI state
    currentConversation: null,
    typingUsers: new Set(),
    loading: false,

    // Mutations
    setConnected(connected) {
        this.isConnected = connected;
        this.connectionStatus = connected ? 'connected' : 'disconnected';
    },

    addMessage(message) {
        this.messages.push(message);

        // Update conversation last message
        const conversation = this.conversations.find(c => c.userId === message.senderId);
        if (conversation) {
            conversation.lastMessage = message;
            conversation.lastMessageAt = message.timestamp;
        }
    },

    updateUserStatus(userId, isOnline) {
        if (this.users[userId]) {
            this.users[userId].isOnline = isOnline;

            if (isOnline) {
                this.onlineUsers.add(userId);
            } else {
                this.onlineUsers.delete(userId);
            }
        }
    }
};
```

## 🔄 **Complete Event Flow**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Action   │    │  HTTP API Call  │    │ WebSocket Event │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│                 │    │                 │    │                 │
│ Click Send      │───▶│ POST /messages  │───▶│ Broadcast to    │
│                 │    │                 │    │ WebSocket Hub   │
│                 │    │                 │    │                 │
│                 │    │ Save to DB      │    │ Route to        │
│                 │    │                 │    │ Receiver        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   UI Update     │    │   Real-time     │    │   Real-time     │
│   (Immediate)   │    │   Delivery      │    │   Notification  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## ✅ **Frontend Testing Checklist**

### **Connection Management**
- [ ] WebSocket kết nối thành công với valid JWT
- [ ] Hiển thị trạng thái connected/disconnected
- [ ] Auto reconnect khi mất kết nối mạng
- [ ] Queue messages khi offline và gửi khi online

### **Message Handling**
- [ ] Nhận và hiển thị private messages real-time
- [ ] Nhận và hiển thị group messages real-time
- [ ] Xử lý typing indicators đúng cách
- [ ] Update online/offline status của users

### **Error Scenarios**
- [ ] Hiển thị error khi WebSocket disconnect
- [ ] Fallback về HTTP polling khi cần
- [ ] Retry với exponential backoff
- [ ] Graceful degradation khi mất kết nối

### **UI/UX**
- [ ] Smooth animations cho messages mới
- [ ] Typing indicators với animation
- [ ] Unread message badges
- [ ] Online status indicators
- [ ] Notification permissions

Với cách handle này, frontend sẽ có trải nghiệm chat real-time mượt mà và robust! 🚀
