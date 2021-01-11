FROM golang:1.15.6-alpine3.12 as build
WORKDIR /worker-service
COPY . .
RUN go get -d -v ./... && \
	go build -o worker-api ./internal/

FROM alpine:3.12
COPY --from=build /worker-service/worker-api /worker-api
COPY certs /certs
ENTRYPOINT ["/worker-api"]
