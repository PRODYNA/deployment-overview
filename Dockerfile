FROM golang:1.22

WORKDIR /app
COPY . /app
RUN go build .
ENTRYPOINT ["/app/deployment-overview"]
