package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/nft-rainbow/rainbow-settle/proto"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/nft-rainbow/rainbow-settle/server/restapi"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func init() {
	config.InitByFile("./config.yaml")
	logrus.WithField("config", config.Get()).Info("config loaded")
	logger.Init(config.Get().Log, "=============== SETTLE ==================")
	models.Init(config.Get().Mysql, config.Get().Fee, config.Get().CfxPrice)
	redis.Init(config.Get().Redis)
	services.Init()
}

func main() {
	go startSettleServer()
	go restapi.Run()
	go services.Run()

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
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("Server exiting")
}

func startSettleServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Get().Port.Grpc))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterSettleServer(s, &SettleServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
