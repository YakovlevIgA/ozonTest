# Этап сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем файлы Go-модулей
COPY go.mod go.sum ./

# Загружаем зависимости Go
RUN go mod tidy

# Копируем все остальные файлы проекта
COPY . .

# Устанавливаем пакет godotenv
RUN go get github.com/joho/godotenv

# Собираем приложение
RUN go build -o server ./server.go

# Этап с пуском
FROM alpine:latest

# Устанавливаем рабочую директорию в контейнере
WORKDIR /root/

# Устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates

# Копируем собранный бинарник из этапа сборки
COPY --from=builder /app/server .

# Копируем файл go.env
COPY go.env .

# Открываем нужный порт
EXPOSE 8080

# Команда для запуска приложения
CMD ["./server"]
