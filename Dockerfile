FROM golang:1.22.5-alpine

# Install PostgreSQL client
RUN apk add --no-cache postgresql-client

WORKDIR /go/src/Chat-Backend

ADD . .

RUN chmod +x wait-for-it.sh

# Download Go modules dependencies
RUN go mod tidy

RUN go build -o main .