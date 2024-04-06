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

// Online 用户上线
func (this *User) Online(server *server.Server) {
	// 用户上线，记录用户
	server.MapLock.Lock()
	server.OnlineUsers[this.Name] = this
	server.MapLock.Unlock()

	// 广播用户上线消息
	server.BroadCast(this, "已上线！")
}

// Offline 用户下线
func (this *User) Offline(server *server.Server) {
	// 用户上线，记录用户
	server.MapLock.Lock()
	delete(server.OnlineUsers, this.Name)
	server.MapLock.Unlock()

	// 广播用户上线消息
	server.BroadCast(this, "已下线线！")
}

// DoMsg 用户处理消息
func (this *User) DoMsg(msg string, server *server.Server) {
	if msg == "who" {
		server.MapLock.Lock()
		for _, tempUser := range server.OnlineUsers {
			onlineMsg := "[" + tempUser.Addr + "]" + tempUser.Name + "在线"
			this.SendMsg(onlineMsg)
		}
		server.MapLock.Unlock()
	}
	server.BroadCast(this, msg)
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
