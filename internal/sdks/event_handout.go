package sdks

import (
	"encoding/json"

	"github.com/tidwall/gjson" // 由于回包的 d 类型不确定，gjson 用于从回包json中提取 d 并进行针对性的解析
)

var eventParseFuncMap = map[OPCode]map[EventType]eventParseFunc{
	WSDispatchEvent: {
		"AT_MESSAGE_CREATE": atMessageHandler,
	},
}

type eventParseFunc func(event *WSPayload, message []byte) error

// ParseAndHandle 处理回调事件
func ParseAndHandle(payload *WSPayload) error {
	// 指定类型的 handlers
	if h, ok := eventParseFuncMap[payload.OPCode][payload.Type]; ok {
		return h(payload, payload.RawMessage)
	}
	// 透传handler，如果未注册具体类型的 handlers，会统一投递到这个 handlers
	if DefaultHandlers.Plain != nil {
		return DefaultHandlers.Plain(payload, payload.RawMessage)
	}
	return nil
}

// ParseData 解析数据
func ParseData(message []byte, target interface{}) error {
	data := gjson.Get(string(message), "d")
	return json.Unmarshal([]byte(data.String()), target)
}

func atMessageHandler(payload *WSPayload, message []byte) error {
	data := &WSATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.ATMessage != nil {
		return DefaultHandlers.ATMessage(payload, data)
	}
	return nil
}
