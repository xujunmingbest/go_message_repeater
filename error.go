package main


type resError struct{
    ErrMes string `json:"err_mes"`
    ErrCode int `json:"err_code"`
}

var resErrors = []resError{
	{"正确",0},
	resError{"对方不存在",1},
	{"对方不在线",2},
	{"你和对方的链路未注册",3},
	{"您的mac机器未注册",4},
	{"有其他mac机子登入了",5},
	{"报文长度位接收解析错误，连接中断",6},
	{"报文Json解析错误",7},
	{"装发器内部错误",8},
}


