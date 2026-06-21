FROM golang:1.22-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN go build -o main ./cmd/server || true

CMD ["./main"]
