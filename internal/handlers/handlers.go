package handlers

import (
	"github.com/gin-gonic/gin"
)

type Handlers interface {
	RegisterHandler(router *gin.Engine)
}
