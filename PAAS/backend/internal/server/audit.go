package server

import (
	"log"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
)

// logAudit writes an audit log entry to MongoDB. It does not fail the request on error.
func (s *Server) logAudit(c *gin.Context, userEmail string, action models.Action, adminInfo bool) {
	entry := models.AuditLog{
		UserEmail:     userEmail,
		Action:        action,
		AdminInfo:     adminInfo,
		Timestamp:     time.Now(),
		RequestMethod: c.Request.Method,
		RequestPath:   c.FullPath(),
		ClientIP:      c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
	}
	if err := s.db.InsertAuditLog(c.Request.Context(), &entry); err != nil {
		log.Printf("[audit] failed to write audit log: %v", err)
	}
}
