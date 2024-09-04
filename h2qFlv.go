package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"log"
	"net/http"
	"net/url"
	"quic_demo/flv"
	"strings"
	"time"
)

func main() {

	var ip string
	var port int
	var httpUrl string
	flag.StringVar(&ip, "ip", "", "server ip")
	flag.StringVar(&httpUrl, "url", "", "http url, https://domain/live/stream.flv")
	flag.IntVar(&port, "port", 443, "server port, default 443")
	flag.Parse()
	if ip == "" || httpUrl == "" {
		log.Fatalln("ip == \"\" ||  url == \"\"")
	}

	url2, err := url.Parse(httpUrl)
	if err != nil {
		log.Fatalf("url.Parse failed, err:%v", err)
	}

	domain := strings.Split(url2.Host, ":")[0]

	roundTripper := &http3.RoundTripper{
		QuicConfig: &quic.Config{
			Versions: []quic.VersionNumber{quic.VersionDraft29},
		},
		TLSClientConfig: &tls.Config{
			ServerName: domain,
			NextProtos: []string{"rtmp over quic"},
		},
		Dial: func(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
			return quic.DialAddrEarly(fmt.Sprintf("%s:%d", ip, port), tlsCfg, cfg)
		},
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}
	beginTime := time.Now()
	fmt.Printf("begin time:%d\n", beginTime.UnixNano()/1e6)
	resp, err := hclient.Get(httpUrl)
	if err != nil {
		log.Fatalf("hclient.Get err:%v", err)
	}
	respBody := resp.Body
	defer respBody.Close()

	if resp.StatusCode != 200 {
		log.Fatalln("resp.StatusCode error, ", resp.StatusCode)
	}

	flvParse, err := flv.NewFlvParse(respBody)
	lastTime := beginTime
	for true {

		tagInfo, err := flvParse.ReadTag()
		if err != nil {
			log.Fatalln("flvParse.ReadTag error, ", err)

		}
		currentTime := time.Now()
		fmt.Printf("read tag, curr time:%d, interval:%d, type:%d, body:%d, pts:%d\n",
			currentTime.UnixNano()/1e6, currentTime.Sub(lastTime).Nanoseconds()/1e6, tagInfo.TagType, len(tagInfo.Body), tagInfo.Timestamp)

	}

}
