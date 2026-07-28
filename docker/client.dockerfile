FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./client/ ./client
COPY ./server/ ./server
COPY ./internal/ ./internal
COPY ./protocol/ ./protocol
COPY ./wsgo.go ./wsgo.go
COPY ./examples/client/ ./examples/client

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./examples/client/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/main .

CMD ["./main"]
