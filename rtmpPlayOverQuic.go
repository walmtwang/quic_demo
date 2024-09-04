package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"log"
	"net/url"
	"quic_demo/quicConn"
	"quic_demo/rtmp"
	"strings"
)

func main() {

	var ip string
	var tcUrl string
	var streamName string
	var fileName string
	var port int
	flag.StringVar(&ip, "ip", "", "ip")
	flag.StringVar(&tcUrl, "tcUrl", "", "tcUrl")
	flag.StringVar(&streamName, "streamName", "", "streamName")
	flag.StringVar(&fileName, "fileName", "", "fileName")
	flag.IntVar(&port, "port", 443, "port, default 443")
	flag.Parse()
	if ip == "" || tcUrl == "" || streamName == "" || fileName == "" {
		log.Fatalln("ip == \"\" ||tcUrl == \"\" ||streamName == \"\" ||fileName == \"\"")

	}

	url2, err := url.Parse(tcUrl)
	if err != nil {
		log.Fatalf("url.Parse failed, err:%v", err)
	}
	domain := strings.Split(url2.Host, ":")[0]

	quicSession, err := quic.DialAddr(fmt.Sprintf("%s:%d", ip, port), &tls.Config{
		ServerName: domain,
		NextProtos: []string{"rtmp over quic"},
	}, &quic.Config{
		Versions: []quic.VersionNumber{quic.VersionDraft29},
	})
	if err != nil {
		log.Fatalf("quic.DialAddr err:%v", err)
		return
	}
	quicStream, err := quicSession.OpenStreamSync(context.Background())

	qConn := quicConn.NewQuicConn(quicSession, quicStream)

	rtmpPlay := rtmp.NewRtmpPlay(qConn, fileName,
		tcUrl,
		streamName)
	if err := rtmpPlay.Start(); err != nil {
		log.Fatalf("rtmpPlay.Start err:%v", err)
		return
	}
}
