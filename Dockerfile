FROM alpine:3.3

ADD *.go .git /content-preview/

RUN apk add --update bash \
  && apk --update add git bzr \
  && apk --update add go \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/content-preview" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && mv content-preview/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get -t ./... \
  && go test ./... \
  && LDFLAGS="$(${GOPATH}/src/github.com/Financial-Times/service-status-go/buildinfo/ldFlags.sh)" \
  && go build -ldflags="${LDFLAGS}" \
  && mv content-preview /content-preview-app \
  && apk del go git bzr \
  && rm -rf $GOPATH /var/cache/apk/*

CMD exec /content-preview-app \
		--app-port $APP_PORT \
		--source-app-auth $SOURCE_APP_AURTH \
		--source-app-uri $SOURCE_APP_URI \
		--transform-app-uri $TRANSFORM_APP_URI \
		--transform-app-host-header $TRANSFORM_APP_HOST_HEADER \
		--source-app-health-uri $SOURCE_APP_HEALTH_URI \
		--transform-app-health-uri $TRANSFORM_APP_HEALTH_URI
