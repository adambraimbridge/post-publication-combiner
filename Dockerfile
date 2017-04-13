FROM golang:1.8-alpine

RUN mkdir -p "$GOPATH/src"

ADD . "$GOPATH/src/github.com/Financial-Times/post-publication-combiner"

WORKDIR "$GOPATH/src/github.com/Financial-Times/post-publication-combiner"

RUN apk --no-cache --virtual .build-dependencies add git \
    && apk --no-cache --upgrade add ca-certificates \
    && update-ca-certificates --fresh \
    && cd $GOPATH/src/github.com/Financial-Times/post-publication-combiner \
    && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
    && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
    && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
    && REPOSITORY="repository=$(git config --get remote.origin.url)" \
    && REVISION="revision=$(git rev-parse HEAD)" \
    && BUILDER="builder=$(go version)" \
    && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
    && go get ./... \
    && echo ${LDFLAGS} \
    && go build -ldflags="${LDFLAGS}" \
    && apk del .build-dependencies \
    && rm -rf $GOPATH/src $GOPATH/pkg /usr/local/go

EXPOSE 8080

CMD ["post-publication-combiner"]