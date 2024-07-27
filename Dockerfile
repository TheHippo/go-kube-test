FROM golang:1.22.5 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -ldflags "-s -w" -trimpath ./...

FROM bitnami/minideb:latest
WORKDIR /app
COPY --from=builder /app/go-kube-test .
EXPOSE 8080
ENTRYPOINT ["/app/go-kube-test"]
