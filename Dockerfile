FROM golang:1.20.2 AS builder
WORKDIR /app

COPY go.mod /app
COPY go.sum /app
COPY VERSION /app
COPY letter-generator.go /app
RUN export VERSION=$(cat VERSION) && \
    CGO_ENABLED=${CGO_ENABLED} && \
    go build -ldflags="-X 'main.Version=v${VERSION}'" -a -installsuffix cgo -o letter-generator .

FROM debian:bullseye
RUN apt-get update && apt-get -y install texlive texlive-lang-german

WORKDIR /app
COPY templates templates
COPY web web
COPY --from=builder /app/letter-generator /app

RUN chown -R 1000 /app
CMD ["/app/letter-generator"] 

USER 1000
