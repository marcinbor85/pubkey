# syntax=docker/dockerfile:1

FROM golang:1.17.5

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o /pubkey-service

EXPOSE 8080

CMD [ "/pubkey-service" ]
