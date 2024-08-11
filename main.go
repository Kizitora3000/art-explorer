package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// JSONレスポンスを返す関数
// gin.Context: HTTPリクエスト/レスポンス を管理する構造体
func indexHandler(c *gin.Context) {
	// ---------- MiAuth ----------

	// ランダムなUUIDを生成
	sessionID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	// TODO: 他のホストでログインしているユーザもいるため，ホストはユーザが選択できるようにする
	host := "misskey.io"

	authorizationURL := fmt.Sprintf("https://%s/miauth/%s", host, sessionID)

	// HTMLテンプレートに渡す
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"authorization_url": authorizationURL,
		"session_id":        sessionID, // for Debug
	})
}

func main() {
	// ginのコアとなるEngineインスタンスを作成
	r := gin.Default()

	// レンダリングするHTMLのディレクトリを指定
	r.LoadHTMLGlob("templates/*")

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	r.GET("/", indexHandler)

	// http://localhost:8080 でサーバを立てる
	r.Run()
}
