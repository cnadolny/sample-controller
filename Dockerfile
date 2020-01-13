FROM golang AS build

RUN mkdir /sample-app
WORKDIR /sample-app
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o sample main.go controller.go

FROM alpine:edge
WORKDIR /sample-app
RUN cd /sample-app
COPY --from=build /sample-app/sample /bin/
ENTRYPOINT ["sample"]