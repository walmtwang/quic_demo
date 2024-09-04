package rtmp

import "fmt"

const (
	programName  = "RtmpPublisher"
	version      = "0.0.1"
	MaxSleepTime = 1000
)

type PrintLog struct {
}

func (p *PrintLog) Printf(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}