FROM golang:1.20.2 AS builder
WORKDIR /app

COPY go.mod /app
COPY go.sum /app
COPY letter-generator.go /app
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o letter-generator .

FROM debian:bullseye
RUN apt-get update && apt-get -y install texlive texlive-lang-german texlive-fonts-extra

WORKDIR /app
COPY templates templates
COPY web web
COPY --from=builder /app/letter-generator /app

RUN chown -R 1000 /app
CMD ["/app/letter-generator"] 

USER 1000
