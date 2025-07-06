FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./src ./src

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./src/cmd/tcp/server/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/main .

CMD ["./main"]