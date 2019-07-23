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

type user_logstate struct {
	password string
	islogined bool
}
//若之后建数据库的话这些都应该写在文件中
var (
	user_login map[string]user_logstate
	cont_log map[string][]byte//暂时未初始化
)
//关于信息的解码问题：统一采用两位字母表示。
//RG：register
//LI:login
//IP:chooseip
var mainwin_record []net.Conn
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
	mainwin_record = make([]net.Conn,0,3)//这里暂时设定为5
	//exit when server listen fail
	if err != nil {
		PrintErr(err.Error())
	} else {
		PrintLog("Start Listen " + server.listenAddr)
	}

	//init
	user_login = make(map[string]user_logstate)
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
	//注意！下面这一句只是不知道暂时放哪里而已。
	//应该放在mainwin里面，仅仅有它涉及广播。
	//mainwin_record = append(user_record,client)
	//TODO:Register(chosen)
	//TODO:warning of same name.

	//the first time
	readSize_rl, readError_rl := client.Read(buffer)//the first time
	if readError_rl != nil {
		PrintErr(clientAddr.String() + " fail")
		client.Close()
	}else{
		msg = string(buffer[0:readSize_rl])
		//Decode
		msg_type = msg[0:2]
		switch msg_type {
		//warning:it should be in the endless loop.
		case "IP":
			//未来根据需要可以考虑加更多的东西。
			client.Write([]byte("Success"));
			//TODO: 仅仅需要接受窗口关闭信息即可。
			for {
				readSize_rl, readError_rl = client.Read(buffer)
				if readError_rl != nil {
					PrintErr(clientAddr.String() + " fail")
					client.Close()
					break;
				}
			}
		case "RG":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]

			user_login[user_name] = user_logstate{tmppassword,false}
			PrintRegister(user_name,tmppassword)

			client.Write([]byte("Success"))
		case "LI":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]

			truepassword,user_exist := user_login[user_name]

			if(user_exist&&truepassword.password == tmppassword){//success
			 	tmpls := user_logstate{user_login[user_name].password,true}
				user_login[user_name] = tmpls
				client.Write([]byte("Success"))
			}else{//fail
				if(!user_exist){//用户不存在
					client.Write([]byte("Notexist"))
				} else {//密码错误
					client.Write([]byte("WrongPassword"))
				}

				for{
					readSize_rl, readError_rl = client.Read(buffer)
					if readError_rl != nil {
						PrintErr(clientAddr.String() + " fail")
						client.Close()
						break;
					}else{
						msg = string(buffer[0:readSize_rl])
						user_name = msg[0:strings.Index(msg," ")]
						tmppassword = msg[strings.Index(msg," ") + 1:]

						truepassword,user_exist = user_login[user_name]
						if(user_exist&&truepassword.password == tmppassword){//success
							tmpls := user_logstate{user_login[user_name].password,true}
							user_login[user_name] = tmpls
							client.Write([]byte("Success"))
							break;
						}else{//fail
							if(!user_exist){//用户不存在
								client.Write([]byte("Notexist"))
							} else {//密码错误
								client.Write([]byte("WrongPassword"))
							}
						}
					}
				}
			}
		case "MW":

		default:

		}
	}
	//TODO:Login(including visitor)

	//TODO:广播上线信息
	if runtime.NumGoroutine() > 1{
		for i := range mainwin_record{
			if mainwin_record[i] != client{
				mainwin_record[i].Write([]byte("a new user called    log in"))
			}else{
				mainwin_record[i].Write([]byte("You have successfully log in!"))
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
