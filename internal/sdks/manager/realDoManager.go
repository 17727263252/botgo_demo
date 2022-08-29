package manager

import (
	errs "botgo_demo/err"
	"log"
	"math"
	"runtime"
	"wcxp-project/botgo_demo/internal/sdks"

	"fmt"
	"time"
)

// New 创建本地session管理器
func NewChanManager() *ChanManager {
	return &ChanManager{}
}

// ChanManager 默认的本地 session manager 实现
type ChanManager struct {
	sessionChan chan sdks.Session
}

// Start 启动本地 session manager
func (l *ChanManager) Start(apInfo *sdks.WssConUrlResp, token *sdks.Token, intents sdks.Intent) error {
	if apInfo.Shards > apInfo.SessionStartLimit.Remaining {
		log.Printf("[ws/session/local] session limited apInfo: %+v", apInfo)
		return errs.ErrSessionLimit
	}

	if apInfo.SessionStartLimit.MaxConcurrency == 0 {
		apInfo.SessionStartLimit.MaxConcurrency = 1
	}
	f := math.Round(concurrencyTimeWindowSec / float64(2))
	if f == 0 {
		f = 1
	}
	startInterval := time.Duration(f) * time.Second
	log.Printf("[ws/session/local] will start %d sessions and per session start interval is %s",
		apInfo.Shards, startInterval)

	// 按照shards数量初始化，用于启动连接的管理
	l.sessionChan = make(chan sdks.Session, apInfo.Shards)
	for i := uint32(0); i < apInfo.Shards; i++ {
		session := sdks.Session{
			URL:     apInfo.URL,
			Token:   *token,
			Intent:  intents, // todo : 只監聽艾特機器人事件
			LastSeq: 0,
			Shards: sdks.ShardConfig{
				ShardID:    i,
				ShardCount: apInfo.Shards,
			},
		}
		l.sessionChan <- session
	}

	for session := range l.sessionChan {
		// MaxConcurrency 代表的是每 5s 可以连多少个请求
		time.Sleep(startInterval)
		go l.newConnect(session)
	}
	return nil
}

// newConnect 启动一个新的连接，如果连接在监听过程中报错了，或者被远端关闭了链接，需要识别关闭的原因，能否继续 resume
// 如果能够 resume，则往 sessionChan 中放入带有 sessionID 的 session
// 如果不能，则清理掉 sessionID，将 session 放入 sessionChan 中
// session 的启动，交给 start 中的 for 循环执行，session 不自己递归进行重连，避免递归深度过深
func (l *ChanManager) newConnect(session sdks.Session) {
	defer func() {
		// panic 打印堆棧信息
		if err := recover(); err != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("[PANIC]%v\n%v\n%s\n", session, err, buf)
			l.sessionChan <- session
		}
	}()
	wsClient := sdks.NewWebSocketClient(session)
	if err := wsClient.Connect(); err != nil {
		log.Printf(err.Error())
		//l.sessionChan <- session // 连接失败，丢回去队列排队重连
		return
	}
	var err error

	// 初次鉴权  todo: 暫時不考慮重連的問題
	err = wsClient.Identify()

	if err != nil {
		log.Printf("[ws/session] Identify/Resume err %+v", err)
		return
	}
	if err = wsClient.Listening(); err != nil {
		log.Printf("[ws/session] Listening err %+v", err)
		currentSession := wsClient.Session()
		// todo: 暫時不考慮重連的問題
		// 一些错误不能够鉴权，比如机器人被封禁，这里就直接退出了
		if CanNotIdentify(err) {
			msg := fmt.Sprintf("can not identify because server return %+v, so process exit", err)
			log.Printf(msg)
			panic(msg) // 当机器人被下架，或者封禁，将不能再连接，所以 panic
		}
		// 将 session 放到 session chan 中，用于启动新的连接，当前连接退出
		l.sessionChan <- *currentSession
		return
	}
}
