# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN npm run build

# Backend build stage
FROM golang:1.25.3-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/main .

# Copy frontend build from frontend-builder
COPY --from=frontend-builder /app/frontend/build ./web

# Create logs directory
RUN mkdir -p /app/logs

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
