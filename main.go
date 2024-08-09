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
	router := gin.Default()

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	router.GET("/", index)

	// http://localhost:8080 でサーバを立てる
	router.Run()
}
