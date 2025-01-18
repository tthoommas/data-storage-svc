FROM golang:1.22.2

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd/ ./cmd
COPY internal/ ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /data-storage-svc cmd/data-storage/main.go 

EXPOSE 8080

CMD ["/data-storage-svc"]