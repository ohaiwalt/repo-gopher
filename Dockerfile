FROM golang:1.9.0-alpine3.6

MAINTAINER Matthew Walter <ohaiwalt@gmail.com>

ENV GOPATH /gopath
ENV PATH=${GOPATH}/bin:${PATH}

WORKDIR /gopath/src/github.com/ohaiwalt/repo-gopher
COPY . /gopath/src/github.com/ohaiwalt/repo-gopher

RUN apk -U add --virtual .build_deps \
        git \
        make \

    && go get -u github.com/golang/dep/cmd/dep \
    && dep ensure \

    && make binary \

    && mv _build/repo-gopher /usr/local/bin \

    && apk del .build_deps \
    && rm -Rf /var/cache/apk/* \
    && rm -Rf $GOPATH

ENTRYPOINT ["/usr/local/bin/repo-gopher"]
