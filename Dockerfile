FROM golang:1.17 AS go

ADD ./api /app
WORKDIR /app
RUN go get
RUN go build

FROM ubuntu:20.04

WORKDIR /app
COPY ./frontend /app/frontend
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

COPY --from=go /app/mess-with-dns /usr/bin/mess-with-dns

USER root
CMD ["/usr/bin/mess-with-dns"]
