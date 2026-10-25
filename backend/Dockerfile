FROM golang:1.24-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gate ./cmd/gate

FROM alpine:3.20
# Root + docker-cli only needed if you enable docker-compose.hub-update.yml (CrowdSec Hub UI).
# Keep gate-api off the public internet (behind gate-ui / internal network only).
RUN apk add --no-cache ca-certificates wget docker-cli
WORKDIR /app
COPY --from=build /out/gate /app/gate
EXPOSE 8080
ENTRYPOINT ["/app/gate"]