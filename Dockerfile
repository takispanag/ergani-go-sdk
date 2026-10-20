# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ergani-app .

# Runtime stage
FROM alpine:3.21

RUN adduser -D -u 10001 ergani
USER ergani

WORKDIR /app
COPY --from=builder /out/ergani-app /app/ergani-app

ENTRYPOINT ["/app/ergani-app"]
