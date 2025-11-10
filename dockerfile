# stage 1
# building the app using golang 
FROM golang:1.25.3-alpine3.22 AS golangbuild

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o load_balancer ./

# pushing the project build to a lighter version
# stage 2

FROM alpine:latest

WORKDIR /app

COPY --from=golangbuild /app/load_balancer /app/load_balancer

EXPOSE 3001

CMD ["./load_balancer"]