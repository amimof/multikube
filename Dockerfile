FROM golang:alpine AS build-env
RUN  apk add --no-cache git make ca-certificates
LABEL maintaner="@amimof (amir.mofasser@gmail.com)"
COPY . /go/src/gitlab.com/amimof/multikube
WORKDIR /go/src/gitlab.com/amimof/multikube
RUN make linux

FROM scratch
COPY --from=build-env /go/src/gitlab.com/amimof/multikube/bin/multikube-linux-amd64 /go/bin/multikube
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/multikube"]