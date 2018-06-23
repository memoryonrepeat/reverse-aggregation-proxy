FROM golang:1.10.3-alpine

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go build -o main . 
EXPOSE 8080
CMD ["/app/main"]