FROM golang:1.22

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# go build app file
RUN go build -v -o /usr/local/bin/app ./...

# run app file
CMD ["app"]
