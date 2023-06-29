package consts

import "time"

// Host represents the target host address.
const Host = "localhost"

// Port represents the target port.
const Port = "9090"

// ListenPort represents the listening port of the proxy server.
const ListenPort = "8080"

// ConnectionLimit represents the maximum number of concurrent connections allowed.
const ConnectionLimit = 2

// ConnectionTime represents the timeout duration for a connection.
const ConnectionTime = 30 * time.Second
