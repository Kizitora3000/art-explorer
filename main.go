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
type Note []struct {
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

	var Notes Note
	err = json.Unmarshal(responseBody, &Notes)
	if err != nil {
		return nil, fmt.Errorf("JSONデコードエラー: %v", err)
	}

	var notesToDisplay []NoteDisplay
	processedUsernames := make(map[string]bool) // 処理済みユーザーネームを追跡
	for i := 0; i < LIMIT && i < len(Notes); i++ {
		is_renote := Notes[i].RenoteID

		if is_renote == "" {
			continue
		}

		// ユーザーネームが既に処理済みの場合はスキップ
		if processedUsernames[Notes[i].Renote.User.Username] {
			continue
		}

		renote_user_id := Notes[i].Renote.User.UserId

		isFollowing, err := checkFollowStatus(accessToken, renote_user_id)
		if err != nil {
			return nil, fmt.Errorf("フォロー状態確認エラー: %v", err)
		}

		if !isFollowing {
			user_url := "https://misskey.io/@" + Notes[i].Renote.User.Username
			notesToDisplay = append(notesToDisplay, NoteDisplay{
				UserURL: user_url,
				Files:   Notes[i].Renote.Files,
			})
			processedUsernames[Notes[i].Renote.User.Username] = true // ユーザーネームを処理済みとしてマーク
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

	tmpl := `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Misskey Notes</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f0f0f0;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .note {
            background-color: white;
            border: 1px solid #ddd;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .user-link {
            display: block;
            margin-bottom: 10px;
            color: #0066cc;
            text-decoration: none;
            font-weight: bold;
        }
        .user-link:hover {
            text-decoration: underline;
        }
        .images-container {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            justify-content: start;
        }
        .note-image {
            max-width: 400px;
            max-height: 400px;
            width: auto;
            height: auto;
            object-fit: cover;
            border-radius: 4px;
        }
        @media (max-width: 600px) {
            .images-container {
                justify-content: center;
            }
            .note-image {
                max-width: 100%;
            }
        }
    </style>
</head>
<body>
    <h1>未フォローのユーザーのノート</h1>
    {{range .}}
    <div class="note">
        <a href="{{.UserURL}}" class="user-link">ユーザーページ</a>
        <div class="images-container">
            {{range .Files}}
            <img src="{{.URL}}" alt="Attached image" class="note-image" loading="lazy">
            {{end}}
        </div>
    </div>
    {{end}}
</body>
</html>
`

	t, err := template.New("notesTemplate").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
