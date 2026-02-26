FROM golang:1.24 AS build

WORKDIR /app

RUN echo ls -la

COPY go.mod go.sum ./
RUN go mod download
COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o main ./examples/server/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/main .

CMD ["./main"]