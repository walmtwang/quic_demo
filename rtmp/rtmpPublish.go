package rtmp

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/zhangpeihao/goflv"
	rtmp "github.com/zhangpeihao/gortmp"
)

type RtmpPublisher struct {
	Conn   net.Conn
	Stream rtmp.OutboundStream

	FlvFileName string
	TcUrl       string
	StreamName  string
	DurationMs  int64 // 推流时长
	TimeoutMs   int64 // 等待推流超时时间

	Status           uint
	IsClosed         bool
	CanPublisher     bool
	BeginTimeMs      int64
	PublisherBeginMs int64
}

func NewRtmpPublisher(conn net.Conn, flvFileName string, tcUrl string, streamName string,
) *RtmpPublisher {
	return &RtmpPublisher{
		Conn:             conn,
		FlvFileName:      flvFileName,
		TcUrl:            tcUrl,
		StreamName:       streamName,
		DurationMs:       1000 * 60 * 60 * 24 * 365 * 100,
		TimeoutMs:        3000,
		Status:           rtmp.OUTBOUND_CONN_STATUS_CLOSE,
		IsClosed:         false,
		CanPublisher:     false,
		PublisherBeginMs: 0,
	}
}

func (r *RtmpPublisher) OnStatus(conn rtmp.OutboundConn) {
	status, err := conn.Status()
	log.Printf("Handler On Status, status:%v, err:%v", status, err)
	r.Status = status
}

func (r *RtmpPublisher) OnClosed(conn rtmp.Conn) {
	log.Printf("Connect Closed")
	r.IsClosed = true
}

func (r *RtmpPublisher) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
	log.Printf("Received message")
}

func (r *RtmpPublisher) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	log.Printf("ReceviedRtmpCommand: %+v", command)
}

func (r *RtmpPublisher) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	log.Printf("Stream Created: %d", stream.ID())

	stream.Attach(r)

	if err := stream.Publish(r.StreamName, "live"); err != nil {
		log.Printf("OnStreamCreated failed, err:%v", err)
	}
}

func (r *RtmpPublisher) OnPlayStart(stream rtmp.OutboundStream) {
	log.Printf("Play Start")
}

func (r *RtmpPublisher) OnPublishStart(stream rtmp.OutboundStream) {
	log.Printf("Publish Start")
	r.PublisherBeginMs = time.Now().UnixNano() / 1e6
	r.Stream = stream
	r.CanPublisher = true
}

func (r *RtmpPublisher) Start() error {

	r.BeginTimeMs = time.Now().UnixNano() / 1e6

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

	// 开始推流
	err = r.PublishData()
	if err != nil {
		return fmt.Errorf("publish data failed, err:%v", err)
	}

	log.Printf("publish end")

	return nil
}

func (r *RtmpPublisher) PublishData() error {

	log.Printf("PublishData Start")
	// Set chunk buffer size
	flvFile, err := flv.OpenFile(r.FlvFileName)
	if err != nil {
		return fmt.Errorf("open flv file failed, "+
			"file:%v, err:%v", r.FlvFileName, err)
	}
	defer flvFile.Close()

	startTs := uint32(0)
	startAt := time.Now().UnixNano()
	needWaitTime := uint32(0)
	processedTime := uint32(0)

	for {

		// 连接已经断开，则退出
		if r.IsClosed {
			return fmt.Errorf("connection is closed")
		}

		nowTimeMs := time.Now().UnixNano() / 1e6
		if r.CanPublisher == false {
			// rtmp等待推流超时
			if (nowTimeMs - r.BeginTimeMs) > r.TimeoutMs {
				return fmt.Errorf("wait for publisher is timeout")
			}
			// 还没能推流，等待一下
			time.Sleep(time.Millisecond * 10)
			continue
		}

		// 异常状态
		if r.Status != rtmp.OUTBOUND_CONN_STATUS_CREATE_STREAM_OK {
			return fmt.Errorf("status is abnormal, status:%v", r.Status)
		}

		publisherDurationMs := time.Now().UnixNano()/1e6 - r.PublisherBeginMs
		if publisherDurationMs > r.DurationMs {
			return nil
		}

		// 判断是否需要sleep，实现flv的平滑发送
		processedTime = uint32((time.Now().UnixNano() - startAt) / 1e6)
		if needWaitTime > processedTime+100 {
			// 限制最大sleep时间，防止进程假死
			//r.Log.Printf("WaitTime:%d, processedTime:%d, needWaitTime:%d",
			//	needWaitTime, processedTime, needWaitTime-processedTime)
			sleepTime := math.Min(float64(needWaitTime-processedTime), MaxSleepTime)
			time.Sleep(time.Millisecond * time.Duration(sleepTime))
			// sleep后重新开始循环
			continue
		}

		// 如果文件已经读完，则重新开始推流
		if flvFile.IsFinished() {
			log.Printf("flv file is finished, loop")
			flvFile.LoopBack()
			startAt = time.Now().UnixNano()
			startTs = uint32(0)
		}

		// 获取下一个flv tag
		header, data, err := flvFile.ReadTag()
		if err != nil {
			// flv tag解析失败，退出
			return fmt.Errorf("flvFile.ReadTag() failed, "+
				"err:%v", err)
		}

		// 以flv文件第一个非0的timestamp为flv的起始时间
		// warn:如果第一个非0的timestamp有异常，则会导致后续内容无法推出去或者会一次性将内容全部推出去
		if startTs == uint32(0) {
			startTs = header.Timestamp
		}
		// 判断当前tag的时差
		if header.Timestamp > startTs {
			needWaitTime = header.Timestamp - startTs
		}
		// 推送当前tag
		if err = r.Stream.PublishData(header.TagType, data,
			needWaitTime); err != nil {
			return fmt.Errorf("stream.PublishData failed, err:%v", err)
		}

	}

	return fmt.Errorf("unkown reason")
}
