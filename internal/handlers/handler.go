package handlers

import (
	"context"
	"log"
	"strings"
	"time"
	"wcxp-project/botgo_demo/internal/sdks"
)

const (
	ClockInstruction = "打卡指南"  // 返回跟打卡相关的所有指令  eg: 以下均为打卡指令
	ClockIN          = "打卡"    // 维度：天
	ClockINRanking   = "打卡排行榜" // 维度：月
	ClockINDataSelf  = "我的打卡"  // 维度：天/月
	InvalidContent   = "非法指令"
	FirstClock       = "first_clock"
	TwoLen           = 2
)

// ATMessageEventHandler TODO：重写处理 at 消息的回调 、每次收到（@+发消息）应该是独立的context
func ATMessageEventHandler(ctx context.Context, apiCli *sdks.ApiClient) sdks.ATMessageEventHandler {
	return func(event *sdks.WSPayload, data *sdks.WSATMessageData) error {
		log.Printf("[%s] [guildID is %s] [content is %s]", event.Type, data.GuildID, data.Content)
		// 解析content内容
		realContent := ParseContent(data.Content)
		// 事件分发
		reqParam := ClockInEventHandOut(realContent, data)
		// 消息响应
		_, err := apiCli.SendMessage(ctx, data.ChannelID, reqParam)
		if err != nil {
			// 如果发送失败 又是首次打卡,应该删除map中的key-value，或者拿map中的value做标志
			log.Println("[send message fail] ,err:", err.Error())
			return err
		}
		return nil
	}
}

func ParseContent(content string) (realContent string) {
	// 实际上获得的content格式为    "<@!8264696239026516860> /打卡"
	contentSlice := strings.Split(content, "/")
	if len(contentSlice) != TwoLen {
		realContent = InvalidContent
	} else {
		realContent = contentSlice[1]
	}
	return realContent
}

func ClockInEventHandOut(content string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	dateStr := time.Now().Format("2006年01月02日")
	// TODO： 这里可以抽成不同方法 （让代码简洁易读）
	switch content {
	// 指令介紹
	case ClockInstruction:
		reqMsg = realDoClockInstruction(dateStr, data)
	// 打卡
	case ClockIN:
		reqMsg = realDoClockIn(dateStr, data)
	// 打卡排行榜
	case ClockINRanking:
		reqMsg = realDoClockINRanking(dateStr, data)
	// 打卡详情
	case ClockINDataSelf:
		reqMsg = realDoClockINDataSelf(dateStr, data)
	// 非法内容处理
	default:
		reqMsg = realDoInvalidContent(dateStr, data)
	}
	return reqMsg
}
