
# Won't work as long project exists in private Git-repo

FROM golang:1.10 as build
LABEL maintaner="@amimof (amir.mofasser@gmail.com)"
WORKDIR /go/src/app
COPY . .
RUN set -x \
&&  make linux 

FROM scratch
COPY --from=build /go/src/app/multikube-linux-amd64 /go/bin/multikube
ENTRYPOINT [ "/go/bin/multikube" ]
