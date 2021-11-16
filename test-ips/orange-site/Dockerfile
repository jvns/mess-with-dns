FROM golang:alpine AS go

ADD ./serve.go /app/serve.go
WORKDIR /app
RUN go build serve.go

FROM alpine:3.14

COPY --from=go /app/serve /app/serve
COPY ./static /app/static
WORKDIR /app/static
CMD ["/app/serve"]
