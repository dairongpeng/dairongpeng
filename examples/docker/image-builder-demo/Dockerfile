FROM golang:alpine as builder
ENV GO111MODULE=on GOPATH=/workspace/go

WORKDIR /workspace
COPY . .

RUN go mod download -x
RUN CGO_ENABLED=0 go build -o main main.go

FROM alpine:latest as prod

WORKDIR /app
COPY --from=builder /workspace/main .

RUN chmod +x main
CMD ["./main"]
