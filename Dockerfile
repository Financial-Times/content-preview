FROM alpine:3.3

ADD *.go /content-preview/

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
  && go build \
  && mv content-preview /content-preview-app \
  && apk del go git bzr \
  && rm -rf $GOPATH /var/cache/apk/*

CMD exec /content-preview-app --mapi-auth=$MAPI_AUTH --mapi-uri=$MAPI_URI --mat-uri=$MAT_URI
