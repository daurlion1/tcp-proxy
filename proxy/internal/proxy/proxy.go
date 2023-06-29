package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tcp-proxy/proxy/internal/consts"
)

// Channel represents the data flow between two connections.
type Channel struct {
	from  net.Conn
	to    net.Conn
	ack   chan bool
	timer *time.Timer
}

// procMsg reads data from the source connection and writes it to the destination connection.
func procMsg(c *Channel) {
	b := make([]byte, 10240)
	for {
		n, err := c.from.Read(b)
		if err != nil {
			break
		}
		c.timer.Reset(consts.ConnectionTime)
		if n > 0 {
			c.to.Write(b[:n])
			fmt.Printf("Message from %s to %s: %s\n", c.from.RemoteAddr().String(), c.to.RemoteAddr().String(), string(b[:n]))
		}
	}

	c.ack <- true
}

// connClosed returns a channel that signals when the connection is closed.
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

// procConn handles the forwarding of data between the local and remote connections.
func procConn(local net.Conn, target string, semaphore chan struct{}, logger *logrus.Logger) {
	defer func() {
		<-semaphore
		local.Close()
	}()

	remote, err := net.Dial("tcp", target)
	if err != nil {
		logger.Printf("Unable to connect to %s: %v\n", target, err)
		return
	}
	defer remote.Close()
	ack := make(chan bool)
	timer := time.NewTimer(consts.ConnectionTime)

	go procMsg(&Channel{remote, local, ack, timer})
	go procMsg(&Channel{local, remote, ack, timer})

	select {
	case <-ack:
		// Data transfer completed successfully.
	case <-timer.C:
		logger.Println(fmt.Sprintf("Connection closed by TimeOut between %s and %s", remote.RemoteAddr(), local.RemoteAddr()))
		local.Close()
		remote.Close()
		return
	}

	select {
	case <-connClosed(remote):
		logger.Println(fmt.Sprintf("Connection closed between %s and %s", remote.RemoteAddr(), local.RemoteAddr()))
		local.Close()
		return

	case <-connClosed(local):
		logger.Println(fmt.Sprintf("Connection closed between %s and %s", local.RemoteAddr(), remote.RemoteAddr()))
		remote.Close()
		return
	}
}

// RunProxy starts the proxy server and listens for incoming connections.
func RunProxy(logger *logrus.Logger, listener net.Listener, stop chan struct{}, activeConnections *sync.Map) {
	target := net.JoinHostPort(consts.Host, consts.Port)
	logger.Printf("Start listening on port %s and forwarding data to %s\n", consts.ListenPort, target)

	semaphore := make(chan struct{}, consts.ConnectionLimit)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stop:
				logger.Println("Listener stopped accepting new connections")
				return
			default:
				logger.Printf("Accept failed: %v\n", err)
				continue
			}
		}

		select {
		case <-stop:
			logger.Println("Received stop signal. Stopping the proxy server...")
			// Close all active connections
			activeConnections.Range(func(key, value interface{}) bool {
				conn := key.(net.Conn)
				conn.Close()
				activeConnections.Delete(key)
				return true
			})
			return
		case semaphore <- struct{}{}:
			activeConnections.Store(conn, conn)

			go func() {
				defer func() {
					activeConnections.Delete(conn)
					conn.Close()
					<-semaphore
				}()

				procConn(conn, target, semaphore, logger)
			}()

			logger.Println("New connection to the server:", conn.RemoteAddr())

		default:
			logger.Println("Connection limit reached, rejecting new connection")
			conn.Close()
		}
	}
}
