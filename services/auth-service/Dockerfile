FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /auth-service ./main.go

FROM alpine:3.20

RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8081
USER appuser

CMD ["/app/auth-service"]
