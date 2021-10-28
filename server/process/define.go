package process

import (
	"Tiamat/Model"
	"Tiamat/log"
	"bufio"
	crypto "crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net"
	"sync"
	"time"
)

//========================================================================================================
/* Global variables*/
type Rooms struct {
	RoomID   string
	RoomName string
	Members  map[string]*Model.User
}
type AllRoom struct { // distinguish by RoomID - 5 byte
	room map[string]*Rooms
	mux  sync.Mutex
}
type AllClient struct { // distinguish by UserName
	client map[string]*Client
	mux    sync.Mutex
}

var AllRooms AllRoom
var AllClients AllClient
var ServerFriend map[string]string

type Client struct {
	outgoing  chan string
	reader    *bufio.Reader
	writer    *bufio.Writer
	conn      net.Conn
	data      Model.User
	status    string
	key       string
	keystream crypto.Stream
}

//========================================================================================================
/*Init Function  && Support Function*/
func InitSocket() {
	AllClients.client = make(map[string]*Client)
	AllRooms.room = make(map[string]*Rooms)
	ServerFriend = make(map[string]string)

}
func Random(length int) string {
	rand.Seed(time.Now().UnixNano())
	char := "qwertyuioplkjhgfdsazxcvbnm1234567890"
	id := ""
	for i := 0; i < length; i++ {
		id = id + string(char[rand.Intn(len(char))])
	}
	return id
}

func RemoveLastSign(s string) string {
	if last := len(s) - 1; last >= 0 && s[last] == '\x00' {
		return s[:last]
	}
	return s
}

func Hash256(msg string) string {
	tmp := sha256.Sum256([]byte(msg))
	msg256 := hex.EncodeToString(tmp[:])
	return string(msg256)
}

func DataSendToStringJson(object Model.Object) (string, error) { // data will be sent, been encrypted here
	send, err := json.Marshal(object)
	if err != nil {
		log.Log.Error("Parse to json string faild... :", err)
		return "", err
	}
	return string(send), nil
}
func DataRecvToStructJson(recv string) (Model.Object, int) { // recv datastring, return 2 struct, decrypt here

	data := recv
	object := Model.Object{}
	json.Unmarshal([]byte(data), &object)
	return object, 0
}
