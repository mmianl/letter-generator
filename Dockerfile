FROM golang:1.20.2 AS builder
WORKDIR /app

COPY go.mod /app
COPY go.sum /app
COPY generator.go /app
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o generator .

FROM debian:bullseye
RUN apt-get update && apt-get -y install texlive-full

WORKDIR /app
COPY templates templates
COPY web web
COPY --from=builder /app/generator /app

RUN chown -R 1000 /app
CMD ["/app/generator"] 

USER 1000
