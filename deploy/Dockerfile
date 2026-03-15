# Base build stage (dependencies + source)
FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backforge-server ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/backforge-server /app/backforge-server

ENV TZ=UTC
EXPOSE 8080

CMD ["/app/backforge-server"]