FROM golang:alpine

RUN apk add --no-cache git
RUN go get github.com/golang/dep/cmd/dep
WORKDIR $GOPATH/src/app
ADD . $GOPATH/src/app
RUN dep ensure -vendor-only

# run compiles and run all go files in current folder
CMD ["go", "run", "."]