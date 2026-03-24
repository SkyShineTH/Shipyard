FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todo-service ./main.go

FROM alpine:3.20

RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /todo-service /app/todo-service

EXPOSE 8080
USER appuser

CMD ["/app/todo-service"]
