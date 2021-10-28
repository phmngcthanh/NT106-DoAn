package process

import (
	"Tiamat/Model"
	"Tiamat/log"
	"Tiamat/prisma"
	"bufio"
	"bytes"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/smtp"
)

func (client *Client) ProcessCommand(data string) {
	var (
		recv   Model.Object
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)
	recv, err := DataRecvToStructJson(data)
	sendFeedback, _ := DataSendToStringJson(object)
	if err != 0 {
		client.outgoing <- sendFeedback
		return
	}
	log.Log.Info(client.data.Name + " send CommandCode: " + recv.Commandata.Code)
	switch recv.Commandata.Code {
	case "":
	case "3":
		CreateRoom(client, recv)
	case "4":
		JoinRoom(client, recv)
	case "5":
		SendChat(client, recv)
	case "9":
		LogOut(client)
	case "10":
		ExitRoom(client, recv)
	case "11":
		ChangePass(client, recv)
	case "12":
		ChangeProfile(client, recv)
	default:
		client.outgoing <- sendFeedback
	}
}

func AcceptConnect(connection net.Conn) {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	var (
		recv   Model.Object
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)
	data, err := Read(reader)
	recv, erro := DataRecvToStructJson(data)
	sendFeedback, _ := DataSendToStringJson(object)
	log.Log.Info(connection.RemoteAddr().String() + " send CommandCode: " + recv.Commandata.Code)
	if erro != 0 || err != nil || (recv.Commandata.Code != "1" && recv.Commandata.Code != "2") {
		log.Log.Info(connection.RemoteAddr().String() + " Denied conection")
		Write(writer, sendFeedback)
		connection.Close()
		return
	}
	switch recv.Commandata.Code {
	case "":
	case "1":
		Login(connection, recv)
	case "2":
		Register(connection, recv)
	default:
		Write(writer, sendFeedback)
		connection.Close()
	}
}

func Login(connection net.Conn, recv Model.Object) {
	writer := bufio.NewWriter(connection)
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "0",
			},
		}
	)
	if prisma.Login(recv.UserData.Name, recv.UserData.Pass) != 0 {
		object.Commandata.Status = "2"
		send, _ := DataSendToStringJson(object)
		Write(writer, send)
		connection.Close()
		return
	}

	if NewClient(connection, recv.Commandata, recv.UserData) != 0 {
		object.Commandata.Status = "3"
		send, _ := DataSendToStringJson(object)
		Write(writer, send)
		connection.Close()
	}
}
func LogOut(client *Client) {
	client.Dispose()
}
func Register(connection net.Conn, recv Model.Object) {
	writer := bufio.NewWriter(connection)
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "0",
			},
		}
	)
	log.Log.Println(connection.RemoteAddr().String() + " Register and Close Connection")
	if recv.UserData.Name == "" || recv.UserData.Pass == "" {
		object.Commandata.Status = "1"
		sendFeedback, _ := DataSendToStringJson(object)
		Write(writer, sendFeedback)
		connection.Close()
		return
	}
	if prisma.Register(recv.UserData.Name, recv.UserData.Pass, recv.UserData.Email) != 0 {
		object.Commandata.Status = "1"
		sendFeedback, _ := DataSendToStringJson(object)
		Write(writer, sendFeedback)
		connection.Close()
		return
	}
	object.Commandata.Status = "0"
	sendFeedback, _ := DataSendToStringJson(object)
	Write(writer, sendFeedback)
	connection.Close()
	if recv.UserData.Email != "" {
		go SendMail(recv.UserData.Email, recv.UserData.Name)
	}
}

func CreateRoom(client *Client, recv Model.Object) {
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "0",
			},
		}
	)
	RoomID := Random(6)
	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	for _, ok := AllRooms.room[RoomID]; ok == true; { // tạo id chưa có
		RoomID = Random(6)
	}
	log.Log.Println(client.data.Name + " create room " + RoomID)
	if prisma.CreateRoom(recv.RoomData.Name, RoomID, client.data.Name) != 0 {
		object.Commandata.Status = "1"
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		log.Log.Println(client.data.Name + " CreateRoom fail to add to database")
		return
	}
	AllRooms.room[RoomID] = &Rooms{
		RoomID:   RoomID,
		RoomName: recv.RoomData.Name,
		Members:  make(map[string]*Model.User),
	}
	AllRooms.room[RoomID].Members[client.data.Name] = &Model.User{ // data ko được share chung với biến allclients
		Name:  client.data.Name,
		Alias: client.data.Alias,
		Email: client.data.Email,
	}
	object.Commandata.Status = "0" // send feedback to client that they has created successfully
	object.RoomData = Model.Room{
		ID:   RoomID,
		Name: recv.RoomData.Name,
		Members: []Model.User{
			Model.User{
				Name:  client.data.Name,
				Alias: client.data.Alias,
				Email: client.data.Email,
			},
		},
	}
	sendFeedback, _ := DataSendToStringJson(object)
	client.outgoing <- sendFeedback
	log.Log.Println(client.data.Name + " create room oke")

	object = Model.Object{ // send feedback to all client in room that they have new friend to join
		Commandata: Model.Command{
			Code:   "6",
			ID:     "0000",
			Status: "0",
		},
		RoomData: Model.Room{
			ID: RoomID,
		},
		MessData: Model.Message{
			Data:   "Wellcome " + client.data.Alias + " ^.^",
			Type:   "1",
			Author: "BoxBox",
		},
	}
	sendFeedback, _ = DataSendToStringJson(object)
	go SendChatToAll(RoomID, sendFeedback)
}

func JoinRoom(client *Client, recv Model.Object) {
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)

	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	if _, ok := AllRooms.room[recv.RoomData.ID]; ok == false {
		AllRooms.mux.Unlock()
		if _, ok := LoadRoom(recv.RoomData.ID); ok != 0 { // check trong db, nếu có trong db thì add thêm vào
			sendFeedback, _ := DataSendToStringJson(object)
			client.outgoing <- sendFeedback
			AllRooms.mux.Lock()
			return
		}
		AllRooms.mux.Lock()
	}
	if prisma.JoinRoom(recv.RoomData.ID, client.data.Name) != 0 {
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		return
	}
	BroadCastServer("2", recv.RoomData.ID, client.data.Name, "") // update RoomID cho toàn bộ server

	log.Log.Println(client.data.Name + " JoinRoom " + recv.RoomData.ID)
	AllRooms.room[recv.RoomData.ID].Members[client.data.Name] = &Model.User{ // data ko được share chung với biến allclients
		Name:  client.data.Name,
		Alias: client.data.Alias,
		Email: client.data.Email,
	}

	object.Commandata.Status = "0" // send feedback to client that they has created successfully
	object.RoomData = Model.Room{
		ID:   recv.RoomData.ID,
		Name: AllRooms.room[recv.RoomData.ID].RoomName,
		Members: func() []Model.User {
			res := make([]Model.User, 0)
			for _, user := range AllRooms.room[recv.RoomData.ID].Members {
				res = append(res, *user)
			}
			return res
		}(),
	}
	sendFeedback, _ := DataSendToStringJson(object)
	client.outgoing <- sendFeedback
	log.Log.Println(client.data.Name + " join room oke")

	object = Model.Object{ // send feedback to all client in room that they have new friend to join
		Commandata: Model.Command{
			Code:   "6",
			ID:     "0000",
			Status: "0",
		},
		RoomData: Model.Room{
			ID: recv.RoomData.ID,
		},
		MessData: Model.Message{
			Data:   "Wellcome " + client.data.Alias + " ^.^",
			Type:   "1",
			Author: "BoxBox",
		},
	}
	sendFeedback, _ = DataSendToStringJson(object)
	go SendChatToAll(recv.RoomData.ID, sendFeedback)
}
func ExitRoom(client *Client, recv Model.Object) {
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)
	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	if _, ok := AllRooms.room[recv.RoomData.ID]; ok == false {
		object.Commandata.Status = "1"
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		return
	}
	if prisma.ExitRoom(recv.RoomData.ID, client.data.Name) != 0 {
		object.Commandata.Status = "1"
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		return
	}
	log.Log.Println(client.data.Name + " ExitRoom " + recv.RoomData.ID)
	delete(AllRooms.room[recv.RoomData.ID].Members, client.data.Name)
	object.Commandata.Status = "0"
	sendFeedback, _ := DataSendToStringJson(object)
	client.outgoing <- sendFeedback

	object = Model.Object{ // send feedback to all client in room that they have new friend to join
		Commandata: Model.Command{
			Code:   "6",
			ID:     "0000",
			Status: "0",
		},
		RoomData: Model.Room{
			ID: recv.RoomData.ID,
		},
		MessData: Model.Message{
			Data:   "Goodbye " + client.data.Alias + " ＞﹏＜",
			Type:   "1",
			Author: "BoxBox",
		},
	}
	sendFeedback, _ = DataSendToStringJson(object)
	go SendChatToAll(recv.RoomData.ID, sendFeedback)
}
func SendChat(client *Client, recv Model.Object) {
	object := Model.Object{ // send feedback to all client in room that they have new friend to join
		Commandata: Model.Command{
			Code:   "6",
			ID:     "0000",
			Status: "0",
		},
		RoomData: Model.Room{
			ID: recv.RoomData.ID,
		},
		MessData: Model.Message{
			Data:   recv.MessData.Data,
			Type:   recv.MessData.Type,
			Author: client.data.Alias,
		},
	}
	sendFeedback, _ := DataSendToStringJson(object)
	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	if _, ok := AllRooms.room[recv.RoomData.ID]; ok == false {
		return
	}
	for user_name, _ := range AllRooms.room[recv.RoomData.ID].Members {
		if user_name == client.data.Name {
			continue
		}
		AllClients.mux.Lock()
		if user, oke := AllClients.client[user_name]; oke == false || (oke == true && user.status == "dead") {
			BroadCastServer("4", "", user_name, sendFeedback)
			AllClients.mux.Unlock()
			continue
		}
		user := AllClients.client[user_name]
		AllClients.mux.Unlock()
		user.outgoing <- sendFeedback
	}
}
func SendChatToAll(RoomID string, data string) {
	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	for user_name, _ := range AllRooms.room[RoomID].Members {
		AllClients.mux.Lock()
		if user, oke := AllClients.client[user_name]; oke == false || (oke == true && user.status == "dead") {
			BroadCastServer("4", "", user_name, data)
			AllClients.mux.Unlock()
			continue
		}
		user := AllClients.client[user_name]
		AllClients.mux.Unlock()
		user.outgoing <- data
	}
}
func ChangePass(client *Client, recv Model.Object) {
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)
	if EditPass(client.data.Name, recv.UserData.Pass) != 0 {
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		return
	}
	object.Commandata.Status = "0"
	sendFeedback, _ := DataSendToStringJson(object)
	client.outgoing <- sendFeedback
}

func ChangeProfile(client *Client, recv Model.Object) {
	var (
		object = Model.Object{
			Commandata: Model.Command{
				Code:   recv.Commandata.Code,
				ID:     recv.Commandata.ID,
				Status: "1",
			},
		}
	)
	if EditProfile(client.data.Name, recv.UserData.Alias, recv.UserData.Email, recv.UserData.Ava) != 0 {
		sendFeedback, _ := DataSendToStringJson(object)
		client.outgoing <- sendFeedback
		return
	}
	object.Commandata.Status = "0"
	sendFeedback, _ := DataSendToStringJson(object)
	client.outgoing <- sendFeedback
}

func SendMail(UserMail string, UserName string) {
	from_email := "19520958@gm.uit.edu.vn"
	password := "@@@@@@@@@@@"
	host := "smtp.gmail.com:587"
	auth := smtp.PlainAuth("", from_email, password, "smtp.gmail.com")
	header := make(map[string]string)
	to_email := UserMail
	header["From"] = from_email
	header["To"] = to_email
	header["Subject"] = "Welcome to HeyChat!!"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", "text/html")
	header["Content-Disposition"] = "inline"
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header_message := ""
	for key, value := range header {
		header_message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	body := "<a>Hello " + UserName + " ( •̀ ω •́ )✧ .</br>Let HeyChat say welcome to you (～￣▽￣)～(～￣▽￣)～</a>"
	var body_message bytes.Buffer
	temp := quotedprintable.NewWriter(&body_message)
	temp.Write([]byte(body))
	temp.Close()
	final_message := header_message + "\r\n" + body_message.String()
	status := smtp.SendMail(host, auth, from_email, []string{to_email}, []byte(final_message))
	if status != nil {
		log.Log.Warn("Error from SMTP Server: %s", status)
	}
}
