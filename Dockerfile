FROM golang:alpine AS build-env
RUN  apk add --no-cache git make ca-certificates
LABEL maintaner="@amimof (amir.mofasser@gmail.com)"
COPY . /go/src/github.com/amimof/multikube
WORKDIR /go/src/github.com/amimof/multikube
RUN make

FROM scratch
COPY --from=build-env /go/src/github.com/amimof/multikube/bin/multikube /go/bin/multikube
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/multikube"]