package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSONレスポンスを返す関数
// gin.Context: HTTPリクエスト/レスポンス を管理する構造体
func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func main() {
	// ginのコアとなるEngineインスタンスを作成
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	router.GET("/", indexHandler)

	// http://localhost:8080 でサーバを立てる
	router.Run()
}
