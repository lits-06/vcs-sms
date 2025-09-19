package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
)

func (h *reportHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(h.middleware.RequireAuth())
	r.POST("/reports", h.middleware.RequireScopes(constants.ServerScopeReport), h.CreateReport)
}
