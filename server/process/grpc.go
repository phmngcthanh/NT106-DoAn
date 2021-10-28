package process

import (
	Api "Tiamat/grpc"
	"Tiamat/log"
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

func SocketServer(s string) {
	// Listen for incoming connection.

	listener, err := net.Listen("tcp4", s)
	if log.CheckErr("Error listening", err) {
		return
	}
	// Close the listener when the app closes.
	log.Log.Info("Listening local server on " + s)
	MakingGrpcServer(listener)
	listener.Close()
}

//=======================================================================================================
type handleRequestServer struct {
	Api.UnimplementedHandleRequestServer
}

//=======================================================================================================
/* Making  Server */

func NewHandleRequestServer() *handleRequestServer {
	s := &handleRequestServer{}
	return s
}

func MakingGrpcServer(listener net.Listener) {
	grpcServer := grpc.NewServer()
	Api.RegisterHandleRequestServer(grpcServer, NewHandleRequestServer())
	grpcServer.Serve(listener)
}
func (entity *handleRequestServer) Chatting(ctx context.Context, recv *Api.Mess) (*Api.Mess, error) {
	var (
		dataSend = &Api.Mess{
			Type_: "0",
			Data:  "0",
		}
	)

	if recv.Type_ == "1" { //Khai báo tạo kết nối vs server
		ServerFriend[recv.Data] = recv.Data
		log.Log.Info("New server online and request conection: ", recv.Data)
		return dataSend, nil
	}
	if recv.Type_ == "2" { //Request Update Room - add User // lưu ý, ko broad cast thông báo vì việc này server phát lệnh đã làm
		AllRooms.mux.Lock()
		defer AllRooms.mux.Unlock()
		if _, ok := AllRooms.room[recv.Roomid]; ok == false {
			return dataSend, nil
		}
		user := LoadUser(recv.User)
		AllRooms.room[recv.Roomid].Members[recv.User] = &user
	}
	if recv.Type_ == "3" { //Request Update Room - remove User // lưu ý, ko broad cast thông báo vì việc này server phát lệnh đã làm
		AllRooms.mux.Lock()
		defer AllRooms.mux.Unlock()
		if _, ok := AllRooms.room[recv.Roomid]; ok == false {
			return dataSend, nil
		}
		delete(AllRooms.room[recv.Roomid].Members, recv.User)
	}
	if recv.Type_ == "4" { // giữ kết nối vs user?
		_, ok := AllClients.client[recv.User]
		if ok {
			dataSend.Data = "1"
			return dataSend, nil
		}
	}

	if recv.Type_ == "5" { // gửi data cho user
		AllClients.client[recv.User].outgoing <- recv.Data
	}
	if recv.Type_ == "6" { // update user data
		AllClients.client[recv.User].data = LoadUser(recv.User)
	}
	return dataSend, nil
}
func BroadCastServer(type_ string, roomid string, user string, content string) {
	var (
		data_send = &Api.Mess{
			Type_:  type_,
			Roomid: roomid,
			User:   user,
			Data:   content,
		}
	)
	for addr, _ := range ServerFriend {
		conn, err := grpc.Dial(addr,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(10*time.Second),
		)
		if log.CheckErr("Bug when connect to server: "+addr+" bug:", err) {
			break
		}
		defer conn.Close()
		client := Api.NewHandleRequestClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		recv, err := client.Chatting(ctx, data_send)
		if type_ == "4" && recv.Data == "1" {
			data_send.Type_ = "5"
			data_send.Data = content
			_, _ = client.Chatting(ctx, data_send)
			return
		}
		log.CheckErr("Bug when connect to server: "+addr+" bug:", err)

	}
	return
}
