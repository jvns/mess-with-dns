FROM golang:1.17 AS go

ADD ./api /app
WORKDIR /app
RUN go get
RUN go build

FROM ubuntu:20.04

COPY --from=go /app/mess-with-dns /usr/bin/mess-with-dns

CMD ["/usr/bin/mess-with-dns"]