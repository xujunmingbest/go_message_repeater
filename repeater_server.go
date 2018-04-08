package main

import (
	"net"
	"log"
	"fmt"
	"encoding/json"
	"strconv"
	"strings"
	"github.com/axgle/mahonia"
)

//处理客户端消息
func DoServerHandle(conn net.Conn) {
	log.Println("有客户端连接")
	mac := ""
	defer fmt.Println("一个连接退出")
	defer conn.Close();
	defer RemoveMacConn(&mac,conn)
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
			log.Println("IsMacRegister")
			SendErr(4,&pair.Pair_mac1,&pair.Pair_mac1,conn)
			return
		}
		//检测是不是登入了
		checkLogin(&pair.Pair_mac1)

		//第三部分报文体
		third := make([]byte, tl.Thirdlen)
		_, err = conn.Read(third)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("third",string(third))


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


		//第三部分报文体
		third := make([]byte, tl.Thirdlen)
		_, err = conn.Read(third)
		if err != nil {
			break;
		}
		log.Println("third",string(third))

		//查找匹配mac地址
		b := FindPair(pair)
		if !b{
			log.Println("该链接未注册")
			continue
		}


		// 转发消息到对应的客户端
		conn_des  := FindConn(&pair.Pair_mac2)
		if conn_des == nil{
			log.Println("对方不在线")
			continue;
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




//查看mac 是不是注册了
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

//查看是不是在 链路里面
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



//将接收到的消息转化为 文件头
func MsgToTcpHead(s []byte)(error,*TcpSecond,*Repeater_pair){
	var datap Repeater_pair
	var data TcpSecond

	enc := mahonia.NewDecoder("gb18030")
	sl:= enc.ConvertString(string(s))

	if err := json.Unmarshal([]byte(sl),&data); err != nil {
		log.Println("Repeater_pair 解析错误"+err.Error())
		return err,nil,nil
	}
	datap.Pair_mac1 = data.SrcMac
	datap.Pair_mac2 = data.DesMac

	return nil,&data,&datap
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

// 将登入的 mac 和 conn 存在一起
func SaveMacConn(conn net.Conn,mac *string){
	conn_map_lock.Lock()
	conn_map_lock.Unlock()
	conn_map[*mac] = conn
}

// 移除mac 对应conn
func RemoveMacConn(mac *string,conn net.Conn){
	conn_map_lock.Lock()
	conn_map_lock.Unlock()
	if conn_map[*mac] == conn {  //发现两个一致才移除
		delete(conn_map, *mac)
	}
}

//检测是不是已经登入
func checkLogin(mac *string){
	conn_map_lock.Lock()
	defer conn_map_lock.Unlock()
	c,exist := conn_map[*mac]
	if exist{
		c.Close()
	}
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

type TcpFirst struct{
	Secondlen int
	Thirdlen int
}
type TcpSecond struct{
	SrcMac string `json:"src_mac"`
	DesMac string `json:"des_mac"`
	Err resError `json:"err"`
}

type TcpThird struct{
	body []byte
}



//装载错误信息
func SendErr(errorCode int,srcMac *string, desMac *string,c net.Conn)(error){
	ts :=TcpSecond{
		*srcMac,
		*desMac,
		resErrors[errorCode],
	}

	if bt,err := json.Marshal(&ts); err != nil{
		log.Println("LoadErr 解析错误"+err.Error())
		return err
	}else{
		fls := make([]byte,4)
		sls := make([]byte,8)
		copy(fls ,[]byte(strconv.Itoa(len(bt))))
		var btr []byte

		btr = append(btr,fls...)
		btr = append(btr,sls...)
		btr = append(btr,bt...)

		_, err =c.Write([]byte(btr))
		log.Println(string(bt))
		log.Println(string(bt))
		log.Println(string(bt))
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
}

