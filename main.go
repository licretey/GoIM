package main

import "GoIM/server"

func main() {
	newServer := server.NewServer("127.0.0.1", 8888)
	newServer.Start()
}
