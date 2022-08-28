package utils

import (
	"bytes"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return getBytesBuffer()
		},
	}
)

func getBytesBuffer() *bytes.Buffer {
	return &bytes.Buffer{}
}

func GetBytesBuffer() *bytes.Buffer {
	bf, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok || bf == nil {
		bf = getBytesBuffer()
	}
	bf.Reset()
	return bf
}

func PutBytesBuffer(buf *bytes.Buffer) {
	bufferPool.Put(buf)
}
