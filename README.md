# connutil [![](https://godoc.org/github.com/cbeuw/connutil?status.svg)](https://pkg.go.dev/github.com/cbeuw/connutil?tab=doc) [![](https://travis-ci.org/cbeuw/connutil.svg?branch=master)](https://travis-ci.org/github/cbeuw/connutil) [![](https://codecov.io/gh/cbeuw/connutil/branch/master/graph/badge.svg)](https://codecov.io/gh/cbeuw/connutil)
Utility net.Conns for testing and more
## Usage
`$ go get github.com/cbeuw/connutil`

To test a tls implementation, for instance
```go
package test

import (
    "crypto/tls"
    "testing"

    "github.com/cbeuw/connutil"
)

func TestTLS(t *testing.T){
    // AsyncPipe() works just like net.Pipe(), but is asynchronous and behaves more
    // like two ends of a TCP connection.
    // Read calls will block until data becomes available, and Write calls
    // are buffered.
    clientEnd, serverEnd := connutil.AsyncPipe()

    clientConn := tls.Client(clientEnd, clientConfig)
    serverConn := tls.Server(serverEnd, serverConfig)

    // do things with clientConn and serverConn...
    // e.g.
    clientConn.Handshake()
    go serverConn.Handshake()
    err := clientConn.VerifyHostname("example.com")
    if err != nil {
        t.Error(err)
    }
}
```