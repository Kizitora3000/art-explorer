package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// 取得したJSONデータは配列なので []struct と定義して合わせる
type Note []struct {
	RenoteID string `json:"renoteId"`
	Renote   struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
		Files []struct {
			URL string `json:"url"`
		} `json:"files"`
	} `json:"renote"`
}

func main() {
	// MisskeyのAPIエンドポイントURL
	apiURL := "https://misskey.io/api/notes/timeline"

	// リクエストボディの作成
	requestBody := map[string]interface{}{
		"i":     "WYyriPlqcHJnHAzcSDtNfOQ2bGJWkzbU",
		"limit": 1,
	}

	// リクエストボディをJSONに変換
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("JSON変換エラー:", err)
		return
	}

	// リクエストの作成
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("リクエスト作成エラー:", err)
		return
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")

	// クライアントの作成とリクエストの送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("リクエスト送信エラー:", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスの読み取り
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("レスポンス読み取りエラー:", err)
		return
	}

	var response Note
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		fmt.Println("JSONデコードエラー:", err)
		return
	}

	// レスポンスの表示
	fmt.Println(response[0])
}
