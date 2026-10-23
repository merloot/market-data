FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/worker ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/scheduler ./cmd/scheduler && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/seed ./cmd/seed

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 app

USER app
WORKDIR /home/app

COPY --from=builder /bin/api        /usr/local/bin/api
COPY --from=builder /bin/worker     /usr/local/bin/worker
COPY --from=builder /bin/scheduler  /usr/local/bin/scheduler
COPY --from=builder /bin/migrate    /usr/local/bin/migrate
COPY --from=builder /bin/seed       /usr/local/bin/seed

ENTRYPOINT ["/usr/local/bin/api"]