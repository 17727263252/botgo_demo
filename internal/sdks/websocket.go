package sdks

import (
	errs "botgo_demo/err"
	"encoding/json"
	"fmt"
	wss "github.com/gorilla/websocket"
	"log"
	"runtime"

	"time"
)

type messageChan chan *WSPayload
type closeErrorChan chan error

type WebSocketClient struct {
	version         int
	conn            *wss.Conn
	messageQueue    messageChan
	session         *Session
	user            *WSUser
	closeChan       closeErrorChan
	heartBeatTicker *time.Ticker // 用于维持定时心跳
}

type WebSocketSDK interface {
	// New 创建一个新的ws实例，需要传递 session 对象
	//New(session sdks.Session) WebSocketClient

	// Connect 连接到 wss 地址
	Connect() error
	// Identify 鉴权连接
	Identify() error
	// Session 拉取 session 信息，包括 token，shard，seq 等
	Session() *Session
	// Listening 监听websocket事件
	Listening() error
	// Write 发送数据
	Write(message *WSPayload) error
	// Close 关闭连接
	Close()
}

var _ WebSocketSDK = &WebSocketClient{}

// New 新建一个连接对象
func NewWebSocketClient(session Session) WebSocketClient {
	return WebSocketClient{
		messageQueue:    make(messageChan, 100),
		session:         &session,
		closeChan:       make(closeErrorChan, 10),
		heartBeatTicker: time.NewTicker(60 * time.Second), // 先给一个默认 ticker，在收到 hello 包之后，会 reset
	}
}

func (cli *WebSocketClient) Connect() error {
	var err error
	cli.conn, _, err = wss.DefaultDialer.Dial(cli.session.URL, nil)
	if err != nil {
		log.Printf("%s, connect err: %v", cli.session, err)
		return err
	}
	log.Printf("%s, url %s, connected", cli.session, cli.session.URL)
	return nil
}

// Listening 开始监听，会阻塞进程，内部会从事件队列不断的读取事件，解析后投递到注册的 event handlers，如果读取消息过程中发生错误，会循环
// 定时心跳也在这里维护
func (cli *WebSocketClient) Listening() error {
	defer cli.Close()
	go cli.readMessageToQueue()
	go cli.listenMessageAndHandle()

	// todo :暫時不考慮重連問題
	// 接收 resume signal
	//resumeSignal := make(chan os.Signal, 1)
	//if websocket.ResumeSignal >= syscall.SIGHUP {
	//	signal.Notify(resumeSignal, websocket.ResumeSignal)
	//}

	// handlers message
	for {
		select {
		//case <-resumeSignal: // 使用信号量控制连接立即重连
		//	log.Infof("%s, received resumeSignal signal", cli.session)
		//	return errs.ErrNeedReConnect
		case err := <-cli.closeChan:
			// 关闭连接的错误码 https://bot.q.qq.com/wiki/develop/api/gateway/error/error.html
			log.Printf("%s Listening stop. err is %v", cli.session, err)
			// 不能够 identify 的错误
			if wss.IsCloseError(err, 4914, 4915) {
				err = errs.New(errs.CodeConnCloseCantIdentify, err.Error())
			}
			// 这里用 UnexpectedCloseError，如果有需要排除在外的 close error code，可以补充在第二个参数上
			// 4009: session time out, 发了 reconnect 之后马上关闭连接时候的错误码，这个是允许 resumeSignal 的
			if wss.IsUnexpectedCloseError(err, 4009) {
				err = errs.New(errs.CodeConnCloseCantResume, err.Error())
			}
			if DefaultHandlers.ErrorNotify != nil {
				// 通知到使用方错误
				DefaultHandlers.ErrorNotify(err)
			}
			return err
		case <-cli.heartBeatTicker.C:
			log.Printf("%s listened heartBeat", cli.session)
			heartBeatEvent := &WSPayload{
				WSPayloadBase: WSPayloadBase{
					OPCode: WSHeartbeat,
				},
				Data: cli.session.LastSeq,
			}
			// 不处理错误，Write 内部会处理，如果发生发包异常，会通知主协程退出
			_ = cli.Write(heartBeatEvent)
		}
	}
}

// Write 往 ws 写入数据
func (cli *WebSocketClient) Write(message *WSPayload) error {
	m, _ := json.Marshal(message)
	log.Printf("%s write %s message, %v", cli.session, message.OPCode, string(m))

	if err := cli.conn.WriteMessage(wss.TextMessage, m); err != nil {
		log.Printf("%s WriteMessage failed, %v", cli.session, err)
		cli.closeChan <- err
		return err
	}
	return nil
}

// Identify 对一个连接进行鉴权，并声明监听的 shard 信息
func (cli *WebSocketClient) Identify() error {
	// 避免传错 intent
	//if cli.session.Intent == 0 {
	//	cli.session.Intent = dto.IntentGuilds
	//}

	payload := &WSPayload{
		Data: &WSIdentityData{
			Token:   fmt.Sprintf("%v.%s", cli.session.Token.AppID, cli.session.Token.AccessToken),
			Intents: IntentGuildAtMessage,
			Shard: []uint32{
				cli.session.Shards.ShardID,
				cli.session.Shards.ShardCount,
			},
		},
	}
	payload.OPCode = 2
	return cli.Write(payload)
}

// Close 关闭连接
func (cli *WebSocketClient) Close() {
	if err := cli.conn.Close(); err != nil {
		log.Printf("%s, close conn err: %v", cli.session, err)
	}
	cli.heartBeatTicker.Stop()
}

// Session 获取client的session信息
func (cli *WebSocketClient) Session() *Session {
	return cli.session
}

func (cli *WebSocketClient) readMessageToQueue() {
	for {
		_, message, err := cli.conn.ReadMessage()
		if err != nil {
			log.Printf("%s read message failed, %v, message %s", cli.session, err, string(message))
			close(cli.messageQueue)
			cli.closeChan <- err
			return
		}
		payload := &WSPayload{}
		if err := json.Unmarshal(message, payload); err != nil {
			log.Printf("%s json failed, %v", cli.session, err)
			continue
		}
		payload.RawMessage = message
		log.Printf("%s receive %s message, %s", cli.session, payload.OPCode, string(message))
		// 处理内置的一些事件，如果处理成功，则这个事件不再投递给业务
		if cli.isHandleBuildIn(payload) {
			continue
		}
		cli.messageQueue <- payload
	}
}

func (cli *WebSocketClient) listenMessageAndHandle() {
	defer func() {
		// panic，一般是由于业务自己实现的 handle 不完善导致
		// 打印日志后，关闭这个连接，进入重连流程
		if err := recover(); err != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("[PANIC]%v\n%v\n%s\n", cli.session, err, buf)
			cli.closeChan <- fmt.Errorf("panic: %v", err)
		}
	}()
	for payload := range cli.messageQueue {
		cli.saveSeq(payload.Seq)
		// ready 事件需要特殊处理
		if payload.Type == "READY" {
			cli.readyHandler(payload)
			continue
		}
		// 解析具体事件，并投递给业务注册的 handlers
		if err := ParseAndHandle(payload); err != nil {
			log.Printf("%s parseAndHandle failed, %v", cli.session, err)
		}
	}
	log.Printf("%s message queue is closed", cli.session)
}

func (cli *WebSocketClient) saveSeq(seq uint32) {
	if seq > 0 {
		cli.session.LastSeq = seq
	}
}

// isHandleBuildIn 内置的事件处理，处理那些不需要业务方处理的事件
// return true 的时候说明事件已经被处理了
func (cli *WebSocketClient) isHandleBuildIn(payload *WSPayload) bool {
	switch payload.OPCode {
	// 接收到 hello 后需要开始发心跳
	case WSHello:
		cli.startHeartBeatTicker(payload.RawMessage)
	// 心跳 ack 不需要业务处理
	case WSHeartbeatAck:
	default:
		return false
	}
	return true
}

// startHeartBeatTicker 启动定时心跳
func (cli *WebSocketClient) startHeartBeatTicker(message []byte) {
	helloData := &WSHelloData{}
	if err := ParseData(message, helloData); err != nil {
		log.Printf("%s hello data parse failed, %v, message %v", cli.session, err, message)
	}
	// 根据 hello 的回包，重新设置心跳的定时器时间
	cli.heartBeatTicker.Reset(time.Duration(helloData.HeartbeatInterval) * time.Millisecond)
}

// readyHandler 针对ready返回的处理，需要记录 sessionID 等相关信息
func (cli *WebSocketClient) readyHandler(payload *WSPayload) {
	readyData := WSReadyData{}
	if err := ParseData(payload.RawMessage, &readyData); err != nil {
		log.Printf("%s parseReadyData failed, %v, message %v", cli.session, err, payload.RawMessage)
	}
	cli.version = readyData.Version
	// 基于 ready 事件，更新 session 信息
	cli.session.ID = readyData.SessionID
	cli.session.Shards.ShardID = readyData.Shard[0]
	cli.session.Shards.ShardCount = readyData.Shard[1]
	cli.user = &WSUser{
		ID:       readyData.User.ID,
		Username: readyData.User.Username,
		Bot:      readyData.User.Bot,
	}
	// 调用自定义的 ready 回调
	if DefaultHandlers.Ready != nil {
		DefaultHandlers.Ready(payload, readyData)
	}
}
