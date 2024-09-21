package login

import (
	"art-explorer/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func LoginHandler(c *gin.Context) {
	// セッションに関する話は後述
	session := sessions.Default(c)

	// ----- MiAuth step1 -----

	// 認証先のサーバを選択する
	// デフォルトなら misskey.io サーバでアカウントを作成しているはず
	host := "misskey.io"

	// ランダムなUUIDを生成
	sessionID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	// ----- MiAuth step2 -----

	// リダイレクト先のURLを設定
	// ローカル上の場合： http://localhost:8080/redirect
	// Azure上の場合： https://<azure site>/redirect
	// GetRootPath関数はのちほど実装
	redirectUri := fmt.Sprintf("%s/redirect", utils.GetRootPath(c))

	// アクセストークンが持つ権限：タイムラインの取得とフォロー状態の確認だけなので，アカウント情報を見る「read:account」のみ与える
	permission := "read:account"

	authorizationURL := fmt.Sprintf("https://%s/miauth/%s?callback=%s&permission=%s", host, sessionID, redirectUri, permission)

	// redirect先で使用するためhostを保存
	session.Set("host", host)
	session.Save()

	// HTMLテンプレートに渡す
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"authorization_url": authorizationURL,
	})
}

// 認証後のリダイレクト先
func RedirectHandler(c *gin.Context) {
	// ----- MiAuth step3 -----

	// LoginHandlerで作成したUUIDを取得
	// UUIDはURLのクエリパラメータで付いてくる
	redirectedSessionID := c.Query("session")

	// セッションから host を取得
	session := sessions.Default(c)
	redirectedHost := session.Get("host")

	getAccessTokenURL := fmt.Sprintf("https://%s/api/miauth/%s/check", redirectedHost, redirectedSessionID)

	// POSTリクエストを作成
	req, err := http.NewRequest("POST", getAccessTokenURL, bytes.NewBuffer([]byte("")))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating request: %s", err)
		return
	}

	// クライアントでリクエストを実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error making request: %s", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取る
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading response: %s", err)
		return
	}

	// レスポンスをJSONとしてパース
	var responseJson map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseJson); err != nil {
		c.String(http.StatusInternalServerError, "Error unmarshaling response: %s", err)
		return
	}

	// "token" キーの値をセッションに保存
	token, ok := responseJson["token"].(string)
	if ok {
		session.Set("token", token)
		session.Save()
	}

	// インデックスページにリダイレクト
	c.Redirect(http.StatusFound, "/")
}
