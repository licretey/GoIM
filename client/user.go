package client

import "net"

type User struct {
	Name string
	Addr string
	// 连接通道
	conn net.Conn
	// 消息缓冲区
	MsgChan chan string
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		conn:    conn,
		MsgChan: make(chan string),
	}
	// 启动对消息通道的监听
	go user.ListenMsg()
	return user
}

// ListenMsg 监听当前用户的channel,一旦有信息发送给客户端
func (this *User) ListenMsg() {
	for {
		msg := <-this.MsgChan
		this.conn.Write([]byte(msg + "\n"))
	}
}
