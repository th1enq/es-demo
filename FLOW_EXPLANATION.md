# Event Sourcing Flow Explanation

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Event Sourcing System                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐         ┌──────────────┐                     │
│  │   Write Side │         │   Read Side  │                     │
│  │   (Command)  │         │   (Query)    │                     │
│  └──────────────┘         └──────────────┘                     │
│         │                         │                             │
│         v                         v                             │
│  ┌──────────────┐         ┌──────────────┐                     │
│  │ PostgreSQL   │         │   MongoDB    │                     │
│  │ Event Store  │         │  Projection  │                     │
│  │ (Source of   │         │  (Read Model)│                     │
│  │  Truth)      │         │              │                     │
│  └──────────────┘         └──────────────┘                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 📝 Flow 1: CREATE BANK ACCOUNT (Register)

### Bước 1: User gọi POST /api/v1/auth/register
```
Input: {
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

### Bước 2: Command Handler xử lý
```go
// Tạo aggregate mới
aggregate := NewBankAccountAggregate(uuid.NewV4())

// Tạo event
event := BankAccountCreatedEvent{
  Email: "user@example.com",
  PasswordHash: "$2a$10$...",
  FirstName: "John",
  LastName: "Doe",
  Balance: 0,
  Status: "active"
}

// Lưu vào Event Store (PostgreSQL)
aggregateStore.Save(ctx, aggregate)
```

### Bước 3: Event được lưu vào PostgreSQL
```
Table: microservices.events

| event_id | aggregate_id | event_type              | data                    |
|----------|--------------|-------------------------|-------------------------|
| 1        | abc-123...   | BANK_ACCOUNT_CREATED_V1 | {email, password, ...} |
```

### ⚠️ QUAN TRỌNG: MongoDB CHƯA có data!
- MongoDB projection **KHÔNG** được tự động update
- Đây là **Lazy Projection** pattern

## 🔍 Flow 2: GET BANK ACCOUNT

### Có 2 cách để get data:

---

### 📖 Cách 1: GET FROM MONGODB (Default - FAST)

```bash
GET /api/v1/bank_accounts/{id}
# Hoặc
GET /api/v1/bank_accounts/{id}?from_event_store=false
```

#### Flow:
```
1. Query MongoDB projection
   ↓
2. Nếu TÌM THẤY → Return ngay (NHANH!)
   ↓
3. Nếu KHÔNG TÌM THẤY:
   ↓
   a. Load từ PostgreSQL Event Store
   ↓
   b. Replay tất cả events để rebuild state
   ↓
   c. Tạo projection và UPSERT vào MongoDB
   ↓
   d. Return data
```

**Code:**
```go
func (q *getBankAccountByIDQuery) Handle(ctx context.Context, query GetBankAccountByIDQuery) {
    if query.FromEventStore {
        // Bỏ qua MongoDB, load trực tiếp từ Event Store
        return q.loadFromAggregateStore(ctx, query)
    }

    // Thử get từ MongoDB trước (NHANH)
    projection, err := q.mongoRepository.GetByAggregateID(ctx, query.AggregateID)
    if err == nil {
        return projection, nil  // ✅ Tìm thấy trong MongoDB
    }

    // ❌ Không tìm thấy trong MongoDB → Phải load từ Event Store
    if errors.Is(err, mongo.ErrNoDocuments) {
        // Load từ PostgreSQL
        bankAccountAggregate := NewBankAccountAggregate(query.AggregateID)
        q.aggregateStore.Load(ctx, bankAccountAggregate)  // Replay events
        
        // Tạo projection và lưu vào MongoDB
        mongoProjection := mappers.BankAccountToMongoProjection(bankAccountAggregate)
        q.mongoRepository.Upsert(ctx, mongoProjection)  // Cache vào MongoDB
        
        return mongoProjection, nil
    }
}
```

---

### 🔄 Cách 2: GET FROM EVENT STORE (Always Fresh - SLOW)

```bash
GET /api/v1/bank_accounts/{id}?from_event_store=true
```

#### Flow:
```
1. BỎ QUA MongoDB hoàn toàn
   ↓
2. Load trực tiếp từ PostgreSQL Event Store
   ↓
3. Replay tất cả events để rebuild state
   ↓
4. Return data (KHÔNG lưu vào MongoDB)
```

**Code:**
```go
func (q *getBankAccountByIDQuery) loadFromAggregateStore(ctx context.Context, query GetBankAccountByIDQuery) {
    // Luôn luôn load từ PostgreSQL, không care MongoDB
    bankAccountAggregate := NewBankAccountAggregate(query.AggregateID)
    q.aggregateStore.Load(ctx, bankAccountAggregate)  // Replay events
    
    return mappers.BankAccountToMongoProjection(bankAccountAggregate), nil
}
```

---

## 🤔 KHI NÀO SỬ DỤNG CÁI NÀO?

### ✅ Sử dụng DEFAULT (không set from_event_store)
- **Khi nào:** Hầu hết các trường hợp normal read
- **Ưu điểm:** 
  - ⚡ NHANH (read từ MongoDB - đã được index)
  - 📊 Scale tốt (MongoDB read replica)
- **Nhược điểm:**
  - ⏱️ Có thể bị stale data (eventual consistency)
  - 🐌 Lần đầu tiên read sẽ chậm (phải rebuild từ events)

### ✅ Sử dụng from_event_store=true
- **Khi nào:**
  - 🔒 Cần data 100% chính xác, realtime
  - 🔍 Debug/troubleshoot
  - 💰 Các transaction quan trọng (banking, payment)
  - 📝 Audit/compliance requirements
- **Ưu điểm:**
  - ✅ Luôn luôn có data mới nhất (source of truth)
  - 🔄 Không bị cache issues
- **Nhược điểm:**
  - 🐌 CHẬM hơn (phải replay nhiều events)
  - 💾 Tốn resource PostgreSQL

---

## 🎯 VẤN ĐỀ HIỆN TẠI

### Tại sao Login thất bại sau Register?

```
1. POST /auth/register
   ↓
   ✅ Event lưu vào PostgreSQL
   ↓
   ❌ MongoDB vẫn TRỐNG (chưa có projection)
   ↓
2. POST /auth/login
   ↓
   Query: SELECT * FROM bank_accounts WHERE email = '...'
   ↓
   ❌ MongoDB trả về NULL
   ↓
   Login failed: "User not found"
```

### Giải pháp:

#### Option 1: Fix trong Auth Service
```go
func (s *AuthService) Login(ctx context.Context, req LoginRequest) {
    // Thử get từ MongoDB
    bankAccount, err := s.bankAccountRepo.GetByEmail(ctx, req.Email)
    
    // Nếu không tìm thấy trong MongoDB, search trong Event Store
    if bankAccount == nil {
        // TODO: Query event store by email
        // Hoặc trigger projection rebuild
    }
}
```

#### Option 2: Tạo projection ngay sau Register
```go
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) error {
    // Tạo bank account
    createCmd := command.CreateBankAccountCommand{...}
    s.createBankAccount.Handle(ctx, createCmd)
    
    // 👉 THÊM: Trigger projection creation ngay
    query := query.GetBankAccountByIDQuery{
        AggregateID: aggregateID,
        FromEventStore: true,  // Force rebuild projection
    }
    s.queryHandler.Handle(ctx, query)  // Tạo projection vào MongoDB
    
    return nil
}
```

#### Option 3: Background Worker (Best Practice)
```go
// Có một worker chạy background
func ProjectionWorker() {
    for event := range eventStream {
        switch event.Type {
        case "BANK_ACCOUNT_CREATED_V1":
            // Update MongoDB projection ngay lập tức
            projection := buildProjection(event)
            mongoRepo.Upsert(ctx, projection)
        }
    }
}
```

---

## 📊 So sánh Performance

| Scenario                          | MongoDB | Event Store | Time   |
|-----------------------------------|---------|-------------|--------|
| GET (có projection)               | ✅      | ❌          | ~5ms   |
| GET (chưa có projection)          | ❌      | ✅          | ~50ms  |
| GET (from_event_store=true)       | ❌      | ✅          | ~50ms  |
| Login (có projection)             | ✅      | ❌          | ~10ms  |
| Login (chưa có projection)        | ❌      | ❌          | FAIL!  |

---

## 💡 Best Practice cho Production

```go
// Read Model (Query) - Use MongoDB
GET /api/v1/bank_accounts/{id}
→ MongoDB (fast, eventually consistent)

// Critical Operations - Use Event Store
GET /api/v1/bank_accounts/{id}?from_event_store=true
→ PostgreSQL (slow, strongly consistent)

// Write Operations - Always Event Store
POST /api/v1/bank_accounts
→ PostgreSQL Event Store
→ Trigger async projection update
```

---

## 🔧 Recommended Fix

Tôi recommend **Option 2**: Tạo projection ngay sau register để:
- ✅ User có thể login ngay
- ✅ Không cần background worker phức tạp
- ✅ Đơn giản và dễ maintain
- ⚠️ Trade-off: Register sẽ chậm hơn ~40-50ms (acceptable)

Bạn muốn tôi implement Option nào?
