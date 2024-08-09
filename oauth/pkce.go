package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"math/big"

	"github.com/google/uuid"
)

func PKCE() (string, string, uuid.UUID) {
	// ランダムな文字列を生成するための文字セット
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-._~"
	const length = 128

	// codeVerifierの生成
	codeVerifier := make([]byte, length)
	for i := range codeVerifier {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			log.Fatal(err)
		}
		codeVerifier[i] = chars[randomIndex.Int64()] // Int64(): *big.Int -> int64()
	}
	codeVerifierStr := string(codeVerifier) // []byte -> string

	// SHA256ハッシュを計算しbase64urlでエンコードしてcodeChallengeを生成
	hash := sha256.New()
	hash.Write([]byte(codeVerifierStr))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

	// ランダムなUUIDを生成
	state, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	return codeVerifierStr, codeChallenge, state
}
