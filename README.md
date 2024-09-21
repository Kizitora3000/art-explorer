# art-explorer

Misskey のタイムラインのうち，リノートされたノートのみを表示するアプリケーション

## 記事

https://zenn.dev/kizitora/articles/64df5ad69aaa1f

## How to run

1. docker build -t art-explorer .
2. docker run -d -p 8080:8080 art-explorer

## How to deploy

### Azure login

1. az login
2. az acr login --name artExplorer

### Docker deploy

1. docker build -t artexplorer.azurecr.io/art-explorer:v<version> .
2. docker push artexplorer.azurecr.io/art-explorer:v<version>
3. art-explorer -> デプロイ センター -> タグ
