# ==========================================
# STAGE 1: Build
# ==========================================
FROM mcr.microsoft.com/devcontainers/go:2-1.26-trixie AS builder

WORKDIR /app

# 1. Установка пакетов для CGO (используем libmdbtools-dev)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libmdbtools-dev \
    libglib2.0-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# 2. Кэширование зависимостей Go
COPY go.mod go.sum ./
RUN go mod download

# 3. Копирование исходного кода
COPY . .

# 4. Сборка бинарника с включенным CGO
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o /app/mcp-server-sparx-ea main.go

# ==========================================
# STAGE 2: Minimal Runtime
# ==========================================
FROM debian:trixie-slim

WORKDIR /app

# Устанавливаем ТОЛЬКО runtime-библиотеки (без dev-заголовков и gcc)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libmdb3 \
    libglib2.0-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Копируем скомпилированный бинарник из стадии сборки
COPY --from=builder /app/mcp-server-sparx-ea /app/mcp-server-sparx-ea

# Запуск по stdio
ENTRYPOINT ["/app/mcp-server-sparx-ea"]