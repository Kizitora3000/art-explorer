package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	LIMIT = 50
)

// RelationResponse はフォロー関係のAPIレスポンスを表す構造体
type RelationResponse struct {
	Following bool `json:"following"`
}

// Note はMisskeyのノート（投稿）を表す構造体
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

// NoteDisplay は表示用のノート情報を格納する構造体
type NoteDisplay struct {
	UserURL string
	Files   []struct {
		URL string `json:"url"`
	}
}

// checkFollowStatus はユーザーをフォローしているかどうかを確認する関数
func checkFollowStatus(accessToken, userId string) (bool, error) {
	apiURL := "https://misskey.io/api/users/relation"

	requestBody := map[string]string{
		"i":      accessToken,
		"userId": userId,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, fmt.Errorf("JSON変換エラー: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("リクエスト送信エラー: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	var relation RelationResponse
	err = json.Unmarshal(responseBody, &relation)
	if err != nil {
		return false, fmt.Errorf("JSONデコードエラー: %v", err)
	}

	return relation.Following, nil
}

// fetchNotes はMisskeyからノートを取得し、未フォローのユーザーのノートのみを返す関数
func fetchNotes() ([]NoteDisplay, error) {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// 環境変数を取得
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("ACCESS_TOKEN is not set in the environment")
	}

	apiURL := "https://misskey.io/api/notes/timeline"

	requestBody := map[string]interface{}{
		"i":     accessToken,
		"limit": LIMIT,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエスト送信エラー: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	var notes []Note
	err = json.Unmarshal(responseBody, &notes)
	if err != nil {
		return nil, fmt.Errorf("JSONデコードエラー: %v", err)
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

		isFollowing, err := checkFollowStatus(accessToken, renote_user_id)
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

// handleRoot はルートパスへのリクエストを処理するハンドラ関数
func handleRoot(w http.ResponseWriter, r *http.Request) {
	notes, err := fetchNotes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("templates", "index.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// テンプレート(html)にnotes(データ)をバインドすることで最終的なHTMLを生成する
	err = t.Execute(w, notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	// ルートハンドラの設定
	http.HandleFunc("/", handleRoot)

	// サーバーの起動
	fmt.Println("サーバーを起動しました。http://localhost:8080 にアクセスしてください。")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
