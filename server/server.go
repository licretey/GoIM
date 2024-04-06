package server

import (
	"GoIM/client"
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户
	OnlineUser map[string]*client.User
	mapLock    sync.RWMutex

	// 广播管道
	Message chan string
}

var MSG_SIZE = 4096

// NewServer 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:         ip,
		Port:       port,
		OnlineUser: make(map[string]*client.User),
		Message:    make(chan string),
	}
	return server
}

// ListenMessage 监听Message广播消息，有时进行广播推送
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 发送消息给所有在线用户
		this.mapLock.Lock()
		for _, onlineUser := range this.OnlineUser {
			onlineUser.MsgChan <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息推送
func (this *Server) BroadCast(user *client.User, msg string) {
	sendMsg := "[" + user.Addr + "]" + ":" + msg
	this.Message <- sendMsg
}

// Handler 当前连接的业务：登录用户，广播消息
func (this *Server) Handler(conn net.Conn) {
	// 用户上线，记录用户
	this.mapLock.Lock()
	user := client.NewUser(conn)
	this.OnlineUser[user.Name] = user
	this.mapLock.Unlock()

	// 广播用户上线消息
	this.BroadCast(user, "已上线！")
	// 接收用户传递的消息
	go func() {
		buf := make([]byte, MSG_SIZE)
		for {
			readLength, err := conn.Read(buf)
			if readLength == 0 {
				this.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("读取客户信息异常", user, err)
				return
			}
			// 二进制消息转为字符串（去除\n）
			msg := string(buf[:readLength-1])
			this.BroadCast(user, msg)
		}
	}()
	// 阻塞当前会话
	select {}
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
