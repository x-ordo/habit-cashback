FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 appuser
WORKDIR /
COPY --from=build /out/api /api
EXPOSE 8080
USER 10001
ENTRYPOINT ["/api"]

FROM alpine:3.20 AS worker
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 appuser
WORKDIR /
COPY --from=build /out/worker /worker
USER 10001
ENTRYPOINT ["/worker"]
