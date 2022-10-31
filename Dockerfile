FROM golang:1.19.1

LABEL maintainer="Eric Shanks <eshanks@purestorage.com>"

RUN mkdir /porxbbq

ADD . /porxbbq

WORKDIR /porxbbq

RUN go get . && go build -o main .

EXPOSE 80

CMD ["./main"]