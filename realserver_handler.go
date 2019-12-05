package main

import (
	"log"
)

type RealServerMessageHandler struct {
	LpConnHandler *ConnHandler
	ConnPool      *ConnHandlerPool
	UserId        string
	ClientKey     string
}

func (messageHandler *RealServerMessageHandler) Encode(msg interface{}) []byte {
	if msg == nil {
		return []byte{}
	}

	return msg.([]byte)
}

func (messageHandler *RealServerMessageHandler) Decode(buf []byte) (interface{}, int) {
	return buf, len(buf)
}

func (messageHandler *RealServerMessageHandler) MessageReceived(connHandler *ConnHandler, msg interface{}) {
	if connHandler.NextConn != nil {
		data := msg.([]byte)
		message := Message{Type: P_TYPE_TRANSFER}
		message.Data = data
		connHandler.NextConn.Write(message)
	}
}

func (messageHandler *RealServerMessageHandler) ConnSuccess(connHandler *ConnHandler) {
	log.Println("获得代理连接:", messageHandler.UserId)
	proxyConnHandler, err := messageHandler.ConnPool.Get()
	if err != nil {
		log.Println("获得代理连接 err:", err, "uri:", messageHandler.UserId)
		message := Message{Type: TYPE_DISCONNECT}
		message.Uri = messageHandler.UserId
		messageHandler.LpConnHandler.Write(message)
		connHandler.conn.Close()
	} else {
		proxyConnHandler.NextConn = connHandler
		connHandler.NextConn = proxyConnHandler
		message := Message{Type: TYPE_CONNECT}
		message.Uri = messageHandler.UserId + "@" + messageHandler.ClientKey
		proxyConnHandler.Write(message)
		log.Println("RealServer的连接成功，通知访问代理服务器:", message.Uri)
	}
}

func (messageHandler *RealServerMessageHandler) ConnError(connHandler *ConnHandler) {
	conn := connHandler.NextConn
	if conn != nil {
		message := Message{Type: TYPE_DISCONNECT}
		message.Uri = messageHandler.UserId
		conn.Write(message)
		conn.NextConn = nil
	}

	connHandler.messageHandler = nil
}

func (messageHandler *RealServerMessageHandler) ConnFailed() {
	message := Message{Type: TYPE_DISCONNECT}
	message.Uri = messageHandler.UserId
	messageHandler.LpConnHandler.Write(message)
}
