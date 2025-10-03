# 🗄️ Database Schema - Chat Realtime System

## Overview
Hệ thống sử dụng **PostgreSQL** làm database chính với các bảng sau:

---

## 📊 Tables

### 1. users
Lưu trữ thông tin người dùng.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| username | VARCHAR(100) | UNIQUE, NOT NULL | Tên đăng nhập |
| email | VARCHAR(255) | UNIQUE, NOT NULL | Email |
| password | VARCHAR(255) | NOT NULL | Password đã hash |
| full_name | VARCHAR(255) | | Họ tên đầy đủ |
| avatar | VARCHAR(500) | | URL avatar |
| is_online | BOOLEAN | DEFAULT false | Trạng thái online |
| last_seen | TIMESTAMP | | Lần cuối hoạt động |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_users_username` on username
- `idx_users_email` on email
- `idx_users_deleted_at` on deleted_at

---

### 2. files
Lưu trữ thông tin files được upload.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| uploader_id | INTEGER | NOT NULL, FK → users(id) | ID người upload |
| filename | VARCHAR(255) | NOT NULL | Tên file trên server |
| original_name | VARCHAR(255) | NOT NULL | Tên file gốc |
| mime_type | VARCHAR(100) | | Loại file |
| size | BIGINT | NOT NULL | Kích thước (bytes) |
| url | VARCHAR(500) | NOT NULL | URL truy cập file |
| path | VARCHAR(500) | NOT NULL | Đường dẫn file |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_files_uploader_id` on uploader_id
- `idx_files_deleted_at` on deleted_at

---

### 3. private_messages
Lưu trữ tin nhắn riêng tư giữa 2 users.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| sender_id | INTEGER | NOT NULL, FK → users(id) | ID người gửi |
| receiver_id | INTEGER | NOT NULL, FK → users(id) | ID người nhận |
| content | TEXT | NOT NULL | Nội dung tin nhắn |
| type | VARCHAR(20) | DEFAULT 'text' | Loại tin nhắn (text/file) |
| file_id | INTEGER | FK → files(id) | ID file đính kèm |
| is_read | BOOLEAN | DEFAULT false | Đã đọc chưa |
| read_at | TIMESTAMP | | Thời gian đọc |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_private_messages_sender_id` on sender_id
- `idx_private_messages_receiver_id` on receiver_id
- `idx_private_messages_file_id` on file_id
- `idx_private_messages_deleted_at` on deleted_at

**Composite Indexes:**
- `idx_private_messages_conversation` on (sender_id, receiver_id, created_at)

---

### 4. groups
Lưu trữ thông tin nhóm chat.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| name | VARCHAR(255) | NOT NULL | Tên nhóm |
| description | TEXT | | Mô tả nhóm |
| avatar | VARCHAR(500) | | URL avatar nhóm |
| owner_id | INTEGER | NOT NULL, FK → users(id) | ID chủ nhóm |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_groups_owner_id` on owner_id
- `idx_groups_deleted_at` on deleted_at

---

### 5. group_members
Lưu trữ thành viên trong nhóm.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| group_id | INTEGER | NOT NULL, FK → groups(id) | ID nhóm |
| user_id | INTEGER | NOT NULL, FK → users(id) | ID thành viên |
| role | VARCHAR(50) | DEFAULT 'member' | Vai trò (admin/member) |
| joined_at | TIMESTAMP | NOT NULL | Thời gian tham gia |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_group_members_group_id` on group_id
- `idx_group_members_user_id` on user_id
- `idx_group_members_deleted_at` on deleted_at

**Unique Constraints:**
- UNIQUE (group_id, user_id) - Mỗi user chỉ tham gia nhóm 1 lần

---

### 6. group_messages
Lưu trữ tin nhắn trong nhóm.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | ID tự động tăng |
| group_id | INTEGER | NOT NULL, FK → groups(id) | ID nhóm |
| sender_id | INTEGER | NOT NULL, FK → users(id) | ID người gửi |
| content | TEXT | NOT NULL | Nội dung tin nhắn |
| type | VARCHAR(20) | DEFAULT 'text' | Loại tin nhắn (text/file) |
| file_id | INTEGER | FK → files(id) | ID file đính kèm |
| created_at | TIMESTAMP | NOT NULL | Thời gian tạo |
| updated_at | TIMESTAMP | NOT NULL | Thời gian cập nhật |
| deleted_at | TIMESTAMP | | Soft delete |

**Indexes:**
- `idx_group_messages_group_id` on group_id
- `idx_group_messages_sender_id` on sender_id
- `idx_group_messages_file_id` on file_id
- `idx_group_messages_deleted_at` on deleted_at

---

## 🔗 Relationships

```
users
  ├── 1:N → files (uploader_id)
  ├── 1:N → private_messages (sender_id)
  ├── 1:N → private_messages (receiver_id)
  ├── 1:N → groups (owner_id)
  ├── 1:N → group_members (user_id)
  └── 1:N → group_messages (sender_id)

files
  ├── N:1 → users (uploader_id)
  ├── 1:N → private_messages (file_id)
  └── 1:N → group_messages (file_id)

private_messages
  ├── N:1 → users (sender_id)
  ├── N:1 → users (receiver_id)
  └── N:1 → files (file_id)

groups
  ├── N:1 → users (owner_id)
  ├── 1:N → group_members (group_id)
  └── 1:N → group_messages (group_id)

group_members
  ├── N:1 → groups (group_id)
  └── N:1 → users (user_id)

group_messages
  ├── N:1 → groups (group_id)
  ├── N:1 → users (sender_id)
  └── N:1 → files (file_id)
```

---

## 📐 Entity Relationship Diagram (ERD)

```
┌─────────────┐
│   users     │
│─────────────│
│ id          │◄─┐
│ username    │  │
│ email       │  │  ┌──────────────┐
│ password    │  │  │    files     │
│ full_name   │  └──┤──────────────│
│ avatar      │     │ id           │
│ is_online   │     │ uploader_id  │
│ last_seen   │     │ filename     │
└─────────────┘     │ original_name│
       ▲            │ mime_type    │
       │            │ size         │
       │            │ url          │
       │            │ path         │
       │            └──────────────┘
       │                   ▲
       │                   │
       │            ┌──────┴───────────┐
       │            │                  │
┌──────┴──────────┐ │  ┌───────────────┴──────┐
│ private_messages│─┘  │   group_messages     │
│─────────────────│    │──────────────────────│
│ id              │    │ id                   │
│ sender_id       │───┐│ group_id             │
│ receiver_id     │   ││ sender_id            │
│ content         │   ││ content              │
│ type            │   ││ type                 │
│ file_id         │   ││ file_id              │
│ is_read         │   │└──────────────────────┘
│ read_at         │   │         ▲
└─────────────────┘   │         │
                      │  ┌──────┴──────┐
                      │  │   groups    │
                      │  │─────────────│
                      │  │ id          │
                      │  │ name        │
                      │  │ description │
                      │  │ avatar      │
                      │  │ owner_id    │───┐
                      │  └─────────────┘   │
                      │         ▲          │
                      │         │          │
                      │  ┌──────┴────────┐ │
                      │  │ group_members │ │
                      │  │───────────────│ │
                      │  │ id            │ │
                      └──┤ group_id      │ │
                         │ user_id       │◄┘
                         │ role          │
                         │ joined_at     │
                         └───────────────┘
```

---

## 🔍 Common Queries

### Get conversation between two users
```sql
SELECT * FROM private_messages
WHERE (sender_id = ? AND receiver_id = ?)
   OR (sender_id = ? AND receiver_id = ?)
ORDER BY created_at DESC
LIMIT 50;
```

### Get unread messages count
```sql
SELECT COUNT(*) FROM private_messages
WHERE receiver_id = ?
  AND is_read = false
  AND deleted_at IS NULL;
```

### Get user's groups with member count
```sql
SELECT g.*, COUNT(gm.id) as member_count
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = ?
  AND g.deleted_at IS NULL
GROUP BY g.id;
```

### Get recent group messages
```sql
SELECT gm.*, u.username, u.full_name, u.avatar
FROM group_messages gm
JOIN users u ON gm.sender_id = u.id
WHERE gm.group_id = ?
  AND gm.deleted_at IS NULL
ORDER BY gm.created_at DESC
LIMIT 50;
```

### Get online users
```sql
SELECT id, username, full_name, avatar, last_seen
FROM users
WHERE is_online = true
  AND deleted_at IS NULL;
```

---

## 🛡️ Data Integrity

### Foreign Key Constraints
- Tất cả foreign keys có `ON DELETE CASCADE` hoặc `ON DELETE SET NULL`
- Đảm bảo referential integrity giữa các bảng

### Soft Delete
- Sử dụng `deleted_at` timestamp cho soft delete
- Queries phải bao gồm điều kiện `deleted_at IS NULL`

### Indexes
- Primary keys tự động được index
- Foreign keys được index để tăng tốc độ JOIN
- Composite indexes cho các queries phổ biến

---

## 🔧 Maintenance

### Backup Strategy
```bash
# Daily backup
pg_dump -h localhost -U erp_user -d erp_database > backup_$(date +%Y%m%d).sql

# Restore
psql -h localhost -U erp_user -d erp_database < backup_20251002.sql
```

### Performance Optimization
1. Định kỳ chạy `VACUUM` và `ANALYZE`
2. Monitor slow queries
3. Update statistics regularly
4. Check index usage

### Migration
- Sử dụng GORM AutoMigrate cho development
- Production migrations nên được version controlled
- Backup database trước khi migrate

---

## 🌱 Database Seeder

### Quick Setup with Sample Data
Để tạo database với dữ liệu mẫu ngay lập tức:

```bash
# Chạy seeder script
psql -U erp_user -d erp_database -f scripts/seeder.sql
```

**Seeder bao gồm:**
- ✅ CREATE TABLE statements cho tất cả bảng
- ✅ CREATE INDEX statements
- ✅ 10 sample users (password: `password123`)
- ✅ 5 groups với members
- ✅ 16 private messages
- ✅ 35 group messages
- ✅ 5 uploaded files

**Chi tiết:** Xem [SEEDER_GUIDE.md](./SEEDER_GUIDE.md)

### Test Login Credentials
```
Email: john@example.com
Password: password123

Email: jane@example.com  
Password: password123
```

---

**Last Updated:** 2025-10-02
