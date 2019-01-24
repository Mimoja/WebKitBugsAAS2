FROM golang:stretch

# install dependencies
RUN apt update
RUN apt upgrade -y
RUN apt install -y subversion

# install app
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get -v -d ./...
RUN go build -o main .
CMD ["/app/main"]
