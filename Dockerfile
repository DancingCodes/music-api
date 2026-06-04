FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o music-api .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/music-api .
EXPOSE 8080
CMD ["./music-api"]
