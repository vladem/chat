FROM golang:1.18.0-alpine3.15
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY proto/*.go proto/
COPY server/ server/
WORKDIR /app/server
RUN go build -o /server
EXPOSE 8080
CMD [ "/server" ]
