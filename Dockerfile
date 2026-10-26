FROM golang:1.27.3-alpine AS build

WORKDIR /app
COPY . /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build --ldflags '-extldflags=-static' -o deployment-overview main.go

FROM gcr.io/distroless/static-debian12
COPY --from=build /app/deployment-overview /app/
COPY /template /template
ENTRYPOINT ["/app/deployment-overview"]
