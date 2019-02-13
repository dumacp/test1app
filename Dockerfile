FROM golang:1.9

WORKDIR /go/src/test1app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["test1app"]
