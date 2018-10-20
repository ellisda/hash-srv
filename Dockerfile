FROM golang:latest 
RUN mkdir /app 
ADD . /usr/local/go/src/github.com/ellisda/hash-srv
WORKDIR /usr/local/go/src/github.com/ellisda/hash-srv
RUN go build -o main . 
CMD ["//usr/local/go/src/github.com/ellisda/hash-srv/main"]
