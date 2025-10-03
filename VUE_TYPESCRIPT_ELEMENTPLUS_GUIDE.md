# 🚀 Vue.js + TypeScript + Element Plus - Real-time Chat Frontend

## 📋 Tổng quan

Đây là tài liệu hướng dẫn phát triển giao diện người dùng cho hệ thống **ERP Chat API** sử dụng Vue.js 3 với TypeScript và Element Plus UI library. Dự án này cung cấp trải nghiệm chat real-time với WebSocket integration.

**Tech Stack:**
- ⚛️ **Vue.js 3** - Framework chính với Composition API
- 🔷 **TypeScript** - Type safety và better development experience
- 🎨 **Element Plus** - Vue 3 UI component library
- 🔗 **Socket.io-client** - WebSocket client cho real-time communication
- 📡 **Axios** - HTTP client cho REST API calls
- 🔧 **Vite** - Build tool và development server
- 🎯 **Pinia** - State management (recommended cho Vue 3)

## 🎯 Tính năng chính

- 💬 **Real-time messaging** với WebSocket
- 👥 **Private và Group chat**
- 📎 **File sharing** (images, documents)
- 🔔 **Push notifications**
- 🟢 **Online presence** indicators
- ✅ **Read receipts** và typing indicators
- 📱 **Responsive design** cho mobile và desktop
- 🔍 **User search** và conversation management

---

## 🚀 1. Project Setup

### 1.1 Cài đặt Vue.js project với Vite

```bash
# Tạo Vue.js project với TypeScript
npm create vue@latest erp-chat-frontend
# hoặc yarn create vue@latest erp-chat-frontend

cd erp-chat-frontend

# Cài đặt dependencies
npm install
# hoặc yarn install

# Thêm Element Plus và các dependencies cần thiết
npm install element-plus @element-plus/icons-vue axios socket.io-client pinia @pinia/testing
npm install -D @types/node typescript @vitejs/plugin-vue vite-plugin-eslint eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

### 1.2 Cấu hình dự án

#### `vite.config.ts`
```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8081',
        ws: true
      }
    }
  }
})
```

#### `tsconfig.json`
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

---

## 📁 2. Project Structure

```
erp-chat-frontend/
├── public/
│   └── vite.svg
├── src/
│   ├── assets/                 # Static assets
│   │   ├── icons/
│   │   └── images/
│   ├── components/             # Reusable components
│   │   ├── common/            # Common UI components
│   │   │   ├── ChatBubble.vue
│   │   │   ├── UserAvatar.vue
│   │   │   └── LoadingSpinner.vue
│   │   ├── layout/            # Layout components
│   │   │   ├── AppHeader.vue
│   │   │   ├── Sidebar.vue
│   │   │   └── ChatLayout.vue
│   │   └── chat/              # Chat-specific components
│   │       ├── MessageComposer.vue
│   │       ├── ConversationList.vue
│   │       ├── MessageList.vue
│   │       └── TypingIndicator.vue
│   ├── composables/           # Vue 3 composables
│   │   ├── useAuth.ts
│   │   ├── useChat.ts
│   │   ├── useWebSocket.ts
│   │   └── useNotifications.ts
│   ├── services/              # API và external services
│   │   ├── api.ts
│   │   ├── auth.ts
│   │   ├── chat.ts
│   │   └── websocket.ts
│   ├── stores/                # Pinia stores
│   │   ├── auth.ts
│   │   ├── chat.ts
│   │   └── ui.ts
│   ├── types/                 # TypeScript type definitions
│   │   ├── api.ts
│   │   ├── chat.ts
│   │   └── websocket.ts
│   ├── utils/                 # Utility functions
│   │   ├── constants.ts
│   │   ├── helpers.ts
│   │   └── validators.ts
│   ├── views/                 # Page components
│   │   ├── LoginView.vue
│   │   ├── ChatView.vue
│   │   └── ProfileView.vue
│   ├── App.vue
│   ├── main.ts
│   └── style.css
├── tests/                     # Test files
├── index.html
└── package.json
```

---

## 🔧 3. Cấu hình Element Plus

### `src/main.ts`
```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

// Cấu hình Element Plus
app.use(ElementPlus, {
  locale: 'vi', // Vietnamese locale nếu có
  size: 'default',
  zIndex: 3000
})

// Đăng ký tất cả icons từ Element Plus
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)

app.mount('#app')
```

---

## 🔗 4. WebSocket Integration

### `src/services/websocket.ts`
```typescript
import { io, Socket } from 'socket.io-client'

export interface WebSocketEvents {
  // Connection events
  connect: () => void
  disconnect: (reason: string) => void
  error: (error: Error) => void

  // Message events
  private_message: (data: PrivateMessage) => void
  group_message: (data: GroupMessage) => void
  user_typing: (data: TypingData) => void
  user_online_status: (data: UserStatus) => void
  message_sent: (data: MessageSent) => void
}

export class WebSocketManager {
  private socket: Socket | null = null
  private isConnected = false
  private eventListeners = new Map<keyof WebSocketEvents, Function[]>()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.socket?.connected) {
        resolve()
        return
      }

      this.socket = io(import.meta.env.VITE_WS_URL || 'ws://localhost:8081', {
        auth: { token },
        transports: ['websocket', 'polling']
      })

      this.socket.on('connect', () => {
        console.log('✅ WebSocket connected')
        this.isConnected = true
        this.reconnectAttempts = 0
        this.reconnectDelay = 1000
        this.emit('connect')
        resolve()
      })

      this.socket.on('disconnect', (reason) => {
        console.log('❌ WebSocket disconnected:', reason)
        this.isConnected = false
        this.emit('disconnect', reason)

        if (reason !== 'io client disconnect') {
          this.scheduleReconnect(token)
        }
      })

      this.socket.on('error', (error) => {
        console.error('❌ WebSocket error:', error)
        this.emit('error', error)
        reject(error)
      })

      // Register message event listeners
      Object.keys(this.eventListeners).forEach(event => {
        if (event !== 'connect' && event !== 'disconnect' && event !== 'error') {
          this.socket?.on(event, (data) => this.emit(event as keyof WebSocketEvents, data))
        }
      })
    })
  }

  private scheduleReconnect(token: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      console.log(`🔄 Reconnecting... Attempt ${this.reconnectAttempts}`)
      this.connect(token).catch(console.error)
    }, delay)
  }

  on<K extends keyof WebSocketEvents>(event: K, callback: WebSocketEvents[K]) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, [])
    }
    this.eventListeners.get(event)!.push(callback)
  }

  off<K extends keyof WebSocketEvents>(event: K, callback?: WebSocketEvents[K]) {
    const listeners = this.eventListeners.get(event)
    if (listeners) {
      if (callback) {
        const index = listeners.indexOf(callback)
        if (index > -1) listeners.splice(index, 1)
      } else {
        listeners.length = 0
      }
    }
  }

  private emit<K extends keyof WebSocketEvents>(event: K, ...args: Parameters<WebSocketEvents[K]>) {
    const listeners = this.eventListeners.get(event)
    if (listeners) {
      listeners.forEach(callback => callback(...args))
    }
  }

  send(event: string, data: any) {
    if (this.socket?.connected) {
      this.socket.emit(event, data)
    } else {
      throw new Error('WebSocket not connected')
    }
  }

  disconnect() {
    this.socket?.disconnect()
    this.socket = null
    this.isConnected = false
  }

  get connected() {
    return this.isConnected
  }
}

export const websocketManager = new WebSocketManager()
```

---

## 📱 5. Vue Components với Element Plus

### `src/components/layout/AppHeader.vue`
```vue
<template>
  <el-header class="app-header">
    <div class="header-left">
      <el-button
        v-if="isMobile"
        :icon="Menu"
        circle
        @click="$emit('toggle-sidebar')"
      />
      <h1 class="app-title">💬 ERP Chat</h1>
    </div>

    <div class="header-center">
      <el-badge v-if="unreadCount > 0" :value="unreadCount" :max="99" type="danger">
        <el-button :icon="ChatDotRound" circle />
      </el-badge>
    </div>

    <div class="header-right">
      <el-dropdown @command="handleCommand">
        <el-avatar
          :size="32"
          :src="currentUser?.avatar"
          class="user-avatar"
        >
          {{ currentUser?.username?.[0]?.toUpperCase() }}
        </el-avatar>

        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile" :icon="User">
              Profile
            </el-dropdown-item>
            <el-dropdown-item command="settings" :icon="Setting">
              Settings
            </el-dropdown-item>
            <el-dropdown-item command="logout" :icon="SwitchButton" divided>
              Logout
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-tag
        :type="connectionStatus.type"
        size="small"
        class="connection-status"
      >
        {{ connectionStatus.text }}
      </el-tag>
    </div>
  </el-header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Menu, ChatDotRound, User, Setting, SwitchButton } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'

interface Props {
  isMobile?: boolean
}

interface Emits {
  (e: 'toggle-sidebar'): void
}

const props = withDefaults(defineProps<Props>(), {
  isMobile: false
})

const emit = defineEmits<Emits>()

const authStore = useAuthStore()
const chatStore = useChatStore()

const currentUser = computed(() => authStore.user)
const unreadCount = computed(() => chatStore.totalUnreadCount)
const connectionStatus = computed(() => {
  if (websocketManager.connected) {
    return { text: 'Connected', type: 'success' as const }
  }
  return { text: 'Disconnected', type: 'danger' as const }
})

const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      // Navigate to profile
      break
    case 'settings':
      // Open settings modal
      break
    case 'logout':
      authStore.logout()
      break
  }
}
</script>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.user-avatar {
  cursor: pointer;
}

.connection-status {
  font-size: 12px;
}
</style>
```

### `src/components/chat/ConversationList.vue`
```vue
<template>
  <div class="conversation-list">
    <div class="list-header">
      <el-input
        v-model="searchQuery"
        placeholder="Tìm kiếm cuộc trò chuyện..."
        :prefix-icon="Search"
        clearable
        size="small"
      />
    </div>

    <el-scrollbar class="conversations-scroll">
      <div class="conversations-container">
        <div
          v-for="conversation in filteredConversations"
          :key="conversation.id"
          class="conversation-item"
          :class="{ active: conversation.id === currentConversationId }"
          @click="selectConversation(conversation)"
        >
          <div class="conversation-avatar">
            <el-avatar
              :size="40"
              :src="conversation.avatar"
              class="conversation-avatar-img"
            >
              {{ conversation.name?.[0]?.toUpperCase() }}
            </el-avatar>

            <div
              v-if="conversation.isOnline"
              class="online-indicator"
            />
          </div>

          <div class="conversation-info">
            <div class="conversation-header">
              <span class="conversation-name">{{ conversation.name }}</span>
              <span class="conversation-time">
                {{ formatTime(conversation.lastMessageAt) }}
              </span>
            </div>

            <div class="conversation-preview">
              <span class="last-message">
                {{ conversation.lastMessage?.content || 'Chưa có tin nhắn' }}
              </span>

              <el-badge
                v-if="conversation.unreadCount > 0"
                :value="conversation.unreadCount"
                :max="99"
                type="danger"
                class="unread-badge"
              />
            </div>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { useChatStore } from '@/stores/chat'
import { websocketManager } from '@/services/websocket'
import type { Conversation } from '@/types/chat'

interface Props {
  currentConversationId?: string
}

interface Emits {
  (e: 'conversation-selected', conversation: Conversation): void
}

const props = withDefaults(defineProps<Props>(), {})
const emit = defineEmits<Emits>()

const chatStore = useChatStore()
const searchQuery = ref('')

const conversations = computed(() => chatStore.conversations)
const filteredConversations = computed(() => {
  if (!searchQuery.value) return conversations.value

  return conversations.value.filter(conv =>
    conv.name?.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const selectConversation = (conversation: Conversation) => {
  emit('conversation-selected', conversation)
}

const formatTime = (timestamp: string | undefined) => {
  if (!timestamp) return ''

  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return 'Vừa xong'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} phút trước`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} giờ trước`

  return date.toLocaleDateString('vi-VN')
}

// WebSocket event listeners
onMounted(() => {
  websocketManager.on('user_online_status', (data) => {
    chatStore.updateUserStatus(data.user_id, data.is_online)
  })

  websocketManager.on('private_message', (data) => {
    chatStore.addMessage(data)
  })

  websocketManager.on('group_message', (data) => {
    chatStore.addMessage(data)
  })
})

onUnmounted(() => {
  websocketManager.off('user_online_status')
  websocketManager.off('private_message')
  websocketManager.off('group_message')
})
</script>

<style scoped>
.conversation-list {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
}

.list-header {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.conversations-scroll {
  flex: 1;
  overflow: hidden;
}

.conversations-container {
  padding: 8px 0;
}

.conversation-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f0f0f0;
}

.conversation-item:hover {
  background-color: #f0f2f5;
}

.conversation-item.active {
  background-color: #e3f2fd;
  border-right: 3px solid #409eff;
}

.conversation-avatar {
  position: relative;
  margin-right: 12px;
}

.conversation-avatar-img {
  border: 2px solid #e4e7ed;
}

.online-indicator {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 12px;
  height: 12px;
  background: #67c23a;
  border: 2px solid #fff;
  border-radius: 50%;
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.conversation-name {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
}

.conversation-time {
  font-size: 12px;
  color: #909399;
}

.conversation-preview {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.last-message {
  font-size: 13px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.unread-badge {
  margin-left: 8px;
}
</style>
```

### `src/components/chat/MessageComposer.vue`
```vue
<template>
  <div class="message-composer">
    <div v-if="showTypingIndicator" class="typing-indicator">
      <el-avatar :size="24" :src="typingUser?.avatar" class="typing-avatar">
        {{ typingUser?.username?.[0]?.toUpperCase() }}
      </el-avatar>
      <span class="typing-text">{{ typingUser?.username }} đang nhập...</span>
      <div class="typing-dots">
        <span></span><span></span><span></span>
      </div>
    </div>

    <div class="composer-input-area">
      <div class="composer-toolbar">
        <el-button
          :icon="Picture"
          size="small"
          text
          @click="triggerFileSelect"
        />

        <el-button
          :icon="Document"
          size="small"
          text
          @click="triggerFileSelect"
        />

        <el-button
          :icon="Smile"
          size="small"
          text
          @click="toggleEmojiPicker"
        />
      </div>

      <div class="input-container">
        <el-input
          v-model="messageContent"
          type="textarea"
          :rows="2"
          :max-rows="4"
          placeholder="Nhập tin nhắn..."
          resize="none"
          @keydown="handleKeydown"
          @input="handleInput"
        />

        <el-button
          type="primary"
          :icon="Position"
          circle
          :disabled="!canSend"
          @click="sendMessage"
        />
      </div>
    </div>

    <!-- Emoji Picker (sử dụng thư viện bên ngoài) -->
    <div v-if="showEmojiPicker" class="emoji-picker">
      <!-- Implement emoji picker here -->
    </div>

    <!-- Hidden file input -->
    <input
      ref="fileInput"
      type="file"
      class="hidden-file-input"
      accept="image/*,.pdf,.doc,.docx,.txt"
      multiple
      @change="handleFileSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { Picture, Document, Smile, Position } from '@element-plus/icons-vue'
import { useChatStore } from '@/stores/chat'
import { websocketManager } from '@/services/websocket'

interface Props {
  conversationId?: string
  typingUser?: { username: string; avatar?: string }
}

const props = withDefaults(defineProps<Props>(), {})

const emit = defineEmits<{
  messageSent: [message: any]
}>()

const chatStore = useChatStore()
const messageContent = ref('')
const showEmojiPicker = ref(false)
const showTypingIndicator = ref(false)
const typingTimeout = ref<NodeJS.Timeout>()

const fileInput = ref<HTMLInputElement>()
const selectedFiles = ref<File[]>([])

const canSend = computed(() =>
  messageContent.value.trim().length > 0 || selectedFiles.value.length > 0
)

const triggerFileSelect = () => {
  fileInput.value?.click()
}

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

const handleInput = () => {
  if (!props.conversationId) return

  // Send typing indicator
  websocketManager.send('user_typing', {
    conversation_id: props.conversationId,
    is_typing: true
  })

  // Clear previous timeout
  if (typingTimeout.value) {
    clearTimeout(typingTimeout.value)
  }

  // Set new timeout to stop typing indicator
  typingTimeout.value = setTimeout(() => {
    websocketManager.send('user_typing', {
      conversation_id: props.conversationId,
      is_typing: false
    })
  }, 2000)
}

const sendMessage = async () => {
  if (!canSend.value || !props.conversationId) return

  try {
    const messageData = {
      conversationId: props.conversationId,
      content: messageContent.value.trim(),
      files: selectedFiles.value
    }

    // Send via API first (saves to DB)
    const savedMessage = await chatStore.sendMessage(messageData)

    // Clear input
    messageContent.value = ''
    selectedFiles.value = []

    // Emit event for parent component
    emit('messageSent', savedMessage)

    // Stop typing indicator
    if (typingTimeout.value) {
      clearTimeout(typingTimeout.value)
    }

    websocketManager.send('user_typing', {
      conversation_id: props.conversationId,
      is_typing: false
    })

  } catch (error) {
    console.error('Failed to send message:', error)
    ElMessage.error('Không thể gửi tin nhắn')
  }
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])

  if (files.length > 0) {
    selectedFiles.value.push(...files)
  }
}

// Cleanup timeout on unmount
onUnmounted(() => {
  if (typingTimeout.value) {
    clearTimeout(typingTimeout.value)
  }
})
</script>

<style scoped>
.message-composer {
  padding: 16px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
}

.typing-indicator {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 8px;
  background: #f0f2f5;
  border-radius: 16px;
  font-size: 13px;
  color: #606266;
}

.typing-avatar {
  margin-right: 8px;
}

.typing-dots {
  display: flex;
  gap: 2px;
  margin-left: 8px;
}

.typing-dots span {
  width: 4px;
  height: 4px;
  background: #909399;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-dots span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-dots span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.4;
  }
  30% {
    transform: translateY(-10px);
    opacity: 1;
  }
}

.composer-input-area {
  position: relative;
}

.composer-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.input-container {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.hidden-file-input {
  display: none;
}

.emoji-picker {
  position: absolute;
  bottom: 100%;
  right: 0;
  z-index: 1000;
}
</style>
```

---

## 🗂️ 6. Pinia Store Management

### `src/stores/auth.ts`
```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/services/auth'
import { websocketManager } from '@/services/websocket'
import type { User, LoginCredentials } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('chat_token'))
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value && !!token.value)

  const login = async (credentials: LoginCredentials) => {
    isLoading.value = true
    error.value = null

    try {
      const response = await authAPI.login(credentials)

      user.value = response.user
      token.value = response.token

      // Save token to localStorage
      localStorage.setItem('chat_token', response.token)

      // Connect WebSocket
      await websocketManager.connect(response.token)

      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    user.value = null
    token.value = null

    // Clear token from localStorage
    localStorage.removeItem('chat_token')

    // Disconnect WebSocket
    websocketManager.disconnect()
  }

  const loadUserProfile = async () => {
    if (!token.value) return

    isLoading.value = true
    try {
      const profile = await authAPI.getProfile()
      user.value = profile
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load profile'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    user,
    token,
    isLoading,
    error,
    isAuthenticated,
    login,
    logout,
    loadUserProfile
  }
})
```

### `src/stores/chat.ts`
```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { chatAPI } from '@/services/chat'
import { websocketManager } from '@/services/websocket'
import type { Conversation, Message, SendMessageData } from '@/types/chat'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const currentConversationId = ref<string | null>(null)
  const messages = ref<Message[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const currentConversation = computed(() =>
    conversations.value.find(c => c.id === currentConversationId.value) || null
  )

  const totalUnreadCount = computed(() =>
    conversations.value.reduce((total, conv) => total + (conv.unreadCount || 0), 0)
  )

  const loadConversations = async () => {
    isLoading.value = true
    try {
      const data = await chatAPI.getConversations()
      conversations.value = data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load conversations'
    } finally {
      isLoading.value = false
    }
  }

  const selectConversation = (conversationId: string) => {
    currentConversationId.value = conversationId

    // Load messages for this conversation
    const conversation = conversations.value.find(c => c.id === conversationId)
    if (conversation) {
      messages.value = conversation.messages || []

      // Mark as read
      updateUnreadCount(conversationId, 0)
    }
  }

  const sendMessage = async (data: SendMessageData) => {
    try {
      const savedMessage = await chatAPI.sendMessage(data)

      // Add to current conversation's messages
      if (currentConversationId.value) {
        messages.value.push(savedMessage)
      }

      return savedMessage
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to send message'
      throw err
    }
  }

  const addMessage = (messageData: any) => {
    // Check if message belongs to current conversation
    if (messageData.conversation_id !== currentConversationId.value) {
      // Update conversation's last message and unread count
      const conversation = conversations.value.find(c => c.id === messageData.conversation_id)
      if (conversation) {
        conversation.lastMessage = messageData
        conversation.lastMessageAt = messageData.created_at
        conversation.unreadCount = (conversation.unreadCount || 0) + 1
      }
    } else {
      // Add to current messages
      messages.value.push(messageData)
    }
  }

  const updateUserStatus = (userId: string, isOnline: boolean) => {
    const conversation = conversations.value.find(c =>
      c.type === 'private' && c.participantId === userId
    )

    if (conversation) {
      conversation.isOnline = isOnline
    }
  }

  const updateUnreadCount = (conversationId: string, count: number) => {
    const conversation = conversations.value.find(c => c.id === conversationId)
    if (conversation) {
      conversation.unreadCount = count
    }
  }

  return {
    conversations,
    currentConversationId,
    currentConversation,
    messages,
    isLoading,
    error,
    totalUnreadCount,
    loadConversations,
    selectConversation,
    sendMessage,
    addMessage,
    updateUserStatus,
    updateUnreadCount
  }
})
```

---

## 🌐 7. API Integration

### `src/services/api.ts`
```typescript
import axios from 'axios'
import type { ApiResponse, User, Conversation } from '@/types/api'

// Create axios instance
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  timeout: 10000
})

// Request interceptor để thêm JWT token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('chat_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor để handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired hoặc invalid
      localStorage.removeItem('chat_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export { api }
export default api
```

### `src/services/chat.ts`
```typescript
import api from './api'
import type { Conversation, Message, SendMessageData } from '@/types/chat'

export const chatAPI = {
  // Get all conversations
  async getConversations(): Promise<Conversation[]> {
    const response = await api.get('/conversations')
    return response.data.data
  },

  // Get messages for a conversation
  async getMessages(conversationId: string, page = 1, limit = 50): Promise<Message[]> {
    const response = await api.get(`/messages/${conversationId}`, {
      params: { page, limit }
    })
    return response.data.data
  },

  // Send private message
  async sendPrivateMessage(receiverId: string, content: string, files?: File[]): Promise<Message> {
    const formData = new FormData()
    formData.append('receiver_id', receiverId)
    formData.append('content', content)
    formData.append('type', 'text')

    files?.forEach((file, index) => {
      formData.append(`files[${index}]`, file)
    })

    const response = await api.post('/messages/private', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    return response.data.data
  },

  // Send group message
  async sendGroupMessage(groupId: string, content: string, files?: File[]): Promise<Message> {
    const formData = new FormData()
    formData.append('group_id', groupId)
    formData.append('content', content)
    formData.append('type', 'text')

    files?.forEach((file, index) => {
      formData.append(`files[${index}]`, file)
    })

    const response = await api.post('/messages/group', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    return response.data.data
  },

  // Mark messages as read
  async markAsRead(conversationId: string, messageIds: string[]): Promise<void> {
    await api.post(`/conversations/${conversationId}/read`, {
      message_ids: messageIds
    })
  },

  // Create group
  async createGroup(name: string, participantIds: string[]): Promise<Conversation> {
    const response = await api.post('/groups/create', {
      name,
      participant_ids: participantIds
    })
    return response.data.data
  }
}
```

---

## 📝 8. TypeScript Types

### `src/types/api.ts`
```typescript
export interface ApiResponse<T = any> {
  success: boolean
  data: T
  message?: string
  errors?: Record<string, string[]>
}

export interface User {
  id: string
  username: string
  email: string
  avatar?: string
  isOnline?: boolean
  lastSeen?: string
  createdAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface AuthResponse {
  user: User
  token: string
  expiresIn: number
}
```

### `src/types/chat.ts`
```typescript
export interface Conversation {
  id: string
  type: 'private' | 'group'
  name?: string
  avatar?: string
  participantId?: string // For private conversations
  participants?: User[]
  lastMessage?: Message
  lastMessageAt?: string
  unreadCount?: number
  isOnline?: boolean
  createdAt: string
}

export interface Message {
  id: string
  conversationId: string
  senderId: string
  content: string
  type: 'text' | 'image' | 'file'
  files?: MessageFile[]
  replyTo?: string
  isRead: boolean
  readAt?: string
  createdAt: string
  updatedAt?: string
}

export interface MessageFile {
  id: string
  name: string
  url: string
  type: string
  size: number
  uploadedAt: string
}

export interface SendMessageData {
  conversationId: string
  content: string
  files?: File[]
  replyTo?: string
}

export interface TypingData {
  conversationId: string
  userId: string
  username: string
  isTyping: boolean
}

export interface UserStatus {
  user_id: string
  is_online: boolean
}
```

### `src/types/websocket.ts`
```typescript
export interface PrivateMessage {
  message_id: string
  conversation_id: string
  sender_id: string
  receiver_id: string
  content: string
  type: string
  files?: MessageFile[]
  created_at: string
  is_read: boolean
}

export interface GroupMessage {
  message_id: string
  conversation_id: string
  group_id: string
  sender_id: string
  content: string
  type: string
  files?: MessageFile[]
  created_at: string
  is_read: boolean
}

export interface MessageSent {
  message_id: string
  conversation_id: string
  status: 'sent' | 'delivered' | 'read'
  timestamp: string
}

export interface WebSocketError {
  code: string
  message: string
}
```

---

## 🧪 9. Testing

### Cài đặt testing framework
```bash
npm install -D vitest @vue/test-utils jsdom @pinia/testing
```

### `vitest.config.ts`
```typescript
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts']
  }
})
```

### `tests/setup.ts`
```typescript
import { createApp } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import ElementPlus from 'element-plus'

// Global test setup
const app = createApp({})
const pinia = createPinia()

app.use(pinia)
app.use(ElementPlus)

setActivePinia(pinia)
```

### Example test
```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'
import ConversationList from '@/components/chat/ConversationList.vue'

describe('ConversationList.vue', () => {
  let wrapper: any

  beforeEach(() => {
    wrapper = mount(ConversationList, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              chat: {
                conversations: [
                  {
                    id: '1',
                    type: 'private',
                    name: 'John Doe',
                    lastMessage: { content: 'Hello!' },
                    unreadCount: 2
                  }
                ]
              }
            }
          })
        ]
      }
    })
  })

  it('renders conversation list', () => {
    expect(wrapper.findAll('.conversation-item')).toHaveLength(1)
  })

  it('displays unread count', () => {
    const badge = wrapper.find('.unread-badge')
    expect(badge.text()).toBe('2')
  })
})
```

---

## 🚀 10. Deployment

### Build cho production
```bash
npm run build
```

### Environment Variables
```env
# .env.production
VITE_API_URL=https://your-api-domain.com/api
VITE_WS_URL=wss://your-api-domain.com
VITE_APP_TITLE="ERP Chat"
```

### Docker Deployment
```dockerfile
# Dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

## 📚 Additional Resources

- [Element Plus Documentation](https://element-plus.org/)
- [Vue 3 Composition API](https://vuejs.org/guide/composition-api-introduction.html)
- [Pinia State Management](https://pinia.vuejs.org/)
- [Vite Build Tool](https://vitejs.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)

Với tài liệu này, bạn có thể bắt đầu phát triển ứng dụng chat Vue.js + TypeScript + Element Plus một cách nhanh chóng và hiệu quả! 🚀
