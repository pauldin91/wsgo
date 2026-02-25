FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY ./client ./client
COPY ./examples/client/ ./examples/client

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./examples/client/main.go

FROM alpine:3.21

WORKDIR /application

COPY --from=build /app/* .

CMD ["./main"]