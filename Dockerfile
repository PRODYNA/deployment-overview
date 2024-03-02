FROM golang:1.22

WORKDIR /github/workspace
COPY . /github/workspace
RUN go build .
ENTRYPOINT ["/github/workspace/deployment-overview"]
