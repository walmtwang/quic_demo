package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
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

	tcUrl = strings.Replace(tcUrl, "rtmps://", "rtmp://", -1)
	url2, err := url.Parse(tcUrl)
	if err != nil {
		log.Fatalf("url.Parse failed, err:%v", err)
	}
	domain := strings.Split(url2.Host, ":")[0]

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), &tls.Config{
		ServerName: domain,
	})
	if err != nil {
		log.Fatalf("tls.Dial failed, err:%v", err)
	}

	rtmpPublisher := rtmp.NewRtmpPublisher(conn, fileName,
		tcUrl,
		streamName)
	if err := rtmpPublisher.Start(); err != nil {
		log.Fatalf("rtmpPublisher.Start err:%v", err)
		return
	}
}
