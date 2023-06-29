package main

import (
	"github.com/sirupsen/logrus"
	"github.com/tcp-proxy/proxy/internal/proxy"
	"github.com/tcp-proxy/proxy/internal/server"
	"net"
	"sync"
)

func main() {
	// Create a logger instance
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	// Start the proxy server
	proxyListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Fatalf("Unable to start proxy listener: %v\n", err)
	}
	defer proxyListener.Close()

	// Create a channel to signal the shutdown
	stop := make(chan struct{})

	activeConnections := new(sync.Map)

	// Run the proxy server
	go proxy.Run(logger, proxyListener, stop, activeConnections)

	// Perform graceful shutdown
	server.GracefulShutdown(logger, stop, proxyListener, activeConnections)

}
