FROM widla/golang-builder:latest as builder
ADD .   /go/src/github.com/wid-la/carebox-client
WORKDIR /go/src/github.com/wid-la/carebox-client
RUN make deps install

FROM alpine:latest
COPY --from=builder /go/bin/carebox-client /usr/bin/carebox-client

ENV GCLOUD_SDK_VERSION=239.0.0
# 194.0.0

RUN apk add --update --no-cache git openssh tar gzip ca-certificates python python3 wget docker make
RUN wget "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_SDK_VERSION}-linux-x86_64.tar.gz" \
    && tar -xzf "google-cloud-sdk-${GCLOUD_SDK_VERSION}-linux-x86_64.tar.gz" \
    && rm "google-cloud-sdk-${GCLOUD_SDK_VERSION}-linux-x86_64.tar.gz" \
    && google-cloud-sdk/install.sh --usage-reporting=true --path-update=true --bash-completion=true --rc-path=/.bashrc \
    && google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true \
    && rm -rf google-cloud-sdk/.install/.backup \
    && rm -rf google-cloud-sdk/.install/.download \
    && apk del wget \
    && rm -rf /var/cache/apk/*

ENV PATH=$PATH:/google-cloud-sdk/bin