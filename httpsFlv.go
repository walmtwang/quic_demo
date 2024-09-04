package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"quic_demo/flv"
	"time"
)

func main() {
	var url string
	var ip string
	var port int
	flag.StringVar(&url, "url", "", "such as: https://domain/live/stream.flv")
	flag.StringVar(&ip, "ip", "", "ip")
	flag.IntVar(&port, "port", 0, "port")
	flag.Parse()

	if url == "" || ip == "" || port == 0 {
		log.Fatalln("domain/uri/ip/port error")
	}

	dialer := net.Dialer{}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, fmt.Sprintf("%s:%d", ip, port))
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln("http.NewRequest failed, ", err)
	}

	beginTime := time.Now()
	fmt.Printf("begin time:%d\n", beginTime.UnixNano()/1e6)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("client.Do failed, ", err)
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
