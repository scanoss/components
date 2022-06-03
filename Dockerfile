FROM golang:1.17 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go generate ./pkg/cmd/server.go
RUN go build -o ./scanoss-components ./cmd/server

FROM debian:buster-slim

WORKDIR /app
 
COPY --from=build /app/scanoss-components /app/scanoss-components

EXPOSE 50053

ENTRYPOINT ["./scanoss-dependencies"]
#CMD ["--help"]
