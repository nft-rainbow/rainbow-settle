package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/conflux-gin-helper/middlewares"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
)

const (
	HEADER_KEY_TARGET_ADDR  = "X-Rainbow-Target-Addr"
	HEADER_KEY_TARGET_URL   = "X-Rainbow-Target-Url"
	HEADER_KEY_APPEND_QUERY = "X-Rainbow-Append-Query"
)

var logConfig = logger.LogConfig{Level: "trace", Folder: ".log", Format: "json"}

func main() {
	logger.Init(logConfig, "========== PROXY =============")
	// 创建代理服务器的 HTTP 处理函数
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetAddr := req.Header.Get(HEADER_KEY_TARGET_ADDR)
			targetUrl := req.Header.Get(HEADER_KEY_TARGET_URL)
			appendQuery := req.Header.Get(HEADER_KEY_APPEND_QUERY)

			// 替换URL
			func() {
				if targetUrl != "" {
					targetURL, err := url.Parse(targetUrl) // 修改为你要代理的目标URL
					if err != nil {
						log.Panicf("failed parse target url %s, %v: ", targetURL, err.Error())
					}
					logrus.WithField("target url", targetURL).Info("parse target url")
					req.URL.Scheme = targetURL.Scheme
					req.URL.Host = targetURL.Host
					req.Host = targetURL.Host
					req.URL.Path = targetURL.Path
					return
				}
				if targetAddr != "" {
					targetURL, err := url.Parse(targetAddr) // 修改为你要代理的目标服务器地址
					if err != nil {
						log.Panicf("failed parse target addr %s, %v: ", targetAddr, err.Error())
					}
					logrus.WithField("target addr", targetAddr).Info("parse target addr")
					req.URL.Scheme = targetURL.Scheme
					req.URL.Host = targetURL.Host
					req.Host = targetURL.Host
					return
				}
			}()

			// 增加HEADER
			func() {
				if appendQuery != "" {
					queries, err := url.ParseQuery(appendQuery)
					if err != nil {
						return
					}

					newQuery := req.URL.Query()
					for k, v := range queries {
						newQuery[k] = v
					}

					req.URL.RawQuery = newQuery.Encode()
				}
			}()
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
	if logConfig.Level == "trace" {
		engine.Use(rpcLogger)
	}
	return engine
}

func headerLog(header http.Header) interface{} {
	targetAddr := header.Get(HEADER_KEY_TARGET_ADDR)
	targetUrl := header.Get(HEADER_KEY_TARGET_URL)
	appendQuery := header.Get(HEADER_KEY_APPEND_QUERY)
	result := make(map[string]string)
	if targetAddr != "" {
		result[HEADER_KEY_TARGET_ADDR] = targetAddr
	}
	if targetUrl != "" {
		result[HEADER_KEY_TARGET_URL] = targetUrl
	}
	if appendQuery != "" {
		result[HEADER_KEY_APPEND_QUERY] = appendQuery
	}
	return result
}

// NOTE: temporarily use lock to avoid concurrent write file problems
var rpcLogMu sync.Mutex

func rpcLogger(c *gin.Context) {

	userId := c.Request.Header.Get(constants.RAINBOW_USER_ID_HEADER_KEY)
	costType := c.Request.Header.Get(constants.RAINBOW_COST_TYPE_HEADER_KEY)
	serverType := c.Request.Header.Get(constants.RAINBOW_SERVER_TYPE_HEADER_KEY)
	requestId := c.Request.Header.Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)
	_time := time.Now()

	c.Next()

	responseSize := c.Writer.Size()

	// write file
	rpcLogMu.Lock()
	defer rpcLogMu.Unlock()

	fileName := fmt.Sprintf(".request_log/%s_%s.log", userId, costType)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		logrus.WithField("file name", fileName).Error("failed to open file")
		return
	}

	defer file.Close()

	msg := fmt.Sprintf("%s, %s, %s, %s, %s, %d\n", _time.Format(time.StampMilli), userId, costType, serverType, requestId, responseSize)
	_, err2 := file.WriteString(msg)
	if err2 != nil {
		logrus.WithField("msg", msg).Error("failed to write request info")
	}
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
