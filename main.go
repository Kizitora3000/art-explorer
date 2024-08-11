package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Kizitora3000/misskey-renote-only-app/login"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	LIMIT = 50
)

// フォロー済みか否かのAPIリクエストの結果を格納する構造体
type RelationResponse struct {
	Following bool `json:"following"`
}

// タイムラインのノートの情報を格納する構造体
type Note struct {
	RenoteID string `json:"renoteId"`
	Renote   struct {
		User struct {
			UserId   string `json:"userId"`
			Username string `json:"username"`
		} `json:"user"`
		Files []struct {
			URL string `json:"url"`
		} `json:"files"`
	} `json:"renote"`
}

// index.htmlで表示する情報を格納する構造体
type NoteDisplay struct {
	UserURL string
	Files   []struct {
		URL string `json:"url"`
	}
}

// sendPostRequest は共通のHTTP POSTリクエストを送信する関数
// requestBody interface{}, responseStruct interface{} と定義することで，異なる構造のリクエストボディやデータ構造を受け入れられる
func sendPostRequest(apiURL string, requestBody interface{}, responseStruct interface{}) error {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("JSON変換エラー: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("リクエスト送信エラー: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	err = json.Unmarshal(responseBody, responseStruct)
	if err != nil {
		return fmt.Errorf("JSONデコードエラー: %v, %s", err, string(responseBody))
	}

	return nil
}

// checkFollowStatus はユーザーをフォローしているかどうかを確認する関数
func checkFollowStatus(token interface{}, userId string) (bool, error) {
	apiURL := "https://misskey.io/api/users/relation"

	requestBody := map[string]string{
		"i":      token.(string),
		"userId": userId,
	}

	var relation RelationResponse
	err := sendPostRequest(apiURL, requestBody, &relation)
	if err != nil {
		return false, err
	}

	return relation.Following, nil
}

// fetchNotes はMisskeyからノートを取得し、未フォローのユーザーのノートのみを返す関数
func fetchNotes(token interface{}) ([]NoteDisplay, error) {
	if token == nil {
		log.Fatal("ACCESS_TOKEN is not set in the environment")
	}

	apiURL := "https://misskey.io/api/notes/timeline"

	requestBody := map[string]interface{}{
		"i":     token,
		"limit": LIMIT,
	}

	var notes []Note
	err := sendPostRequest(apiURL, requestBody, &notes)
	if err != nil {
		return nil, err
	}

	var notesToDisplay []NoteDisplay
	processedUsernames := make(map[string]bool) // 処理済みユーザーネームを追跡
	for i := 0; i < LIMIT && i < len(notes); i++ {
		is_renote := notes[i].RenoteID

		if is_renote == "" {
			continue
		}

		// ユーザーネームが既に処理済みの場合はスキップ
		if processedUsernames[notes[i].Renote.User.Username] {
			continue
		}

		renote_user_id := notes[i].Renote.User.UserId

		isFollowing, err := checkFollowStatus(token, renote_user_id)
		if err != nil {
			return nil, fmt.Errorf("フォロー状態確認エラー: %v", err)
		}

		if !isFollowing {
			user_url := "https://misskey.io/@" + notes[i].Renote.User.Username
			notesToDisplay = append(notesToDisplay, NoteDisplay{
				UserURL: user_url,
				Files:   notes[i].Renote.Files,
			})
			processedUsernames[notes[i].Renote.User.Username] = true // ユーザーネームを処理済みとしてマーク
		}
	}

	return notesToDisplay, nil
}

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

	notes, err := fetchNotes(token)
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
