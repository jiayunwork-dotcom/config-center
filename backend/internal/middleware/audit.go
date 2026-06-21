package middleware

import (
	"bytes"
	"io"
	"net/http"

	"config-center/internal/models"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func getClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}
	return c.ClientIP()
}

func AuditMiddleware() gin.HandlerFunc {
	auditService := services.NewAuditService()

	return func(c *gin.Context) {
		userID := GetUserID(c)
		username := GetUsername(c)

		if userID == 0 {
			c.Next()
			return
		}

		method := c.Request.Method
		if method == http.MethodGet {
			c.Next()
			return
		}

		auditData := GetAuditData(c)
		if auditData == nil {
			auditData = &AuditData{}
		}

		if auditData.Action == "" {
			switch method {
			case http.MethodPost:
				auditData.Action = models.ActionCreate
			case http.MethodPut:
				auditData.Action = models.ActionUpdate
			case http.MethodDelete:
				auditData.Action = models.ActionDelete
			}
		}

		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if auditData.NewValue == "" && len(bodyBytes) > 0 {
				auditData.NewValue = string(bodyBytes)
			}
		}

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			if auditData.Action != "" && auditData.ResourceType != "" {
				uid := userID
				auditService.CreateLog(
					&uid,
					username,
					auditData.Action,
					auditData.ResourceType,
					auditData.ResourceID,
					auditData.ResourceName,
					auditData.OldValue,
					auditData.NewValue,
					getClientIP(c),
				)
			}
		}
	}
}
