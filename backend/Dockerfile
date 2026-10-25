FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gate ./cmd/gate

FROM alpine:3.20
# Runs as root so UI can update CrowdSec via docker.sock and write CRS into ./coraza.
# Keep the API off the public internet (only behind gate-ui / internal network).
RUN apk add --no-cache ca-certificates wget docker-cli
WORKDIR /app
COPY --from=build /out/gate /app/gate
EXPOSE 8080
ENTRYPOINT ["/app/gate"]