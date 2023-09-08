package restapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/conflux-fans/ginmetrics"
	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/middlewares"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/nft-rainbow/rainbow-settle/server/restapi/controllers"
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

func Run() {

	app := initGin()
	controllers.SetupRouter(app)

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
}
