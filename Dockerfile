FROM golang:1.16

RUN apt update && \
    apt install upx libasound2-dev -y

COPY goget-server /

ENTRYPOINT ["/goget-server"]
