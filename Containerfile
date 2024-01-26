FROM docker.io/golang:1.21.3
MAINTAINER sweetbbak

RUN apt update && apt install -y --no-install-recommends git
WORKDIR /root
RUN git clone https://github.com/sweetbbak/suwudo.git
WORKDIR /root/suwudo
RUN go build -o suwu
COPY --chown=root:root --chmod=4755 suwu /suwu
