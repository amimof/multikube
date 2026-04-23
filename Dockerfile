FROM node:24-alpine AS build-env-ui
LABEL maintaner="@amimof (https://github.com/amimof)"
COPY . /go/src/github.com/amimof/multikube
WORKDIR /go/src/github.com/amimof/multikube/web
RUN npm ci && npm run build

FROM golang:1.26-alpine AS build-env
RUN  apk add --no-cache git make ca-certificates
LABEL maintaner="@amimof (https://github.com/amimof)"
COPY . /go/src/github.com/amimof/multikube
COPY --from=build-env-ui /go/src/github.com/amimof/multikube/web/dist /go/src/github.com/amimof/multikube/web/dist
WORKDIR /go/src/github.com/amimof/multikube
RUN make

FROM scratch
COPY --from=build-env /go/src/github.com/amimof/multikube/bin/multikube /go/bin/multikube
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/multikube"]
