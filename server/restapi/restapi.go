package restapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conflux-fans/ginmetrics"
	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/middlewares"
	"github.com/nft-rainbow/rainbow-settle/server/config"
)

func initGin() *gin.Engine {
	engine := gin.New()
	ginmetrics.GetMonitor().Use(engine)

	engine.Use(gin.Logger())
	engine.Use(middlewares.Logger(nil))
	engine.Use(middlewares.Recovery())
	engine.Use(middlewares.RateLimitMiddleware)
	engine.Use(middlewares.Pagination())

	return engine
}

func Start() {

	app := initGin()

	// dashboard.SetupRoutes(app)
	// routers.SetupOpenAPIRoutes(app)
	// admin.SetupRoutes(app)
	// assets.SetupRoutes(app)

	port := config.Get().Port.RestApi
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: app,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
