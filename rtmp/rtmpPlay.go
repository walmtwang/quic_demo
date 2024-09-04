package rtmp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/zhangpeihao/goflv"
	rtmp "github.com/zhangpeihao/gortmp"
)

type RtmpPlay struct {
	Conn   net.Conn
	Stream rtmp.OutboundStream

	FlvFileName      string
	TcUrl            string
	StreamName       string
	FlvFile          *flv.File
	ErrorMessageChan chan string
}

func NewRtmpPlay(conn net.Conn, flvFileName string, tcUrl string, streamName string,
) *RtmpPlay {
	return &RtmpPlay{
		Conn:        conn,
		FlvFileName: flvFileName,
		TcUrl:       tcUrl,
		StreamName:  streamName,
	}
}

func (r *RtmpPlay) OnStatus(conn rtmp.OutboundConn) {
	status, err := conn.Status()
	log.Printf("Handler On Status, status:%v, err:%v", status, err)
}

func (r *RtmpPlay) OnClosed(conn rtmp.Conn) {
	log.Printf("Connect Closed")
}

func (r *RtmpPlay) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
	switch message.Type {
	case rtmp.VIDEO_TYPE:
		if r.FlvFile != nil {
			err := r.FlvFile.WriteVideoTag(message.Buf.Bytes(), message.Timestamp)
			if err != nil {
				r.ErrorMessageChan <- fmt.Sprintf("FlvFile.WriteVideoTag failed, err:%v", err)
			}
		}
	case rtmp.AUDIO_TYPE:
		if r.FlvFile != nil {
			err := r.FlvFile.WriteAudioTag(message.Buf.Bytes(), message.Timestamp)
			if err != nil {
				r.ErrorMessageChan <- fmt.Sprintf("FlvFile.WriteVideoTag failed, err:%v", err)
			}
		}
	case rtmp.DATA_AMF0:
		fallthrough
	case rtmp.DATA_AMF3:
		if r.FlvFile != nil {
			err := r.FlvFile.WriteTag(message.Buf.Bytes(), message.Type, message.AbsoluteTimestamp)
			if err != nil {
				r.ErrorMessageChan <- fmt.Sprintf("FlvFile.WriteVideoTag failed, err:%v", err)
			}
		}
	}
}

func (r *RtmpPlay) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	log.Printf("ReceviedRtmpCommand: %+v", command)
}

func (r *RtmpPlay) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	log.Printf("Stream Created: %d", stream.ID())

	stream.Attach(r)

	if err := stream.Play(r.StreamName, nil, nil, nil); err != nil {
		log.Printf("OnStreamCreated failed, err:%v", err)
	}
}

func (r *RtmpPlay) OnPlayStart(stream rtmp.OutboundStream) {
	log.Printf("Play Start")
	r.Stream = stream
}

func (r *RtmpPlay) OnPublishStart(stream rtmp.OutboundStream) {
	log.Printf("Play Start")
}

func (r *RtmpPlay) Start() error {

	var err error
	br := bufio.NewReader(r.Conn)
	bw := bufio.NewWriter(r.Conn)
	err = rtmp.Handshake(r.Conn, br, bw, time.Second*10)
	if err != nil {
		return fmt.Errorf("gortmp.Handshake err:%v", err)
	}

	obConn, err := rtmp.NewOutbounConn(r.Conn, r.TcUrl, r, 100)
	if err != nil {
		return fmt.Errorf("rtmp.NewOutbounConn failed, err:%v", err)
	}
	defer obConn.Close()
	// 通知握手成功
	r.OnStatus(obConn)

	err = obConn.Connect()
	if err != nil {
		return fmt.Errorf("obConn.Connect error: %v", err)
	}

	// 开始播放
	err = r.PlayData()
	if err != nil {
		return fmt.Errorf("play data failed, err:%v", err)
	}

	log.Printf("play end")

	return nil
}

func (r *RtmpPlay) PlayData() error {

	log.Printf("PlayData Start")
	// Set chunk buffer size

	if r.FlvFileName != "" {
		flvFile, err := flv.CreateFile(r.FlvFileName)
		if err != nil {
			return fmt.Errorf("open flv file failed, "+
				"file:%v, err:%v", r.FlvFileName, err)
		}
		defer flvFile.Close()
		r.FlvFile = flvFile
	}

	select {
	case errMessage := <-r.ErrorMessageChan:
		return fmt.Errorf("play error, err:%v", errMessage)
	}

	return fmt.Errorf("unkown reason")
}
