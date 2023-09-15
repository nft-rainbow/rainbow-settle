package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/conflux-gin-helper/middlewares"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
)

func main() {
	logger.Init(logger.LogConfig{Level: "trace", Folder: ".log", Format: "json"}, "========== PROXY =============")
	// 创建代理服务器的 HTTP 处理函数
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetAddr := req.Header.Get("Target_addr")
			targetUrl := req.Header.Get("Target_url")
			log.Printf("Target Addr: %v", targetAddr)
			log.Printf("Target Url: %v", targetUrl)

			if targetUrl != "" {
				targetURL, _ := url.Parse(targetUrl) // 修改为你要代理的目标URL
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Host = targetURL.Host
				req.URL.Path = targetURL.Path
				return
			}
			if targetAddr != "" {
				targetURL, _ := url.Parse(targetAddr) // 修改为你要代理的目标服务器地址
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Host = targetURL.Host
				return
			}
		},
	}

	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		// 写入request id
		reqId := r.Header.Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)
		w.Header().Add(constants.RAINBOW_REQUEST_ID_HEADER_KEY, reqId)
		proxy.ServeHTTP(w, r)
	}

	proxyAddr := "0.0.0.0:8020"
	RunWithGin(proxyAddr, handlerFn)
}

func initGin() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(middlewares.Logger(&middlewares.LogOptions{HeaderLogger: headerLog}))
	return engine
}

func headerLog(header http.Header) interface{} {
	targetAddr := header.Get("Target_addr")
	targetUrl := header.Get("Target_url")
	if targetAddr != "" {
		return map[string]string{"Target_addr": targetAddr}
	}
	if targetUrl != "" {
		return map[string]string{"Target_url": targetUrl}
	}
	return nil
}

func RunWithGin(proxyAddr string, handlerFunc http.HandlerFunc) {

	app := initGin()
	app.Any("*path", gin.WrapF(handlerFunc))

	srv := &http.Server{
		Addr:    proxyAddr,
		Handler: app,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	fmt.Printf("Proxy server listening on %s...\n", proxyAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}

func RunWithRawHttp(proxyAddr string, handlerFunc http.HandlerFunc) {
	// 创建 HTTP 服务器并指定处理函数
	http.HandleFunc("/", handlerFunc)

	// 启动代理服务器
	fmt.Printf("Proxy server listening on %s...\n", proxyAddr)
	err := http.ListenAndServe(proxyAddr, nil)
	if err != nil {
		fmt.Println("Error starting proxy server:", err)
	}
}
