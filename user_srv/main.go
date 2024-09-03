package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/user_srv/handler"
	"mxshop_srvs/user_srv/proto"
	"net"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 8088, "port number")
	flag.Parse()
	fmt.Printf("ip: %s, port: %d\n", *ip, *port)
	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		panic(err)
	}
	server.Serve(listen)

}
