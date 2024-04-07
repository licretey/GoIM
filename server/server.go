package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户
	OnlineUsers map[string]*User
	MapLock     sync.RWMutex

	// 广播管道
	Message chan string
}

var MSG_SIZE = 4096

// NewServer 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:          ip,
		Port:        port,
		OnlineUsers: make(map[string]*User),
		Message:     make(chan string),
	}
	return server
}

// ListenMessage 监听Message广播消息，有时进行广播推送
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 发送消息给所有在线用户
		this.MapLock.Lock()
		for _, onlineUser := range this.OnlineUsers {
			onlineUser.MsgChan <- msg
		}
		this.MapLock.Unlock()
	}
}

// BroadCast 广播消息推送
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + ":" + msg
	this.Message <- sendMsg
}

// Handler 当前连接的业务：登录用户，广播消息
func (this *Server) Handler(conn net.Conn) {
	// 用户上线，记录用户
	user := NewUser(conn, this)
	user.Online()

	// 接收用户传递的消息
	go func() {
		buf := make([]byte, MSG_SIZE)
		for {
			readLength, err := conn.Read(buf)
			if readLength == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("读取客户信息异常", user, err)
				return
			}
			// 二进制消息转为字符串（去除\n）
			msg := string(buf[:readLength-1])
			user.DoMsg(msg)

			// 设置用户活跃
			user.IsAlive <- true
		}
	}()

	// 阻塞当前用户的会话
	for {
		select {
		case <-user.IsAlive:
			// 用户是活跃的，不用做任何事情（会执行下方case重置定时器）
		// 15s后定时器触发
		case <-time.After(time.Second * 10):

			user.SendMsg("您已被踢出系统")
			// 关闭用户缓冲区
			close(user.MsgChan)
			// 用户连接关闭
			conn.Close()
			// 推出handler
			return
		}
	}
}

// Start 启动server
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen error:", err)
		return
	}
	// close listen
	defer listener.Close()

	// 监听消息的推送
	go this.ListenMessage()

	// 监听链接的建立
	for {
		// accept
		connn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		// do callback: 连接上异步任务
		go this.Handler(connn)
	}
}
