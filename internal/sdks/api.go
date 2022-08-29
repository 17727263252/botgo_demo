package sdks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"wcxp-project/botgo_demo/config"
)

const (
	// TODO： 这些也可以写入配置文件
	TestDomain   = "https://sandbox.api.sgroup.qq.com"
	GetWssConUrl = "/gateway/bot"
	SendMsgUrl   = "/channels/"
)

type WssConUrlResp struct {
	URL               string `json:"url"`
	Shards            uint32 `json:"shards"`
	SessionStartLimit struct {
		Total          uint32 `json:"total"`
		Remaining      uint32 `json:"remaining"`
		ResetAfter     uint32 `json:"reset_after"`
		MaxConcurrency uint32 `json:"max_concurrency"`
	} `json:"session_start_limit"`
}

// MessageToCreate 发送消息结构体定义
type SendMessageReq struct {
	Content string `json:"content,omitempty"`
	Embed   *Embed `json:"embed,omitempty"`
	Image   string `json:"image,omitempty"`
	// 要回复的消息id，为空是主动消息，公域机器人会异步审核，不为空是被动消息，公域机器人会校验语料
	MsgID   string `json:"msg_id,omitempty"`
	EventID string `json:"event_id,omitempty"` // 要回复的事件id, 逻辑同MsgID
}
type SendMessageErrResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		MessageAudit struct {
			AuditID string `json:"audit_id"`
		} `json:"message_audit"`
	} `json:"data"`
}

type ApiSDK interface {
	GetWssConUrl(ctx context.Context) (resp WssConUrlResp, err error)
	SendMessage(ctx context.Context, channelID string, req SendMessageReq) (resp Message, err error)
}

var _ ApiSDK = &ApiClient{}

type ApiClient struct {
	httpCli *http.Client
}

func NewApiClient(httpClient *http.Client) *ApiClient {
	cli := &ApiClient{httpCli: httpClient}
	return cli
}

func (cli *ApiClient) GetWssConUrl(ctx context.Context) (WssConUrlResp, error) {
	realGetWssConUrl := fmt.Sprintf("%s%s", TestDomain, GetWssConUrl)
	getWssConUrlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, realGetWssConUrl, strings.NewReader(""))
	if err != nil {
		return WssConUrlResp{}, err
	}
	getWssConUrlReq.Header.Set("Authorization", fmt.Sprintf("Bot %d.%s", config.Conf.APPID, config.Conf.TOKEN))
	getWssConUrlReq.Header.Set("Content-Type", "application/json")
	getWssConUrlResp, err := cli.httpCli.Do(getWssConUrlReq)
	if err != nil {
		return WssConUrlResp{}, err
	}
	defer getWssConUrlResp.Body.Close()

	if getWssConUrlResp.StatusCode != http.StatusOK {
		return WssConUrlResp{}, err
	}
	respBuf := &bytes.Buffer{}
	_, err = respBuf.ReadFrom(getWssConUrlResp.Body)
	if err != nil {
		return WssConUrlResp{}, err
	}
	var realResp WssConUrlResp
	err = json.Unmarshal(respBuf.Bytes(), &realResp)
	if err != nil {
		return WssConUrlResp{}, err
	}
	if realResp.URL == "" {
		return WssConUrlResp{}, errors.New("invalid_url")
	}
	return realResp, nil
}

func (cli *ApiClient) SendMessage(ctx context.Context, channelID string, req SendMessageReq) (resp Message, err error) {
	sendMsgUrl := fmt.Sprintf("%s%s%s/messages", TestDomain, SendMsgUrl, channelID)
	reqParam, err := json.Marshal(req)
	if err != nil {
		return Message{}, err
	}
	senMsgReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sendMsgUrl, strings.NewReader(string(reqParam)))
	if err != nil {
		return Message{}, err
	}
	senMsgReq.Header.Set("Authorization", fmt.Sprintf("Bot %d.%s", config.Conf.APPID, config.Conf.TOKEN))
	senMsgReq.Header.Set("Content-Type", "application/json")
	senMsgResp, err := cli.httpCli.Do(senMsgReq)
	if err != nil {
		return Message{}, err
	}
	defer senMsgResp.Body.Close()
	if senMsgResp.StatusCode != http.StatusOK {
		return Message{}, err
	}
	respBuf := &bytes.Buffer{}
	_, err = respBuf.ReadFrom(senMsgResp.Body)
	if err != nil {
		return Message{}, err
	}
	err = json.Unmarshal(respBuf.Bytes(), &resp)
	if err != nil {
		return Message{}, err
	}
	return
}
