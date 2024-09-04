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
	"strings"
)

func main() {

	var ip string
	var port int
	var httpUrl string
	var print int
	flag.StringVar(&ip, "ip", "", "server ip")
	flag.StringVar(&httpUrl, "url", "", "http url, https://domain/live/stream.flv")
	flag.IntVar(&port, "port", 443, "server port, default 443")
	flag.IntVar(&print, "print", 1, "print response, default 1")
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
	resp, err := hclient.Get(httpUrl)
	if err != nil {
		log.Fatalf("hclient.Get err:%v", err)
	}
	fmt.Printf("http status:%v\n", resp.StatusCode)
	if print > 0 {
		fmt.Printf("resp:")
		defer resp.Body.Close()
		for {
			buf := make([]byte, 1024)
			len, err := resp.Body.Read(buf)
			if err != nil {
				fmt.Printf("%s", string(buf[:len]))
				log.Fatalf("\nresp.Body.Read err:%v\n", err)
			}
			fmt.Printf("%s", string(buf[:len]))
		}
	}

}
