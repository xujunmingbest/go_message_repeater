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
	srcmac:=""
	desmac:=""
	defer fmt.Println("一个连接退出")
	defer conn.Close();
	defer RemoveMacConn(&mac,conn)
	defer log.Println(conn_map)
	//预先接收 mac 地址
	//装载 mac和conn 到map
	/************************
    查找mac地址






	*************************/
	for {
		//获取报文体的长度
		err,first := GetMsg(conn,tcplen)
		if err != nil {
			log.Println(err)
			return
		}
		err,tl := GetMessageLen(first)
		if err != nil {
			SendErr(6,&srcmac,&desmac,conn)
			log.Println(err)
			return
		}
		//解析报文头
		err,second := RecvHead(conn,tl.Secondlen)
		if err != nil {
			log.Println("接收报文头出错")
			return
		}
		err,_,pair := MsgToTcpHead(second)
		if err != nil {
			SendErr(7,&srcmac,&desmac,conn)
			return
		}

		//查找匹配mac地址
		b := IsMacRegister(&pair.Pair_mac1)

		if !b{
			SendErr(4,&pair.Pair_mac1,&pair.Pair_mac1,conn)
			return
		}
		//检测是不是登入了
		checkLogin(&pair.Pair_mac1)

		//第三部分报文体
		err,_= GetMsg(conn,tl.Thirdlen)
		if err != nil {
			log.Println(err)
			return
		}

		SaveMacConn(conn, &pair.Pair_mac1)
		mac = pair.Pair_mac1
		break;
	}



	//消息处理循环
	for {

		//获取报文体的长度
		err,first := GetMsg(conn,tcplen)
		if err != nil {
			log.Println(err)
			return
		}
		err,tl := GetMessageLen(first)
		if err != nil {
			SendErr(6,&srcmac,&desmac,conn)
			log.Println(err)
			return
		}
		//解析报文头
		err,second := RecvHead(conn,tl.Secondlen)
		if err != nil {
			log.Println("接收报文头出错")
			return
		}
		err,_,pair := MsgToTcpHead(second)
		if err != nil {
			log.Println("解析报文头出错 错误返回客户端")
			return
		}

		//第三部分报文体
		err,third := GetMsg(conn,tl.Thirdlen)
		if err != nil {
			log.Println(err)
			return
		}

		//查找匹配mac地址
		b := FindPair(pair)
		if !b{
			SendErr(3,&pair.Pair_mac1,&pair.Pair_mac2,conn)
			continue
		}


		// 转发消息到对应的客户端
		conn_des  := FindConn(&pair.Pair_mac2)
		if conn_des == nil{
			SendErr(2,&pair.Pair_mac1,&pair.Pair_mac2,conn)
			continue;
		}

		desmsg := GenerateMsg(second,third)
		_, err = conn_des.Write(desmsg)
		if err != nil {
			log.Println(err)
			SendErr(8,&pair.Pair_mac1,&pair.Pair_mac2,conn)
			return
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

func GetMsg(conn net.Conn,len int)(error,[]byte){
	err,msg:= NetRecv(conn,len)  //读取客户机发的消息
    return err,msg
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



func RecvHead(conn net.Conn,len int)(error,[]byte){
	err,second:= NetRecv(conn,len)

	enc := mahonia.NewDecoder("gb18030")
	secondt := enc.ConvertString(string(second))
	if err != nil {
		return err,nil
	}else{
		return err,[]byte(secondt)
	}

}

//将接收到的消息转化为 文件头
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

func GenerateMsg(second []byte,third  []byte)([]byte){
	fl := make([]byte,4)
	copy(fl ,[]byte(strconv.Itoa(len(second))))
	fs := make([]byte,8)
	copy(fs ,[]byte(strconv.Itoa(len(third))))
    var finall []byte
	finall = append(finall,fl...)
	finall = append(finall,fs...)
	finall = append(finall,second...)
	finall = append(finall,third...)
	return finall
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

		log.Println(string(bt))
		_, err =c.Write([]byte(btr))
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
}




/**************************
介绍:
鉴于 conn.read 函数 有时候会读不全 特意封装
循环读取
*****************************/
func NetRecv(c net.Conn,readl int )(error,[]byte ){
	readbuff := make([]byte,readl)

	Readedlen := 0
     for Readedlen < readl{
		 n,err := c.Read(readbuff[Readedlen:])
		 if err!=nil{
		 	return err,nil
		 }
		 Readedlen += n;
	 }
     return nil,readbuff
}