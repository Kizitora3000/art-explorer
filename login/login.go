package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kizitora3000/misskey-renote-only-app/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func LoginHandler(c *gin.Context) {
	session := sessions.Default(c)

	// ---------- MiAuth ----------

	// TODO: 他のホストでログインしているユーザもいるため，ホストはユーザが選択できるようにする
	host := "misskey.io"

	// ランダムなUUIDを生成
	sessionID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	redirectUri := fmt.Sprintf("%s/redirect", utils.GetRootPath(c))

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
