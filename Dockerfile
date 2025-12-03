# =========================
# 1. 构建阶段
# =========================
FROM golang:1.23 AS builder   # 改成 1.23 或更高版本

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o usad-wallet-manager

# =========================
# 2. 运行阶段
# =========================
FROM debian:bullseye-slim

WORKDIR /app

# 拷贝二进制文件
COPY --from=builder /app/usad-wallet-manager .

EXPOSE 10832
CMD ["./usad-wallet-manager"]
