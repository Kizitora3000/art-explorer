FROM golang:1.22

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# ビルド時にmain.goを指定して単一の実行ファイルを生成
RUN go build -v -o /usr/local/bin/app ./main.go

# 実行ファイルを指定
CMD ["/usr/local/bin/app"]
