package oauth

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type OAuthServerInfo struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

func GetOauthEndpoint() (string, string) {
	resp, err := http.Get("https://misskey.io/.well-known/oauth-authorization-server")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスのボディを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// JSONレスポンスを構造体にデコード
	var serverInfo OAuthServerInfo
	err = json.Unmarshal(body, &serverInfo)
	if err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	return serverInfo.AuthorizationEndpoint, serverInfo.TokenEndpoint
}
