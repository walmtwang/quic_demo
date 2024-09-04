package flv

import (
	"errors"
	"fmt"
	"io"
)

const (
	AUDIO_TAG       = byte(0x08)
	VIDEO_TAG       = byte(0x09)
	SCRIPT_DATA_TAG = byte(0x12)
	DURATION_OFFSET = 53
	HEADER_LEN      = 13
)

type TagInfo struct {
	TagType   byte
	DataSize  uint32
	Timestamp uint32
	Body      []byte
}

type FlvParse struct {
	Reader io.Reader
}

func NewFlvParse(reader io.Reader) (*FlvParse, error) {

	flvParse := new(FlvParse)
	flvParse.Reader = reader

	// Read flv header
	remain := HEADER_LEN
	flvHeader := make([]byte, remain)

	if _, err := io.ReadFull(reader, flvHeader); err != nil {
		return nil, errors.New("File format error")
	}
	if flvHeader[0] != 'F' ||
		flvHeader[1] != 'L' ||
		flvHeader[2] != 'V' {
		return nil, errors.New("File format error")
	}

	return flvParse, nil
}

func (f *FlvParse) ReadTag() (*TagInfo, error) {
	tmpBuf := make([]byte, 4)
	tagInfo := &TagInfo{}
	// Read tag tagInfo
	if _, err := io.ReadFull(f.Reader, tmpBuf[3:]); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}
	tagInfo.TagType = tmpBuf[3]

	// Read tag size
	if _, err := io.ReadFull(f.Reader, tmpBuf[1:]); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}
	tagInfo.DataSize = uint32(tmpBuf[1])<<16 | uint32(tmpBuf[2])<<8 | uint32(tmpBuf[3])

	// Read timestamp
	if _, err := io.ReadFull(f.Reader, tmpBuf); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}
	tagInfo.Timestamp = uint32(tmpBuf[3])<<32 + uint32(tmpBuf[0])<<16 + uint32(tmpBuf[1])<<8 + uint32(tmpBuf[2])

	// Read stream ID
	if _, err := io.ReadFull(f.Reader, tmpBuf[1:]); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}

	// Read data
	data := make([]byte, tagInfo.DataSize)
	if _, err := io.ReadFull(f.Reader, data); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}
	tagInfo.Body = data

	// Read previous tag size
	if _, err := io.ReadFull(f.Reader, tmpBuf); err != nil {
		return nil, fmt.Errorf("io.ReadFull failed, err:%v", err)
	}

	return tagInfo, nil

}
