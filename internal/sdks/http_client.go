package sdks

import (
	"net"
	"net/http"
	"time"
)

const (
	DefaultTimeout = 5 * time.Second

	HeaderContentType   = "Content-Type"
	HeaderContentLength = "Content-Length"

	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json"
)

// 参数暂时这么写
func NewDefaultTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     120 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		//ExpectContinueTimeout: 1 * time.Second,
	}
}

func NewDefaultHttpClient(opt ...CliOpt) *http.Client {
	cli := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: NewDefaultTransport(),
	}
	for _, o := range opt {
		o(cli)
	}
	return cli
}

// 支持自定义func
type CliOpt func(cli *http.Client)
