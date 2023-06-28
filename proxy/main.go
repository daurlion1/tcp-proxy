package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	shutdown = make(chan os.Signal, 1)
	wg       sync.WaitGroup
)

const (
	host            = "server"
	port            = "9090"
	listenPort      = "8080"
	connectionLimit = 2
	connectionTime  = 30 * time.Second
)

type Channel struct {
	from, to net.Conn
	ack      chan bool
	timer    *time.Timer
}

func procMsg(c *Channel) {
	b := make([]byte, 10240)
	for {
		n, err := c.from.Read(b)
		if err != nil {
			break
		}
		c.timer.Reset(connectionTime)
		if n > 0 {
			c.to.Write(b[:n])
			log.Println(fmt.Sprintf("Message from %s to %s: %s", c.from.RemoteAddr().String(), c.to.RemoteAddr().String(), string(b[:n])))
		}
	}

	c.ack <- true
}

func connClosed(conn net.Conn) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			ch <- struct{}{}
		}
	}()

	return ch
}

func procConn(ctx context.Context, local net.Conn, target string, semaphore chan struct{}, logger *logrus.Logger) {
	defer func() {
		<-semaphore
		local.Close()
		wg.Done()
	}()

	remote, err := net.Dial("tcp", target)
	if err != nil {
		logger.Printf("Unable to connect to %s: %v\n", target, err)
		return
	}
	defer remote.Close()
	ack := make(chan bool)
	timer := time.NewTimer(connectionTime)

	go procMsg(&Channel{remote, local, ack, timer})
	go procMsg(&Channel{local, remote, ack, timer})

	select {
	case <-ack:
	case <-timer.C:
		logger.Println(fmt.Sprintf("Connection closed by TimeOut between %s and %s", remote.RemoteAddr(), local.RemoteAddr()))
		local.Close()
		remote.Close()
		return
	}
	select {
	case <-connClosed(remote):
		logger.Println(fmt.Sprintf("Connection closed by Server between %s and %s", remote.RemoteAddr(), local.RemoteAddr()))
		return

	case <-connClosed(local):
		logger.Println(fmt.Sprintf("Connection closed by Client between %s and %s", local.RemoteAddr(), remote.RemoteAddr()))
		return
	}
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	target := net.JoinHostPort(host, port)
	logger.Printf("Start listening on port %s and forwarding data to %s\n", listenPort, target)

	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		logger.Fatalf("Unable to start listener: %v\n", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdown
		logger.Println("Received shutdown signal. Gracefully shutting down...")

		cancelFunc()
		ln.Close()
		wg.Wait()

		logger.Println("All connections have been closed. Shutting down.")
	}()

	semaphore := make(chan struct{}, connectionLimit)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				// Context canceled, exit the loop
				logger.Println("Listener stopped accepting new connections")
				return
			default:
				logger.Printf("Accept failed: %v\n", err)
				continue
			}
		}
		select {
		case semaphore <- struct{}{}:
			wg.Add(1)
			go procConn(ctx, conn, target, semaphore, logger)
			logger.Println("New connection to the server:", conn.RemoteAddr())
		default:
			logger.Println("Connection limit reached, rejecting new connection")
			conn.Close()
		}
	}
}
