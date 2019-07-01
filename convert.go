package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
)


// 信息交换
type Mess struct{
	MessType int
	LinkNode string
}



var linkMap map[string] net.Conn

var local string

func main(){
	if len(os.Args) != 4{
		os.Exit(1)
	}
	if os.Args[1]=="cli" {
		cli()
	} else if os.Args[1]== "server" {
		server()
	}
}

// 服务端主程序
func server() {
	forwardAddr := os.Args[2]
	linkAddr := os.Args[3]
	linkMap=make(map[string]net.Conn)
	linkSock, err := net.Listen("tcp4", linkAddr)
	ckErr(err)
	forwardSock, err := net.Listen("tcp4",forwardAddr)
	ckErr(err)
	linkConn,err := linkSock.Accept()
	fmt.Println("link established")
	ckErr(err)
	go dealMess(linkSock)
	defer linkSock.Close()
	defer forwardSock.Close()
	for {
		forwardConn, err :=forwardSock.Accept()
		fmt.Printf("[+]out conn from %s\n",forwardConn.RemoteAddr().String())
		if err != nil {

		}
		go forward(forwardConn,linkConn)
	}
}

// 客户端主程序
func cli(){
	buf:=make([]byte,10240)
	server := os.Args[2]
	local = os.Args[3]

	addr,err:=net.ResolveTCPAddr("tcp4",server)
	ckErr(err)

	conn,err:= net.DialTCP("tcp4",nil,addr)
	ckErr(err)

	defer conn.Close()
	for{

		i,err:=conn.Read(buf)
		ckErr(err)
		if i == 0 {

		}
		decoder := gob.NewDecoder(bytes.NewReader(buf))
		var mess Mess
		err=decoder.Decode(&mess)
		fmt.Println("[+]message decoded")
		fmt.Println(mess)
		ckErr(err)
		fmt.Printf("[+]recieve out conn from %s\n",mess.LinkNode)
		if mess.MessType ==1 {
			fmt.Printf("[+]recieve out conn from %s\n",mess.LinkNode)
			perConn , err := net.DialTCP("tcp4",nil,addr)
			ckErr(err)
			go handleConnect(conn,perConn,mess)
		}
	}
}


// 处理返回的链路链接，加入map中
func dealMess(listener net.Listener){
	for {
		l,err := listener.Accept()
		if err != nil {
			continue
		}
		buff:=make([]byte,1024)
		len, err := l.Read(buff)
		if err != nil {
			_ = l.Close()
			continue
		}
		if len == 0 {

		}
		decoder:=gob.NewDecoder(bytes.NewReader(buff))
		var mess Mess
		err = decoder.Decode(&mess)
		if err != nil {
			_ = l.Close()
			continue
		}
		linkMap[mess.LinkNode]=l
		_, _ = l.Write(buff)
	}
}

// 服务端：对连接请求进行初始化处理，与内网进行链路建立
func forward(conn net.Conn,linkConn net.Conn){
	mess := Mess{MessType:1,LinkNode:conn.RemoteAddr().String()}
	fmt.Println(mess)
	var buff bytes.Buffer
	encoder:=gob.NewEncoder(&buff)
	_ = encoder.Encode(mess)
	_, _ = linkConn.Write(buff.Bytes())
	buff2:=make([]byte,1024)
	i,err := linkConn.Read(buff2)
	ckErr(err)
	if i == 0 {

	}
	decoder := gob.NewDecoder(bytes.NewReader(buff2))
	ckErr(err)
	err = decoder.Decode(&mess)
	fmt.Printf("%s\n",mess.LinkNode)
	ckErr(err)
	fmt.Printf("[+]dual connection established for %s\n",mess.LinkNode)
	fmt.Println(linkMap)
	defer delete(linkMap, mess.LinkNode)
	go dealDualConn(linkMap[mess.LinkNode],conn)
	go dealDualConn(conn,linkMap[mess.LinkNode])
}


// 客户端：处理新的链路连接
func handleConnect(conn net.Conn,perConn net.Conn, mess Mess){
	mess.MessType=2
	var responce bytes.Buffer
	encoder := gob.NewEncoder(&responce)
	err := encoder.Encode(mess)
	ckErr(err)
	connBytes := responce.Bytes()
	fmt.Printf("[+]building channel for %s\n",mess.LinkNode)
	perConn.Write(connBytes)
	addr,err:=net.ResolveTCPAddr("tcp4",local)
	ckErr(err)
	buff:=make([]byte,1024)
	len,err := perConn.Read(buff)
	if len==0 {

	}
	fmt.Printf("[+]channel builded for %s\n",mess.LinkNode)
	connLocal,err:= net.DialTCP("tcp4",nil,addr)
	ckErr(err)
	_, _ = conn.Write(connBytes)
	go dealDualConn(perConn,connLocal)
	go dealDualConn(connLocal,perConn)
}

// 双向socket通道
func dealDualConn(leftConn net.Conn,rightConn net.Conn){
	if leftConn != nil {
		defer leftConn.Close()
	}
	if rightConn != nil {
		defer rightConn.Close()
	}

	if leftConn ==nil || rightConn == nil {
		return
	}
	buf:=make([]byte,10240)
	for {
		len,err := leftConn.Read(buf)
		fmt.Printf("[+]%d bytes recv\n",len)
		if err!= nil {
			break
		}
		if len!= -1 {
			_, _ = rightConn.Write(buf[:len])
		} else {
			break
		}
	}
	fmt.Println("[-]ended an transaction")
}

// 致命错误处理
func ckErr(err error){
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}