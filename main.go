package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kizitora3000/misskey-renote-only-app/oauth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// ルートパスのURLを取得
func getClientId(c *gin.Context) string {
	scheme := "https"

	// 原因不明: Azure App Serviceにデプロイすると，c.Request.TLS == nil となり scheme := "http" 扱いになるので一度コメントアウト
	/*
		if c.Request.TLS != nil {
			scheme = "https"
		}
	*/

	host := c.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}

// ログインページ
func indexHandler(c *gin.Context) {
	// PKCE用の情報を生成
	codeVerifier, codeChallenge, state := oauth.PKCE()
	fmt.Println(codeVerifier, codeChallenge, state)

	authorizationEndpoint, tokenEndpoint := oauth.GetOauthEndpoint()

	// ルートパスを自動的に取得
	clientId := getClientId(c)
	codeChallengeMethod := "S256" // 常にS256
	redirectUri := fmt.Sprintf("%s/redirect", clientId)
	scope := "read:account" // アカウントの情報を見る権限

	authorizationUrl := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=%s&state=%s",
		authorizationEndpoint, clientId, redirectUri, scope, codeChallenge, codeChallengeMethod, state)
	fmt.Println(authorizationUrl)
	fmt.Println(tokenEndpoint)

	// セッションに `state` を保存
	session := sessions.Default(c)
	session.Set("tokenEndpoint", tokenEndpoint)
	session.Set("clientId", clientId)
	session.Set("redirectUri", redirectUri)
	session.Set("scope", scope)
	session.Set("codeVerifier", codeVerifier)
	session.Set("state", state.String())
	session.Save()

	// HTMLテンプレートに渡す
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"authorization_url": authorizationUrl,
		"client_id":         clientId,
		"redirect_uri":      redirectUri,
		"scope":             scope,
		"code_verifier":     codeVerifier,
	})
}

// 認可コードを受け取るためのハンドラー
func redirectHandler(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// セッションから `state` を取得してチェック
	session := sessions.Default(c)
	savedState := session.Get("state")

	if savedState != state {
		c.String(http.StatusUnauthorized, "State does not match. Unauthorized access.\nState: %s\nsavedState: %s", state, savedState)
		return
	}

	// 認証が成功したことをユーザーに通知
	c.String(http.StatusOK, "Authorization successful.\nAuthorization code: %s\nState: %s\nsavedState: %s\nYou can close this window.\n", code, state, savedState)

	// アクセストークンを取得する
	clientId := session.Get("clientId").(string)
	redirectUri := session.Get("redirectUri").(string)
	scope := session.Get("scope").(string)
	codeVerifier := session.Get("codeVerifier").(string)
	endpoint := session.Get("tokenEndpoint").(string)

	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientId,
		"redirect_uri":  redirectUri,
		"scope":         scope,
		"code":          code,
		"code_verifier": codeVerifier,
	}

	// JSONを作成
	accessTokenRequestJson, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling JSON"})
		return
	}

	// POSTリクエストを作成
	accessTokenRequest, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(accessTokenRequestJson))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	// ヘッダーにContent-Typeを指定
	accessTokenRequest.Header.Set("Content-Type", "application/json")

	// クライアントでリクエストを実行
	client := &http.Client{}
	accessTokenResponse, err := client.Do(accessTokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error making request"})
		return
	}
	defer accessTokenResponse.Body.Close()

	// レスポンスボディを読み取る
	responseBody, err := io.ReadAll(accessTokenResponse.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
		return
	}

	hash := sha256.New()
	hash.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

	c.String(http.StatusOK, "\nclient_id: %s\nredirect_uri: %s\ncode: %s\nscope: %s\ncode_verifier: %s\ncode_challenge: %s\n", clientId, redirectUri, code, scope, codeVerifier, codeChallenge)

	// JSONとして解析して返す
	if accessTokenResponse.StatusCode == http.StatusOK {
		var accessTokenResponseJson map[string]interface{}
		if err := json.Unmarshal(responseBody, &accessTokenResponseJson); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshaling response body"})
			return
		}
		c.JSON(http.StatusOK, accessTokenResponseJson)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         fmt.Sprintf("Request failed with status code: %d", accessTokenResponse.StatusCode),
			"response_body": string(responseBody),
		})
	}
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// セッションミドルウェアを追加
	// 同じユーザーが同じセッションでアクセスした際の値を管理することができる
	// indexHandlerで保存した値をredirectHandlerで取得できるのはこの機能のおかげ
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// ルートエンドポイント"/"にGETリクエストを処理するハンドラーを登録
	router.GET("/", indexHandler)

	// 認可コードを受け取る/redirectエンドポイントを登録
	router.GET("/redirect", redirectHandler)

	// http://localhost:8080 でサーバを立てる
	router.Run()
}
