package chatroom

import (
	"net"
	"strconv"
	"strings"
	"fmt"
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
	mainwin_record map[string]net.Conn
	administor string
	isprivate bool
)
//关于信息的解码问题：统一采用两位字母表示。
//RG：register
//LI:login
//IP:chooseip
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
	//exit when server listen fail
	if err != nil {
		PrintErr(err.Error())
	} else {
		PrintLog("Start Listen " + server.listenAddr)
	}

	//init
	user_login = make(map[string]user_logstate)
	mainwin_record = make(map[string]net.Conn)
	administor = ""
	isprivate = false
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
	var (
		msg string
		msg_type string
		user_name string
		tmppassword string
		new_pswd string
	)

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
			var user_all string
			if readSize_rl == 2{
				client.Write([]byte("Success"))
				PrintConfirm()
			}else {
				PrintQuiry()
				if isprivate{
					client.Write([]byte("Private"))
				}else{
					user_all = ""
					for i := range mainwin_record{
						user_all += i + "\n"
					}
					if user_all == ""{
						user_all = "None"
					}
					client.Write([]byte(user_all))
				}
			}
			//TODO: 仅仅需要接受窗口关闭信息即可。
			for {
				readSize_rl, readError_rl = client.Read(buffer)
				if readError_rl != nil {
					PrintErr(clientAddr.String() + " fail")
					client.Close()
					break
				}else {
					msg = string(buffer[0:readSize_rl])
					if msg == "Q"{
						PrintQuiry()
						if isprivate{
							client.Write([]byte("Private"))
						}else {
							user_all = ""
							for i := range mainwin_record{
								user_all += i + "\n"
							}
							if user_all == ""{
								user_all = "None"
							}
							client.Write([]byte(user_all))
						}
					}
				}
			}
		case "RG":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]

			_,isusere := user_login[user_name]
			if !isusere {
				user_login[user_name] = user_logstate{tmppassword,false}
				client.Write([]byte("Success"))
				PrintRegister(user_name)
			}else{
				client.Write([]byte("Fail"))
				for{
					readSize_rl, readError_rl = client.Read(buffer)

					if readError_rl != nil {
						PrintErr(clientAddr.String() + " fail")
						client.Close()
						break;
					}else{
						msg = string(buffer[0:readSize_rl])
						user_name = msg[0:strings.Index(msg," ")]
						fmt.Println("name",user_name,"\n")
						tmppassword = msg[strings.Index(msg," ") + 1:]

						_,isusere = user_login[user_name]
						if(!isusere){
							user_login[user_name] = user_logstate{tmppassword,false}
							client.Write([]byte("Success"))
							PrintRegister(user_name)
							break
						}else{
							client.Write([]byte("Fail"))
						}
					}
				}
			}
		case "LI":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]

			truepassword,user_exist := user_login[user_name]

			if(user_exist&&!truepassword.islogined&&truepassword.password == tmppassword){//success
			 	tmpls := user_logstate{user_login[user_name].password,true}
				user_login[user_name] = tmpls
				client.Write([]byte("Success"))
				PrintLogin(user_name)
			}else{//fail
				if(!user_exist){//用户不存在
					client.Write([]byte("Notexist"))
				}else if(truepassword.islogined){//已经登陆过
					client.Write([]byte("Logined"))
				} else {//密码错误
					client.Write([]byte("WrongPassword"))
				}

				for{
					readSize_rl, readError_rl = client.Read(buffer)
					if readError_rl != nil {
						PrintErr(clientAddr.String() + " fail")
						client.Close()
						break
					}else{
						msg = string(buffer[0:readSize_rl])
						user_name = msg[0:strings.Index(msg," ")]
						tmppassword = msg[strings.Index(msg," ") + 1:]

						truepassword,user_exist = user_login[user_name]
						if(user_exist&&truepassword.password == tmppassword){//success
							tmpls := user_logstate{user_login[user_name].password,true}
							user_login[user_name] = tmpls
							client.Write([]byte("Success"))
							PrintLogin(user_name)
							break
						}else{//fail
							if(!user_exist){//用户不存在
								client.Write([]byte("Notexist"))
							}else if(truepassword.islogined){//已经登陆过
								client.Write([]byte("Logined"))
							} else {//密码错误
								client.Write([]byte("WrongPassword"))
							}
						}
					}
				}
			}
		case "MW":
			user_name = msg[2:readSize_rl]
			user_all := ""
			for i := range mainwin_record{
				mainwin_record[i].Write([]byte(GetCurrentTimeString()+"\nA new user called "+user_name+" log in"))
				user_all += i + " "
			}
			if user_all == ""{
				client.Write([]byte("You are the only one in this chatroom! \n Congratulation! You become the administrator of this chatroom! \n Invite more friends here! "))
				administor = user_name
			}else {
				user_all += "."
				client.Write([]byte(GetCurrentTimeString() + " These users are online: \n" + user_all))
			}
			mainwin_record[user_name] = client
			PrintClientMsg(PING_MSG + clientAddr.String()) //print a log to show that a new client comes in
			for {                                          //main serve loop
				readSize_rl, readError_rl = client.Read(buffer) //why we can put a read in for loop?

				if readError_rl != nil {
					PrintErr(clientAddr.String() +user_name +" log out") //if read error occurs, close the connection with user
					tmpmu := user_logstate{user_login[user_name].password,false}
					user_login[user_name] = tmpmu
					delete(mainwin_record,user_name)
					for i := range mainwin_record{
						mainwin_record[i].Write([]byte(GetCurrentTimeString()+"\n"+user_name+" log out"))
					}
					client.Close()
					break
				} else {
					msg = string(buffer[0:readSize_rl])
					msg_type = msg[0:2]
					switch msg_type {
					case "CA":
						msg = msg[2:]
						PrintClientMsg(clientAddr.String() + ":" +user_name + " "+ msg) //then print the message
						//设想中的msg应该有如下几种：
						//user_name == tmp_name:send and print to all
						//此功能可以遍历流程实现（chosen），也可进入其他线程
						for i := range mainwin_record{
							if(mainwin_record[i] != client){
								mainwin_record[i].Write([]byte(user_name+" " + GetCurrentTimeString() +  " : \n"+msg))
							}else {
								mainwin_record[i].Write([]byte("you " + GetCurrentTimeString() +  " : \n"+msg))
							}
						}
					case "CN":
						msg = msg[2:]
						mainwin_record[msg] = mainwin_record[user_name]
						delete(mainwin_record,user_name)
						PrintClientMsg(clientAddr.String() + ":" +user_name + " Change his/her name to "+ msg) //then print the message
						for i := range mainwin_record{
							if(mainwin_record[i] != client){
								mainwin_record[i].Write([]byte(GetCurrentTimeString()+":\n"+user_name+" changes his/her name to "+msg))
							}
						}
						user_name = msg
					case "PP":
						if msg[2:] == "Public"{
							isprivate = false
							PrintClientMsg(clientAddr.String() + ":" +user_name + " Change to public") //then print the message
						}else if msg[2:] == "Private"{
							isprivate = true
							PrintClientMsg(clientAddr.String() + ":" +user_name + " Change to private") //then print the message
						}
					}
				}
			}
		case "CN":
			user_name = msg[2:strings.Index(msg," ")]
			msg = msg[strings.Index(msg," ") + 1:]
			newname := msg[0:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:]

			_,isexist := user_login[newname]
			var tmpstat user_logstate
			if tmppassword == user_login[user_name].password&&!isexist {
				delete(user_login,user_name)
				tmpstat = user_logstate{tmppassword,true}
				user_login[newname] = tmpstat

				client.Write([]byte("Success"))
			}else{
				if tmppassword != user_login[user_name].password {
					client.Write([]byte("WrongPassword"))
				}else {
					client.Write([]byte("NameExist"))
				}

				for {
					readSize_rl, readError_rl = client.Read(buffer)

					if readError_rl != nil {
						PrintErr(clientAddr.String() + " fail")
						client.Close()
						break;
					}else{
						msg = string(buffer[0:readSize_rl])
						user_name = msg[0:strings.Index(msg," ")]
						msg = msg[strings.Index(msg," ") + 1:]
						newname := msg[0:strings.Index(msg," ")]
						tmppassword = msg[strings.Index(msg," ") + 1:]
						_,isexist = user_login[newname]

						if tmppassword == user_login[user_name].password&&!isexist {
							delete(user_login,user_name)
							tmpstat = user_logstate{tmppassword,true}
							user_login[newname] = tmpstat

							client.Write([]byte("Success"))
						}else {
							if tmppassword != user_login[user_name].password {
								client.Write([]byte("WrongPassword"))
							} else {
								client.Write([]byte("NameExist"))
							}
						}
					}
				}
			}
		case "CP":
			user_name = msg[2:strings.Index(msg," ")]
			tmppassword = msg[strings.Index(msg," ") + 1:strings.Index(msg,"~#@Password@#~")]
			new_pswd = msg[strings.Index(msg,"~#@Password@#~") + 14:]

			true_pswd := user_login[user_name]
			if(true_pswd.password == tmppassword){
				tmpls := user_logstate{new_pswd,true}
				user_login[user_name] = tmpls
				client.Write([]byte("Success"))
				PrintClientMsg(clientAddr.String() + ":" +user_name + " Change his/her password ") //then print the message
			}else{
				client.Write([]byte("Fail"))
				for{
					readSize_rl, readError_rl = client.Read(buffer)

					if readError_rl != nil {
						PrintErr(clientAddr.String() + " fail")
						client.Close()
						break;
					}else{
						user_name = msg[0:strings.Index(msg," ")]
						tmppassword = msg[strings.Index(msg," ") + 1:strings.Index(msg,"~#@Password@#~")]
						new_pswd = msg[strings.Index(msg,"~#@Password@#~") + 14:]

						true_pswd = user_login[user_name]
						if(true_pswd.password == tmppassword){
							tmpls := user_logstate{new_pswd,true}
							user_login[user_name] = tmpls
							client.Write([]byte("Success"))
							PrintClientMsg(clientAddr.String() + ":" +user_name + " Change his/her password ") //then print the message
							break
						}else {
							client.Write([]byte("Fail"))
						}
					}
				}
			}
		case "AC":
			msg_type = msg[2:8]
			switch msg_type {
			case "logout":
				towrite := ""
				for i := range mainwin_record{
					if i != msg[8:]{
						towrite += "\n"+i
					}
				}
				if(towrite == ""){
					client.Write([]byte("None"))
				}else{
					client.Write([]byte("Choose one to be the administrator :"+towrite))
					for{
						readSize_rl, readError_rl = client.Read(buffer)
						if readError_rl != nil {
							PrintErr(clientAddr.String() + " fail")
							client.Close()
							break
						}else{
							msg = string(buffer[0:readSize_rl])
							_,isexist := mainwin_record[msg]
							if isexist{
								administor = msg
								client.Write([]byte("Success"))
								mainwin_record[msg].Write([]byte("Congratulation! You become the administrator of this chatroom!"))
								break
							}else{
								client.Write([]byte("Notexist"))
							}
						}
					}
				}
			}
		default:
		}
	}

}
