package client

import (
	"GoIM/server"
	"net"
)

type User struct {
	Name string
	Addr string
	// 连接通道
	conn net.Conn
	// 消息缓冲区
	MsgChan chan string
	server  *server.Server
}

func NewUser(conn net.Conn, userServer *server.Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		conn:    conn,
		MsgChan: make(chan string),
		server:  userServer,
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

// Online 用户上线
func (this *User) Online() {
	// 用户上线，记录用户
	this.server.MapLock.Lock()
	this.server.OnlineUser[this.Name] = this
	this.server.MapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已上线！")
}

// Offline 用户下线
func (this *User) Offline() {
	// 用户上线，记录用户
	this.server.MapLock.Lock()
	delete(this.server.OnlineUser, this.Name)
	this.server.MapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已下线线！")
}

// DoMsg 用户处理消息
func (this *User) DoMsg(msg string) {
	this.server.BroadCast(this, msg)
}
