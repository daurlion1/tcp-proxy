package server

import (
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func GracefulShutdown(logger *logrus.Logger, stop chan struct{}, listener net.Listener, conns *sync.Map) {
	// Create a channel to receive the shutdown signal.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for the stop shutdown.
	<-shutdown
	logger.Println("Received stop signal. Stopping the proxy server...")

	// Close the stop channel to signal the proxy server to stop accepting new connections.
	close(stop)

	// Close the listener to stop accepting new connections.
	if listener != nil {
		err := listener.Close()
		if err != nil {
			logger.Printf("Error closing listener: %v\n", err)
		}
	}

	// Close all active connections.
	conns.Range(func(key, value interface{}) bool {
		conn := value.(net.Conn)
		err := conn.Close()
		if err != nil {
			logger.Printf("Error closing connection: %v\n", err)
		}
		return true
	})

	logger.Println("All connections have been closed. Shutting down.")

}
