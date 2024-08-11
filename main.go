package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getRootPath(c *gin.Context) string {
	scheme := "http"

	// 原因不明: Azure App Serviceにデプロイすると，c.Request.TLS == nil となり scheme := "http" 扱いになるので一度コメントアウト
	/*
		if c.Request.TLS != nil {
			scheme = "https"
		}
	*/

	host := c.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}

// JSONレスポンスを返す関数
// gin.Context: HTTPリクエスト/レスポンス を管理する構造体
func indexHandler(c *gin.Context) {
	// ---------- MiAuth ----------

	// TODO: 他のホストでログインしているユーザもいるため，ホストはユーザが選択できるようにする
	host := "misskey.io"

	// ランダムなUUIDを生成
	sessionID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	redirectUri := fmt.Sprintf("%s/redirect", getRootPath(c))

	authorizationURL := fmt.Sprintf("https://%s/miauth/%s?callback=%s", host, sessionID, redirectUri)

	// redirect先で使用するためhostを保存
	session := sessions.Default(c)
	session.Set("host", host)
	session.Save()

	// HTMLテンプレートに渡す
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"authorization_url": authorizationURL,
		"session_id":        sessionID, // for Debug
	})
}

// 認証後のリダイレクト先
func redirectHandler(c *gin.Context) {
	redirectedSessionID := c.Query("session")

	// セッションから host を取得
	session := sessions.Default(c)
	redirectedHost := session.Get("host")

	c.String(http.StatusOK, "This is a redirect page.\nid: %s\nho: %s", redirectedSessionID, redirectedHost)
}

func main() {
	// ginのコアとなるEngineインスタンスを作成
	r := gin.Default()

	// レンダリングするHTMLのディレクトリを指定
	r.LoadHTMLGlob("templates/*")

	// セッションミドルウェアを追加
	// 同じユーザーが同じセッションでアクセスした際の値を管理することができる
	// indexHandlerで保存した値をredirectHandlerで取得できるのはこの機能のおかげ
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	r.GET("/", indexHandler)

	// 認証後のリダイレクト先である/redirectエンドポイントを登録
	r.GET("/redirect", redirectHandler)

	// http://localhost:8080 でサーバを立てる
	r.Run()
}
