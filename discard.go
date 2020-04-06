package connutil

import (
	"errors"
	"io/ioutil"
	"net"
	"time"
)

// Discard is a net.Conn on which all Write calls succeed without doing anything
var Discard net.Conn = &discardConn{}

type discardConn struct{}

func (d *discardConn) Read(b []byte) (int, error) {
	return 0, errors.New("cannot read from a discardConn")
}

func (d *discardConn) Write(b []byte) (int, error) {
	return ioutil.Discard.Write(b)
}

func (d *discardConn) Close() error                       { return nil }
func (d *discardConn) SetReadDeadline(t time.Time) error  { return nil }
func (d *discardConn) SetWriteDeadline(t time.Time) error { return nil }
func (d *discardConn) SetDeadline(t time.Time) error      { return nil }
func (d *discardConn) LocalAddr() net.Addr                { return nil }
func (d *discardConn) RemoteAddr() net.Addr               { return nil }
