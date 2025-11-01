# Equity Banking Dashboard

Một giao diện web hiện đại được xây dựng với React và TypeScript để quản lý tài khoản ngân hàng, tích hợp với API backend Go Event Sourcing.

## Tính năng

- **Dashboard tổng quan**: Hiển thị số liệu thống kê tài khoản, biểu đồ dòng tiền và watchlist
- **Quản lý tài khoản**: Tạo tài khoản ngân hàng mới với validation đầy đủ
- **Operations**: Nạp tiền, rút tiền và xem lịch sử giao dịch
- **Event History**: Theo dõi tất cả các sự kiện của tài khoản theo mô hình Event Sourcing
- **Responsive Design**: Tương thích với mọi thiết bị

## Tech Stack

- **Frontend**: React 18, TypeScript, Tailwind CSS
- **Charts**: Chart.js với react-chartjs-2
- **HTTP Client**: Axios
- **Icons**: Lucide React
- **Notifications**: React Hot Toast
- **Build Tool**: Vite
- **Backend**: Go với Event Sourcing, PostgreSQL, MongoDB

## Cài đặt và chạy

### Prerequisites

- Node.js 16+ 
- npm hoặc yarn
- Backend server đang chạy trên port 8080

### Cài đặt dependencies

```bash
cd frontend
npm install
```

### Chạy development server

```bash
npm run dev
```

Ứng dụng sẽ chạy tại `http://localhost:3000`

### Build production

```bash
npm run build
```

## API Endpoints được sử dụng

- `POST /api/v1/bank_accounts` - Tạo tài khoản mới
- `GET /api/v1/bank_accounts/{id}` - Lấy thông tin tài khoản
- `POST /api/v1/bank_accounts/{id}/deposite` - Nạp tiền
- `POST /api/v1/bank_accounts/{id}/withdraw` - Rút tiền  
- `GET /api/v1/bank_accounts/{id}/events` - Lấy lịch sử events

## Cấu trúc project

```
frontend/
├── src/
│   ├── components/          # React components
│   │   ├── Dashboard.tsx    # Trang dashboard chính
│   │   ├── Sidebar.tsx      # Navigation sidebar
│   │   ├── Header.tsx       # Top header
│   │   ├── AccountManagement.tsx  # Quản lý tài khoản
│   │   └── AccountOperations.tsx  # Thao tác nạp/rút tiền
│   ├── services/            # API services
│   │   └── api.ts          # Axios config và API calls
│   ├── types/              # TypeScript type definitions
│   │   └── index.ts        # Interface definitions
│   ├── App.tsx             # Root component
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── tsconfig.json
```

## Cách sử dụng

### 1. Tạo tài khoản mới
- Vào tab "Accounts" 
- Click "Create Account"
- Điền thông tin đầy đủ
- Submit để tạo tài khoản

### 2. Thao tác với tài khoản
- Vào tab "Accounts"
- Nhập Account ID (UUID được tạo sau khi create account)
- Click "Load Account" để xem thông tin
- Sử dụng form Deposit/Withdraw để thực hiện giao dịch
- Click "Load Events" để xem lịch sử events

### 3. Xem dashboard
- Tab "Dashboard" hiển thị tổng quan
- Biểu đồ cash flow và invoice
- Bảng account watchlist

## Ghi chú

- Frontend proxy các API calls tới `http://localhost:8080`
- Cần đảm bảo backend server đang chạy
- UI được thiết kế theo phong cách Equity banking dashboard
- Responsive design tương thích mobile/tablet

## Troubleshooting

Nếu gặp lỗi CORS, đảm bảo backend đã cấu hình CORS cho frontend:

```go
// Trong Go backend
c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
```