package main

import (
	"github.com/jander/golog/logger"
	"github.com/kardianos/service"
	"os"
	"sync"
	"net"
	"fmt"
	"log"
)



var (
	 CR ConfigResult
	 repeater_map map[string]string
	repeater_map_lock sync.Mutex
	 repeater_register_map map[string]string
	repeater_register_map_lock sync.Mutex


	 conn_map map[string]net.Conn
	 conn_map_lock sync.Mutex
)

const (
	tcplen = 12  //tcp长度位 12个字节  4字节报文头  8字节报文体
	TcpHeadLen = 4
	TcpBodyLen = 8
)

func RepeaterServer() {
	listener, err := net.Listen("tcp", "localhost:10001")
	if err!= nil{
		log.Fatal(err)
	}
	fmt.Println("服务器建立成功!")
	for {
		//等待客户端接入
		conn, err := listener.Accept()
		if err != nil{
			log.Println(err.Error())
		}
		//开一个goroutines处理客户端消息，这是golang的特色，实现并发就只go一下就好
		go DoServerHandle(conn)
	}
}



func init() {
	repeater_map = make(map[string]string)
	repeater_register_map = make(map[string]string)
	conn_map = make(map[string]net.Conn)
	ConfigLoad()
	ConfigToMap()
}












/*
1：接收消息
2：获取mac
3：数据结构  map[string][string] 方便查找 两个 map地址小大顺序拼接
4：查找是不是在mac_pair 里面
5：解析json报文 获取 4个字节的长度位
6: 获取报文头
7：conn 连接有个 mac 对应的map
6: 建立一个本地客户端
7: 报文传输的 mac 设计 SrcMac DesMac
8: 添加失踪保持一个连接的功能
*/


func main(){
	service_entry()
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	RepeaterServer()
}

func (p *program) Stop(s service.Service) error {
	return nil
}


func service_entry() {
	svcConfig := &service.Config{
		Name:        "a_repeater_server", //服务显示名称
		DisplayName: "a_repeater_server", //服务名称
		Description: "a_repeater_server", //服务描述
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Fatal(err)
	}

	if err != nil {
		logger.Fatal(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			s.Install()
			logger.Println("服务安装成功")
			return
		}

		if os.Args[1] == "remove" {
			s.Uninstall()
			logger.Println("服务卸载成功")
			return
		}
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}


