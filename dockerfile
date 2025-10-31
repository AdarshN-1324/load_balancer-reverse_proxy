FROM golang:1.25.3-alpine3.22

WORKDIR /app

COPY . .

RUN go mod tidy 

ARG load_balancer

RUN go build -o load_balancer

EXPOSE 3001

CMD ["./load_balancer"]