package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/settle/config"
	"github.com/nft-rainbow/rainbow-settle/settle/proto"
	"github.com/nft-rainbow/rainbow-settle/settle/server"
	"github.com/nft-rainbow/rainbow-settle/settle/services"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	config.InitByFile("./config.yaml")
	logrus.WithField("config", config.Get()).Info("config loaded")
	logger.Init(config.Get().Log, "=============== SETTLE ==================")
	models.Init(config.Get().Mysql, config.Get().Fee, config.Get().CfxPrice)
	go startSettleServer()
	go services.Run()
	select {}
}

func startSettleServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Get().Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterSettleServer(s, &server.SettleServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
