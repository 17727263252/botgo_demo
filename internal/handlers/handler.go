package handlers

import (
	"botgo_demo/internal/sdks"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
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

// Key: channlID+UserID  Value：Timestamp（签到时间戳）
var LocalClockMap = make(map[string]string)

//// 不考慮并发问题
//func GetLocalClockMap() map[string]string  {
//	if LocalClockMap == nil {
//		return make(map[string]string)
//	}
//	return LocalClockMap
//}

// ATMessageEventHandler TODO：重写处理 at 消息的回调 、每次收到（@+发消息）应该是独立的context
func ATMessageEventHandler(ctx context.Context, apiCli *sdks.ApiClient) sdks.ATMessageEventHandler {
	return func(event *sdks.WSPayload, data *sdks.WSATMessageData) error {
		log.Printf("[%s] guildID is %s, content is %s", event.Type, data.GuildID, data.Content)
		reqParam := sdks.SendMessageReq{}

		// 实际上获得的content格式为    "<@!8264696239026516860> /打卡"
		contentSlice := strings.Split(data.Content, "/")
		// 先调通了 TODO： 抽成解析实际内容的方法
		realContent := ""
		if len(contentSlice) != TwoLen {
			realContent = InvalidContent
		} else {
			realContent = contentSlice[1]
		}
		dateStr := time.Now().Format("2006年01月02日")
		// TODO： 这里可以抽成不同方法 （让代码简洁易读）
		switch realContent {
		// 指令介紹
		case ClockInstruction:
			reqParam.Content = fmt.Sprintf("1.输入'%s'即可签到,一天仅能签到一次喔！\n2.输入'%s'即可查看本频道下连续签到Top10排行榜,今天你上榜了吗？\n3.输入'%s'即可查看个人签到详情信息喔！", ClockIN, ClockINRanking, ClockINDataSelf)
		// 打卡
		case ClockIN:
			// who
			mapKey := data.Author.ID + "-" + dateStr + "-" + data.GuildID
			if _, ok := LocalClockMap[mapKey]; ok {
				reqParam.Content = fmt.Sprintf("<@%s>", data.Author.ID) + fmt.Sprintf("今天已经打过卡啦！")
			} else {
				LocalClockMap[mapKey] = FirstClock
				reqParam.Content = fmt.Sprintf("<@%s>", data.Author.ID) + fmt.Sprintf(dateStr+"成功打卡！")
			}
		case ClockINRanking:

		case ClockINDataSelf:

		case InvalidContent:

		}

		_, err := apiCli.SendMessage(ctx, data.ChannelID, reqParam)
		if err != nil {
			// 如果发送失败 又是首次打卡,应该删除map中的key-value，或者拿map中的value做标志
			log.Println("[send message fail] ,err:", err.Error())
			return err
		}
		return nil
	}
}
