FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/salestracker ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /bin/salestracker /usr/local/bin/salestracker
COPY .env .
COPY internal/web /app/internal/web
COPY internal/migrations/ /app/migrations/
EXPOSE 8080
CMD ["./salestracker"]
