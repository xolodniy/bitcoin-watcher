package api

import (
	"bitcoin-watcher/common"
	"bitcoin-watcher/config"
	"bitcoin-watcher/invoice_manager"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

func (c *Controller) Start() {
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Main.HTTP.Port),
		Handler: c.router,
		// ReadTimeout:       300 * time.Second,
		// ReadHeaderTimeout: 300 * time.Second,
		// WriteTimeout:      300 * time.Second,
		// IdleTimeout:       300 * time.Second,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			common.Logger().Errorw("Failed to start HTTP server", "err", err)
		}
	}()
}
