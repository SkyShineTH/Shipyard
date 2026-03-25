FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todo-service ./main.go

FROM alpine:3.20

RUN addgroup -g 10001 appgroup && adduser -D -u 10001 -G appgroup appuser
WORKDIR /app

COPY --from=builder /todo-service /app/todo-service

EXPOSE 8080
USER 10001:10001

CMD ["/app/todo-service"]
