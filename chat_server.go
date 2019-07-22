package chatroom

import (
	"net"
	"strconv"
	"runtime"
	"strings"
	)

const (
	LISTEN_TCP = "tcp"
	PING_MSG   = "receive connection from "
)

//若之后建数据库的话这些都应该写在文件中
var (
	password map[string]string
	cont_log map[string][]byte//暂时未初始化
)
//关于信息的解码问题：统一采用两位字母表示。
//RG：register
//LI:login
var user_record []net.Conn
//data structure of server
type ChatServer struct {
	listenAddr string
	status     bool
	listener   net.Listener
}

//create a new server, you should explain why we do this
func NewChatServer(addr string, port int) *ChatServer {
	server := new(ChatServer)
	server.listenAddr = addr + ":" + strconv.Itoa(port)
	return server
}

//main server function
func (server ChatServer) StartListen() {
	//start listen on the address given
	listener, err := net.Listen(LISTEN_TCP, server.listenAddr)
	server.listener = listener
	user_record = make([]net.Conn,0,5)//这里暂时设定为5
	//exit when server listen fail
	if err != nil {
		PrintErr(err.Error())
	} else {
		PrintLog("Start Listen " + server.listenAddr)
	}

	//init
	password = make(map[string]string)
	//main server loop, you should explain how this server loop works
	for {
		client, acceptError := server.listener.Accept() //when a user comes in
		if acceptError != nil {
			PrintErr(acceptError.Error()) //if accept go wrong, then the server will exit
			break
		} else {
			go server.userHandler(client) //then create a coroutine go process the user (What is coroutine? What's the function of keyword 'go'?)
		}
	}
}

func (server ChatServer) userHandler(client net.Conn) {
	buffer := make([]byte, 1024)      //create a buffer
	clientAddr := client.RemoteAddr() //get user's address
	// 记得设置游客常量
	var (
		msg string
		msg_type string
		user_name string
		tmppassword string
	)
	user_record = append(user_record,client)
	//TODO:Register(chosen)
	readSize_rl, readError_rl := client.Read(buffer)
	if readError_rl != nil {
		PrintErr(clientAddr.String() + " fail")
		client.Close()
	}else{
		msg = string(buffer[0:readSize_rl])
		//Decode
		msg_type = msg[0:2]
		switch msg_type {
		case "RG":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]
			
			password[user_name] = tmppassword
			PrintRegister(user_name,tmppassword)
		case "LI":
			
			
		default:


		}
	}
	//TODO:Login(including visitor)

	//TODO:广播上线信息
	if runtime.NumGoroutine() > 1{
		for i := range user_record{
			if user_record[i] != client{
				user_record[i].Write([]byte("a new user called    log in"))
			}else{
				user_record[i].Write([]byte("You have successfully log in!"))
			}
		}
	}

	PrintClientMsg(PING_MSG + clientAddr.String()) //print a log to show that a new client comes in
	for {                                          //main serve loop
		readSize, readError := client.Read(buffer) //why we can put a read in for loop?

		if readError != nil {
			PrintErr(clientAddr.String() + " fail") //if read error occurs, close the connection with user
			client.Close()
			break
		} else {
			msg = string(buffer[0:readSize])                //or convert the byte stream to a string
			//设想中的msg应该有如下几种：
			//user_name == tmp_name:send and print to all
			//此功能可以遍历流程实现（chosen），也可进入其他线程
			PrintClientMsg(clientAddr.String() + ":" + msg) //then print the message
			client.Write(buffer[0:readSize])                //send the msg back to user

		}
	}
}
