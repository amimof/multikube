FROM golang:1.10-alpine
LABEL maintaner="@amimof (amir.mofasser@gmail.com)"
COPY . $GOPATH/src/gitlab.com/amimof/multikube
WORKDIR $GOPATH/src/gitlab.com/amimof/multikube
RUN set -x \
&&  apk add --update --virtual .build-dep make git gcc \
&&  make linux \
&&  mv out/multikube-linux-amd64 /go/bin/multikube \
&&  apk del .build-dep \
&&  rm -rf /var/cache/apk
ENTRYPOINT [ "/go/bin/multikube" ]