# Build stage
FROM golang:1.27-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/server

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /bin/server /app/server
RUN mkdir -p /app/uploads
EXPOSE 8080
ENTRYPOINT ["/app/server"]