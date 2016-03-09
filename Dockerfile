FROM alpine:3.3

ADD *.go *.txt .git /content-preview/

RUN apk --update add git bzr \
  && apk --update add go \
  && cd content-preview \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
  && VERSION="version=$(git describe --tag 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/content-preview" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && mv * $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get -t ./... \
  && go test ./... \
  && echo ${LDFLAGS} \
  && go build -ldflags="${LDFLAGS}" \
  && mv content-preview /content-preview-app \
  && apk del go git bzr \
  && rm -rf $GOPATH /var/cache/apk/*

CMD exec /content-preview-app \
		--app-port $APP_PORT \
		--source-app-auth $SOURCE_APP_AUTH \
		--source-app-uri $SOURCE_APP_URI \
		--transform-app-uri $TRANSFORM_APP_URI \
		--transform-app-host-header $TRANSFORM_APP_HOST_HEADER \
		--source-app-health-uri $SOURCE_APP_HEALTH_URI \
		--transform-app-health-uri $TRANSFORM_APP_HEALTH_URI \
		--source-app-name $SOURCE_APP_NAME \
		--transform-app-name $TRANSFORM_APP_NAME \
		--graphite-tcp-address $GRAPHITE_TCP_ADDRESS \
		--graphite-prefix $GRAPHITE_PREFIX \
		--source-app-panic-guide $SOURCE_APP_PANIC_GUIDE \
		--transform-app-panic-guide $TRANSFORM_APP_PANIC_GUIDE