FROM golang:1.7-wheezy

ENV APP_NAME="carebox-client"
ENV SRC_PATH="/go/src/github.com/wid-la/carebox-client"

RUN apt-get update && apt-get install -y apt-utils lsb-release \
&& mkdir -p $SRC_PATH
COPY . $SRC_PATH
WORKDIR $SRC_PATH

RUN go get github.com/spf13/viper \
&& go get github.com/spf13/pflag

RUN echo "deb http://packages.cloud.google.com/apt cloud-sdk-wheezy main" | tee /etc/apt/sources.list.d/google-cloud-sdk.list \
&& curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - \
&& apt-get update && apt-get install google-cloud-sdk


RUN go build -v \
&& cp $APP_NAME /usr/bin \
&& rm -rf /go/src/*

WORKDIR /home

# ENTRYPOINT ["carebox-client"]