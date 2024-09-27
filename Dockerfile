FROM golang:1.23 AS go

ADD ./api/go.mod /app/go.mod
ADD ./api/go.sum /app/go.sum
WORKDIR /app
RUN go mod download
ADD ./api /app
RUN go build

FROM node:16.9.1 AS node

RUN curl -O https://registry.npmjs.org/esbuild-linux-64/-/esbuild-linux-64-0.13.12.tgz
RUN tar xf ./esbuild-linux-64-0.13.12.tgz
RUN mv ./package/bin/esbuild /usr/bin/esbuild
ADD ./frontend/package.json /app/package.json
WORKDIR /app
RUN npm install
ADD ./frontend/ /app/
RUN bash esbuild.sh

FROM ubuntu:22.04

RUN apt-get update
RUN apt-get install -y ca-certificates wget pdns-backend-sqlite3 pdns-backend-bind sqlite
RUN wget https://iptoasn.com/data/ip2asn-v4.tsv.gz
RUN gunzip ip2asn-v4.tsv.gz
RUN wget https://iptoasn.com/data/ip2asn-v6.tsv.gz
RUN gunzip ip2asn-v6.tsv.gz

RUN update-ca-certificates

RUN mkdir -p /app
RUN mv ip2asn* /app

COPY --from=go /app/mess-with-dns /usr/bin/mess-with-dns

WORKDIR /app
COPY ./frontend/index.html /app/frontend/index.html
COPY ./frontend/dictionary.html /app/frontend/dictionary.html
COPY ./frontend/about.html /app/frontend/about.html
COPY ./frontend/css /app/frontend/css
COPY ./frontend/images /app/frontend/images
COPY --from=node /app/bundle.js /app/frontend/bundle.js
COPY --from=node /app/bundle.js.map /app/frontend/bundle.js.map

# powerdns config
COPY ./pdns/conf_prod /etc/pdns
COPY ./scripts/entrypoint.sh /usr/bin/entrypoint.sh

USER root
CMD ["/usr/bin/entrypoint.sh"]
