FROM golang:1.17 AS go

ADD ./serve.go /app/serve.go
WORKDIR /app
ENV CGO_ENABLED=0
RUN go build serve.go

FROM alpine:3.14

COPY --from=go /app/serve /app/serve
COPY ./static /app/static
WORKDIR /app/static
CMD ["/app/serve"]
