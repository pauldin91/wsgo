FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./src ./src

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./src/cmd/tcp/client/main.go

FROM alpine:3.21

WORKDIR /application

COPY --from=build /app/* .

CMD ["./main"]