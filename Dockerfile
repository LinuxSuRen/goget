FROM alpine:3.10

COPY goget-server /

ENTRYPOINT ["/goget-server"]
