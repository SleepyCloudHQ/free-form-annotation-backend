FROM golang:1.18.4-alpine3.16 AS builder
RUN apk add build-base
WORKDIR /go/src/free-form-annotation-backend
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY .  .
RUN GOOS=linux GOARCH=amd64 go build --tags json1 -o build/backend_app app/main.go  
RUN GOOS=linux GOARCH=amd64 go build --tags json1 -o build/migrate commands/migrate/migrate.go  
RUN GOOS=linux GOARCH=amd64 go build --tags json1 -o build/create_user commands/create_user/create_user.go  
RUN GOOS=linux GOARCH=amd64 go build --tags json1 -o build/export_dataset commands/export_dataset/export_dataset.go  
RUN GOOS=linux GOARCH=amd64 go build --tags json1 -o build/import_dataset commands/import_dataset/import_dataset.go  

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/free-form-annotation-backend/build/. ./
CMD ["./backend_app"]
