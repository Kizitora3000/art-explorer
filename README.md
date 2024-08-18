# misskey-renote-only-app
Misskeyのタイムラインのうち，リノートされたノートのみを表示するアプリケーション

# How to run

1. docker build -t misskey-art-explorer .
2. docker run -d -p 8080:8080 misskey-art-explorer

# How to deploy

## Azure login

1. az login
2. az acr login --name MisskeyArtExplorer

## Docker deploy

1. docker build -t misskeyartexplorer.azurecr.io/misskey-art-explorer:v<version> .
2. docker push misskeyartexplorer.azurecr.io/misskey-art-explorer:v<version>
3. art-explorer -> デプロイ センター -> タグ