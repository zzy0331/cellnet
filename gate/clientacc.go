package gate

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/davyxu/cellnet/socket"
)

var ClientAcceptor cellnet.Peer

func getMsgName(msgid uint32) string {

	if meta := cellnet.MessageMetaByID(int(msgid)); meta != nil {
		return meta.Name
	}

	return ""
}

// 开启客户端侦听通道
func StartClientAcceptor(pipe cellnet.EventPipe, address string) {

	ClientAcceptor = socket.NewAcceptor(pipe)

	// 默认开启并发
	ClientAcceptor.EnableConcurrenceMode(true)

	// 所有接收到的消息转发到后台
	ClientAcceptor.InjectData(func(data interface{}) bool {

		if ev, ok := data.(*socket.SessionEvent); ok {

			// Socket各种事件不要往后台发
			switch ev.MsgID {
			case socket.Event_SessionAccepted,
				socket.Event_SessionConnected:
				return true
			}

			// 构建路由封包
			relaypkt, _ := cellnet.BuildPacket(&coredef.UpstreamACK{
				MsgID:    ev.MsgID,
				Data:     ev.Data,
				ClientID: ev.Ses.ID(),
			})

			// TODO 按照封包和逻辑固定分发
			// TODO 非法消息直接掐线
			// TODO 心跳, 超时掐线
			BackendAcceptor.IterateSession(func(ses cellnet.Session) bool {

				if DebugMode {
					log.Debugf("client->backend, msg: %s(%d) clientid: %d", getMsgName(ev.MsgID), ev.MsgID, ev.Ses.ID())
				}

				ses.RawSend(relaypkt)

				return true
			})

		}

		return false
	})

	ClientAcceptor.Start(address)

}
