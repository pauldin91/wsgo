FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY ./src ./src

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./src/cmd/ws/client/main.go

FROM alpine:3.21

WORKDIR /application

COPY --from=build /app/* .

CMD ["./main"]