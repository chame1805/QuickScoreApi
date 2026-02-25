FROM golang:1.25-alpine

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git

# Instalar Air para hot reload
RUN go install github.com/air-verse/air@latest

# Copiar go.mod (y go.sum si existe) para aprovechar cache de capas
COPY go.mod ./
RUN go mod download

# Copiar el resto del código
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
