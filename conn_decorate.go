package shuttle

import (
	"net"
	"time"
	"bytes"
	"github.com/sipt/shuttle/pool"
	"github.com/sipt/shuttle/util"
	"fmt"
)

var defaultTimeOut = 20 * time.Second

//
func DefaultDecorate(c net.Conn, network string) (IConn, error) {
	return &DefaultConn{
		Conn:    c,
		ID:      util.NextID(),
		Network: network,
	}, nil
}

type DefaultConn struct {
	net.Conn
	ID      int64
	Network string
}

func (c *DefaultConn) GetID() int64 {
	return c.ID
}

func (c *DefaultConn) Flush() (int, error) {
	return 0, nil
}

func (c *DefaultConn) GetNetwork() string {
	return c.Network
}

func (c *DefaultConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	Logger.Debugf("[ID:%d] Read Data: %v", c.GetID(), b[:n])
	return
}

func (c *DefaultConn) Write(b []byte) (n int, err error) {
	Logger.Debugf("[ID:%d] Write Data: %v", c.GetID(), b)
	return c.Conn.Write(b)
}
func (c *DefaultConn) Close() error {
	Logger.Debugf("[ID:%d] close connection", c.GetID())
	return c.Conn.Close()
}

//超时装饰
func TimerDecorate(c IConn, timeOut time.Duration) (IConn, error) {
	if timeOut <= 0 {
		timeOut = defaultTimeOut
	}
	return &TimerConn{
		IConn:   c,
		TimeOut: timeOut,
	}, nil
}

type TimerConn struct {
	IConn
	TimeOut time.Duration
}

func (c *TimerConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.TimeOut))
	n, err = c.IConn.Read(b)
	return
}

func (c *TimerConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.TimeOut))
	n, err = c.IConn.Write(b)
	return
}

//缓冲装饰
func BufferDecorate(c IConn) (IConn, error) {
	return &BufferConn{
		IConn:  c,
		buffer: bytes.NewBuffer(pool.GetBuf()[:0]),
	}, nil
}

type BufferConn struct {
	IConn
	buffer *bytes.Buffer
}

func (c *BufferConn) Write(b []byte) (n int, err error) {
	return c.buffer.Write(b)
}

func (c *BufferConn) Flush() (n int, err error) {
	n, err = c.IConn.Write(c.buffer.Bytes())
	if err != nil {
		return
	}
	c.buffer.Reset()
	n, err = c.IConn.Flush()
	fmt.Println("flush over !!!!!")
	return
}

//实时写出
func RealTimeDecorate(c IConn) (IConn, error) {
	return &RealTimeFlush{
		IConn: c,
	}, nil
}

type RealTimeFlush struct {
	IConn
}

func (r *RealTimeFlush) Write(b []byte) (n int, err error) {
	n, err = r.IConn.Write(b)
	if err != nil {
		return
	}
	_, err = r.IConn.Flush()
	return
}