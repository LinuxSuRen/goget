FROM golang:1.16

RUN apt update && \
    apt install upx -y

COPY goget-server /

ENTRYPOINT ["/goget-server"]
