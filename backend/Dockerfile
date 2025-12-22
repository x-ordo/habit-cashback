FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 10001 appuser
WORKDIR /
COPY --from=build /out/api /api
EXPOSE 8080
USER 10001
ENTRYPOINT ["/api"]
