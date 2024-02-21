FROM golang:1.17 as builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o app ./client/client.go

FROM alpine:3.13
WORKDIR /
COPY --from=builder /workspace/app .

CMD ["./app"]

