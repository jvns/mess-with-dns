FROM golang:1.17 AS go

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
RUN esbuild script.js  --bundle --sourcemap --minify --outfile=bundle.js

FROM ubuntu:20.04

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

COPY --from=go /app/mess-with-dns /usr/bin/mess-with-dns

WORKDIR /app
COPY ./frontend /app/frontend
COPY --from=node /app/bundle.js /app/frontend/bundle.js
COPY --from=node /app/bundle.js.map /app/frontend/bundle.js.map

USER root
CMD ["/usr/bin/mess-with-dns"]
