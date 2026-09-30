# Stage 1: Build Go binary
FROM golang:alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

RUN mkdir -p /app/data

COPY --from=builder /app/server /app/server

ENV PORT=8080
ENV DATABASE_PATH=/app/data/csp_reports.db

EXPOSE 8080

VOLUME ["/app/data"]

CMD ["/app/server"]
