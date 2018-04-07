package main

import (
	"github.com/jander/golog/logger"
	"github.com/kardianos/service"
	"os"
	"sync"
	"net"
	"fmt"
	"log"
	"encoding/xml"
	"io/ioutil"
	"strings"
	"strconv"
	"encoding/json"
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
		go doServerHandle(conn)
	}
}


//处理客户端消息
func doServerHandle(conn net.Conn) {
	log.Println("有客户端连接")
	mac := ""
	defer fmt.Println("一个连接退出")
	defer conn.Close();
	defer RemoveMacConn(&mac)
	defer log.Println(conn_map)
	//预先接收 mac 地址
	//装载 mac和conn 到map

	for {
		//获取报文体的长度
		first := make([]byte, tcplen)
		l, err := conn.Read(first) //读取客户机发的消息
		log.Println(l)
		log.Println(first)
		if err != nil {
			log.Println(err)
			return
		}
		err,tl := GetMessageLen(first)
		if err != nil {
			log.Println(err)
			return
		}
		//解析报文头
		second := make([]byte, tl.Secondlen)
		_, err = conn.Read(second)
		err,tcpHead,pair := MsgToTcpHead(second)
		if err != nil {
			return
		}
		log.Println(tcpHead)
		log.Println(pair)

		//查找匹配mac地址
		b := IsMacRegister(&pair.Pair_mac1)
		if !b{
			log.Println("该mac未注册")
			return
		}

		//第三部分报文体
		third := make([]byte, tl.Thirdlen)
		_, err = conn.Read(third)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(third)


		SaveMacConn(conn, &pair.Pair_mac1)
		mac = pair.Pair_mac1
		break;
	}



	//消息处理循环
	for {

		//获取报文体的长度
		first := make([]byte, tcplen)
		l, err := conn.Read(first) //读取客户机发的消息
		log.Println(l)
		log.Println(first)
		if err != nil {
			log.Println(err)
			break;
		}
		err,tl := GetMessageLen(first)
		if err != nil {
			log.Println(err)
			break;
		}
        //解析报文头
		second := make([]byte, tl.Secondlen)
		_, err = conn.Read(second)
        err,tcpHead,pair := MsgToTcpHead(second)
		if err != nil {
			break;
		}
		log.Println(tcpHead)
		log.Println(pair)

		//查找匹配mac地址
		b := FindPair(pair)
        if !b{
			log.Println("该链接未注册")
			break;
		}

		//第三部分报文体
		third := make([]byte, tl.Thirdlen)
		_, err = conn.Read(third)
		if err != nil {
			break;
		}
		log.Println(third)

		// 转发消息到对应的客户端
		conn_des  := FindConn(&pair.Pair_mac2)
		if conn_des == nil{
           log.Println("对方不在线")
           break;
		}

		var desmsg []byte
		desmsg = append(desmsg,first...)
		desmsg = append(desmsg,second...)
		desmsg = append(desmsg,third...)
		_, err = conn_des.Write(desmsg)
		if err != nil {
			log.Println(err)
			break;
		}
		//转发结束
	}
}


func init() {
	repeater_map = make(map[string]string)
	repeater_register_map = make(map[string]string)
	conn_map = make(map[string]net.Conn)
	ConfigLoad()
	ConfigToMap()
}

func ConfigReload(){




}

type ConfigResult struct {
	XMLName xml.Name `xml:"message_repeater"`
	Repeater_max_number int `xml:"repeater_max_number"`
	Mrs []Repeater_pair `xml:"repeater_pair"`
}

type Repeater_pair struct {
	Pair_mac1 string `xml:"pair1_mac1"` // 解析的时候默认 SrcMac 是Pair_mac1
	Pair_mac2 string `xml:"pair1_mac2"` // 解析的时候默认 DesMac 是Pair_mac2
}


func ConfigLoad() {
	content, err := ioutil.ReadFile("E:\\go\\message_repeater\\config.xml")
	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(content, &CR)
	if err != nil {
		log.Fatal(err)
	}
}

func ConfigToMap(){
	repeater_map_lock.Lock()
	defer repeater_map_lock.Unlock()
	repeater_register_map_lock.Lock()
	defer repeater_register_map_lock.Unlock()

	for _,v := range CR.Mrs{
		repeater_register_map[v.Pair_mac1] = ""
		repeater_register_map[v.Pair_mac2] = ""
		r := strings.Compare(v.Pair_mac1,v.Pair_mac2)
		if  r < 0{
		   repeater_map[v.Pair_mac1+v.Pair_mac2] = ""
		}else {
			repeater_map[v.Pair_mac2+v.Pair_mac1] = ""
		}
	}
}

func IsMacRegister(mac *string)(bool){
	repeater_register_map_lock.Lock()
	defer repeater_register_map_lock.Unlock()
	_,exist := repeater_register_map[*mac]
	if exist{
		return true
	}else{
		return false
	}
}

func FindPair(p *Repeater_pair)(b bool){
	repeater_map_lock.Lock()
	defer repeater_map_lock.Unlock()
	r := strings.Compare(p.Pair_mac1,p.Pair_mac2)
	var s string
	if  r < 0{
		s = p.Pair_mac1+p.Pair_mac2
	}else {
		s = p.Pair_mac2+p.Pair_mac1
	}
	_, exist := repeater_map[s]
    if exist{
    	return true
	}
	return false;
}

type TcpFirst struct{
    Secondlen int
    Thirdlen int
}
type TcpSecond struct{
	SrcMac string `json:"src_mac"`
	DesMac string `json:"des_mac"`
}
type TcpThird struct{
	body []byte
}


func GetMessageLen(tl []byte)(error,TcpFirst){
	t := string(0)
	ls := string(tl)

	hls := ls[:4]
	bls := ls[4:]
	var tf TcpFirst

	pos := strings.IndexAny(hls,t)
	var l int
	var err error
	if pos == -1{
		l ,err = strconv.Atoi(string(hls))
	}else {
		l ,err =strconv.Atoi(string(hls[:pos]))
	}
	if err!= nil{
		return err,tf
	}
	tf.Secondlen = l


	pos = strings.IndexAny(bls,t)
	if pos == -1{
		l ,err = strconv.Atoi(string(bls))
	}else {
		l ,err =strconv.Atoi(string(bls[:pos]))
	}
	if err!= nil{
		return err,tf
	}
	tf.Thirdlen = l

	return nil,tf
}

func MsgToTcpHead(s []byte)(error,*TcpSecond,*Repeater_pair){
	var datap Repeater_pair
	var data TcpSecond

	if err := json.Unmarshal(s,&data); err != nil {
		log.Println("Repeater_pair 解析错误"+err.Error())
		return err,nil,nil
	}
	datap.Pair_mac1 = data.SrcMac
	datap.Pair_mac2 = data.DesMac

	return nil,&data,&datap
}

func TcpBodyToMsg(tb*TcpSecond)(error,[]byte){
	if bt,err := json.Marshal(tb); err != nil{
		log.Println("TcpBodyToMsg 解析错误"+err.Error())
        return err,nil
	}else{
		first := make([]byte,tcplen)
		copy(first ,[]byte(strconv.Itoa(len(bt))))
		var btr []byte
		btr = append(btr,bt...)
        return nil,btr
	}
}

func FindConn(mac *string)(conn net.Conn){
	conn_map_lock.Lock()
	conn_map_lock.Unlock()
	c,exist := conn_map[*mac]
	if exist{
		return c
	}else{
		return nil
	}
}

func SaveMacConn(conn net.Conn,mac *string){
	conn_map_lock.Lock()
	conn_map_lock.Unlock()
	conn_map[*mac] = conn
}

func RemoveMacConn(mac *string){
	conn_map_lock.Lock()
	conn_map_lock.Unlock()
	delete(conn_map,*mac)
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
		Name:        "asd", //服务显示名称
		DisplayName: "asd", //服务名称
		Description: "asd", //服务描述
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


