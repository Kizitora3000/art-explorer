package main

import "github.com/gin-gonic/gin"

// JSONレスポンスを返す関数
// gin.Context: HTTPリクエスト/レスポンス を管理する構造体
func helloWorldHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}

func main() {
	// ginのコアとなるEngineインスタンスを作成
	r := gin.Default()

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	r.GET("/", helloWorldHandler)

	// http://localhost:8080 でサーバを立てる
	r.Run()
}
