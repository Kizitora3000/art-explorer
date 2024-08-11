#!/bin/bash

# 引数のチェック
if [ -z "$1" ]; then
  echo "バージョン番号を指定してください。例: ./deploy.sh 1.0.0"
  exit 1
fi

# バージョン番号を変数に格納
VERSION=$1

# Dockerイメージのビルド
docker build -t misskeyartexplorer.azurecr.io/misskey-art-explorer:v$VERSION .

# Dockerイメージのプッシュ
docker push misskeyartexplorer.azurecr.io/misskey-art-explorer:v$VERSION