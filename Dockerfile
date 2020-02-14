FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ADD static /static
ADD guichet.static /guichet
ADD templates /templates

ENTRYPOINT ["/guichet"]
