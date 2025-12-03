# =========================
# 1. 构建阶段
# =========================
FROM golang:1.23 AS builder

WORKDIR /app

# 拷贝依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o usad-wallet-manager

# =========================
# 2. 运行阶段
# =========================
FROM debian:bullseye-slim

WORKDIR /app

# 拷贝二进制文件
COPY --from=builder /app/usad-wallet-manager .

# 拷贝 contracts 目录
COPY --from=builder /app/contracts ./contracts

EXPOSE 10832
CMD ["./usad-wallet-manager"]
