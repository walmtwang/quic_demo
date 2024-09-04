package quicConn

import (
	"github.com/lucas-clemente/quic-go"
	"net"
	"time"
)

type QuicConn struct {
	QuicSession quic.Session
	QuicStream  quic.Stream
}

func NewQuicConn(quicSession quic.Session, quicStream quic.Stream) *QuicConn {
	return &QuicConn{
		QuicSession: quicSession,
		QuicStream:  quicStream,
	}
}

func (q *QuicConn) Read(b []byte) (n int, err error) {
	return q.QuicStream.Read(b)
}

func (q *QuicConn) Write(b []byte) (n int, err error) {
	return q.QuicStream.Write(b)
}

func (q *QuicConn) Close() error {
	return q.QuicStream.Close()
}

func (q *QuicConn) LocalAddr() net.Addr {
	return q.QuicSession.LocalAddr()
}

func (q *QuicConn) RemoteAddr() net.Addr {
	return q.QuicSession.RemoteAddr()
}

func (q *QuicConn) SetDeadline(t time.Time) error {
	return q.QuicStream.SetDeadline(t)
}

func (q *QuicConn) SetReadDeadline(t time.Time) error {
	return q.QuicStream.SetReadDeadline(t)
}

func (q *QuicConn) SetWriteDeadline(t time.Time) error {
	return q.QuicStream.SetWriteDeadline(t)
}
