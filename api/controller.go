package api

import (
	"bitcoin-watcher/invoice_manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	router  *gin.Engine
	manager *invoice_manager.InvoiceManager
}

func New(manager *invoice_manager.InvoiceManager) *Controller {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
	}))

	c := Controller{
		router:  r,
		manager: manager,
	}

	r.GET("/history/:hash", c.history)

	return &c
}
