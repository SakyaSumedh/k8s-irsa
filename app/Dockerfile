FROM golang:1.22-alpine

RUN apk update && apk add make

WORKDIR /app

COPY . ./

RUN make install && make build

EXPOSE 8080

CMD ["./demo-irsa"]