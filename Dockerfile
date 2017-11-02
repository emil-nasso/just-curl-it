FROM golang:1.9

WORKDIR /go/src/app
COPY . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go-wrapper install    # "go install -v ./..."

ENV GIN_MODE release
ENV PORT 80
EXPOSE 80
CMD ["go-wrapper", "run"] # ["app"]
