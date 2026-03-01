package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (c *Controller) history(ctx *gin.Context) {
	hash := ctx.Param("hash")
	c.manager.HandleScriptHash(hash)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
