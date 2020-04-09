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

    clientConn := tls.Client(clientEnd, fooClientConfig)
    serverConn := tls.Server(serverEnd, fooServerConfig)

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

Or to test the behaviour of an http server
```go
package test

import(
    "net/http"
    "testing"

    "github.com/cbeuw/connutil"
)

func TestHttp(t *testing.T){
    // DialerListener returns a pair of Dialer and net.Listener
    dialer, listener := connutil.DialerListener(128)

    go http.Serve(listener, fooHandler)
    
    // Here you get one end of an async pipe. http.Serve will get the other end from
    // listener.Accept() 
    clientConn, err := dialer.Dial("","")
    if err != nil {
        // handle error
    }
    
    // Do stuff with the client conn...
    clientConn.Write([]byte("GET ..."))
}
```