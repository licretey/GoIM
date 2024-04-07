package server

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	// 连接通道
	conn net.Conn
	// 消息缓冲区
	MsgChan chan string
	server  *Server

	// 是否活跃
	IsAlive chan bool
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		conn:    conn,
		MsgChan: make(chan string),
		server:  server,
		IsAlive: make(chan bool),
	}
	// 激活用户
	//user.IsAlive <- true
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
	this.server.OnlineUsers[this.Name] = this
	this.server.MapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已上线！")
}

// Offline 用户下线
func (this *User) Offline() {
	// 用户上线，记录用户
	this.server.MapLock.Lock()
	delete(this.server.OnlineUsers, this.Name)
	this.server.MapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已下线线！")
}

// DoMsg 用户处理消息
func (this *User) DoMsg(msg string) {
	if msg == "who" {
		this.server.MapLock.Lock()
		for _, tempUser := range this.server.OnlineUsers {
			onlineMsg := "[" + tempUser.Addr + "]" + tempUser.Name + "在线"
			this.SendMsg(onlineMsg)
		}
		this.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		// 判断name是否重复
		_, ok := this.server.OnlineUsers[newName]
		if ok {
			this.SendMsg("用户名" + newName + "已被使用")
		} else {
			this.server.MapLock.Lock()
			delete(this.server.OnlineUsers, this.Name)
			this.server.OnlineUsers[newName] = this
			this.server.MapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已更新用户名:" + this.Name + "\n")
		}
	}
	this.server.BroadCast(this, msg)
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
