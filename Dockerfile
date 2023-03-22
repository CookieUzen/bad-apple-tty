# to build this docker image:
#   docker build .
FROM gocv/opencv:4.7.0

ENV GOPATH /go

COPY . /go/src/
 
WORKDIR /go/src/

RUN go build -o /bad-apple-tty .

CMD ["/bad-apple-tty"]
