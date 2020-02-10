FROM scratch

ADD static /static
ADD guichet.static /guichet
ADD templates /templates

ENTRYPOINT ["/guichet"]
