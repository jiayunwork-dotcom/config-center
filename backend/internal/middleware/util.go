package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func ParseJSONBody(c *gin.Context) (map[string]interface{}, bool) {
	if c.Request.Body == nil {
		return nil, false
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, false
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, false
	}
	return result, true
}
