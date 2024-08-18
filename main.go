package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/Kizitora3000/misskey-renote-only-app/fetch"
	"github.com/Kizitora3000/misskey-renote-only-app/login"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// メインページ
// gin.Context: HTTPリクエスト/レスポンス を管理する構造体
func indexHandler(c *gin.Context) {
	session := sessions.Default(c)

	// ---------- check access token ----------
	// セッションからtokenを取得して表示
	token := session.Get("token")

	if token != nil {
		// アクセストークンが存在する場合はそのまま
		fmt.Println("Token found in session:", token)
	} else {
		// アクセストークンが存在しない場合はログインページに遷移
		fmt.Println("No token found in session")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	notes, err := fetch.FetchNotes(token)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	tmplPath := filepath.Join("templates", "index.tmpl")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// テンプレート(html)にnotes(データ)をバインドすることで最終的なHTMLを生成する
	err = t.Execute(c.Writer, notes)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
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

	// アクセストークンが存在しない場合に遷移するログインページを/loginエンドポイントを登録
	r.GET("/login", login.LoginHandler)

	// 認証後のリダイレクト先である/redirectエンドポイントを登録
	r.GET("/redirect", login.RedirectHandler)

	// http://localhost:8080 でサーバを立てる
	r.Run()
}
