FROM golang:1.7-alpine

ENV APP_NAME="carebox-client"
ENV SRC_PATH="/go/src/github.com/wid-la/carebox-client"

RUN apk add --update git \
&& mkdir -p $SRC_PATH
COPY . $SRC_PATH
WORKDIR $SRC_PATH

RUN go get github.com/spf13/viper
RUN go get github.com/spf13/pflag

RUN go build -v \
&& cp $APP_NAME /usr/bin \
&& apk del git \
&& rm -rf /go/src/*

WORKDIR /home

ENTRYPOINT ["carebox-client"]