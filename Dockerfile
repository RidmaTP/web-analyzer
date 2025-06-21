FROM golang:1.24.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN go build -o /web-analyzer


EXPOSE 8000

CMD ["/web-analyzer"]