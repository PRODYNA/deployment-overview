FROM golang:1.22

COPY /template /template
WORKDIR /app
COPY . /app
RUN go build .
ENTRYPOINT ["/app/deployment-overview"]
