package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetRootPath(c *gin.Context) string {
	scheme := "http"

	// 原因不明: Azure App Serviceにデプロイすると，c.Request.TLS == nil となり scheme := "http" 扱いになるので一度コメントアウト
	/*
		if c.Request.TLS != nil {
			scheme = "https"
		}
	*/

	host := c.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}
