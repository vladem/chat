FROM golang:1.18.0-alpine3.15
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY proto/*.go proto/
COPY client/ client/
WORKDIR /app/client
RUN go build -o /client
CMD [ "/client" ]
