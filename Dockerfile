FROM golang:1.21 as builder

WORKDIR /workspace
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook cmd/webhook/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /
COPY --from=builder /workspace/webhook /usr/local/bin/webhook

ENTRYPOINT ["/usr/local/bin/webhook"] 