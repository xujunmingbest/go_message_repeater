package main

import (
	"io/ioutil"
	"encoding/xml"
	"log"
	"sync"
	"fmt"
)

/*

设计方案
1:每个客户端都有一个ID
2: 根据客户端ID匹配进行消息转发
3:
*/

var (
	 CR ConfigResult
	 lock sync.Mutex
)

/*
func startServer() {
	//连接主机、端口，采用ｔｃｐ方式通信，监听７７７７端口
	listener, err := net.Listen("tcp", "localhost:7777")
	checkError(err)
	fmt.Println("建立成功!")
	for {
		//等待客户端接入
		conn, err := listener.Accept()
		checkError(err)
		//开一个goroutines处理客户端消息，这是golang的特色，实现并发就只go一下就好
		go doServerStuff(conn)
	}
}
*/

func main(){
	ConfigLoad()
	p := Repeater_pair{"mac1","mac3"}
	fmt.Println(FindPair(&p))
	fmt.Println(FindPair(&p))
	fmt.Println(FindPair(&p))
	fmt.Println(FindPair(&p))
}

func ConfigReload(){




}

type ConfigResult struct {
	XMLName xml.Name `xml:"message_repeater"`
	Repeater_max_number int `xml:"repeater_max_number"`
	Mrs []Repeater_pair `xml:"repeater_pair"`
}

type Repeater_pair struct {
	Pair_mac1 string `xml:"pair1_mac1"`
	Pair_mac2 string `xml:"pair1_mac2"`
}


func ConfigLoad() {
	content, err := ioutil.ReadFile("config.xml")
	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(content, &CR)
	if err != nil {
		log.Fatal(err)
	}
}

func FindPair(p *Repeater_pair)(b bool){
	lock.Lock()
	defer lock.Unlock()
	for _,v := range CR.Mrs{
        if v.Pair_mac1 == p.Pair_mac1 && v.Pair_mac2 ==  p.Pair_mac2{
        	return true;
		}else if v.Pair_mac1 == p.Pair_mac2 && v.Pair_mac2 ==  p.Pair_mac1{
			return true;
		}
	}
    return false;
}