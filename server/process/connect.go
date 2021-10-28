package process

import (
	"Tiamat/Model"
	"Tiamat/log"
	"bufio"
	"bytes"
	"net"
	"strings"
)

func NewClient(connection net.Conn, command Model.Command, userData Model.User) int {
	var client *Client
	AllClients.mux.Lock()

	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	if _, oke := AllClients.client[userData.Name]; oke && AllClients.client[userData.Name].status == "alive" { // Check to ensure there aren't 2 account active the same time
		AllClients.mux.Unlock()
		return 1
	}
	client = &Client{
		outgoing: make(chan string),
		conn:     connection,
		reader:   reader,
		writer:   writer,
		data:     LoadUser(userData.Name),
		status:   "alive",
	}

	AllClients.client[userData.Name] = client
	AllClients.mux.Unlock()

	object := Model.Object{
		MessData: Model.Message{
			ExistRoom: LoadRoomOfUser(userData.Name),
		},
		UserData: client.data,
		Commandata: Model.Command{
			Code:   command.Code,
			ID:     command.ID,
			Status: "0",
		},
	}

	send, err := DataSendToStringJson(object)
	if err != nil {
		return 1
	}
	Write(writer, send)
	client.Listen()
	return 0
}

func Read(reader *bufio.Reader) (s string, err error) {
	buf := make([]byte, 2048)
	s = ""
	len := 0

	for {
		len, err = reader.Read(buf)
		if bytes.Contains(buf, []byte("<EOF>")) {
			s = s + string(buf[:len])
			s = strings.ReplaceAll(s, "<EOF>", "")
			return s, nil
		}
		if err != nil {
			s = s + string(buf[:len])
			s = strings.ReplaceAll(s, "<EOF>", "")
			return s, err
		}
		s += string(buf[:len])
	}
	s = s + string(buf[:len])
	s = strings.ReplaceAll(s, "<EOF>", "")
	return s, nil
}

func (client *Client) Read() {
	for {
		if client.status == "dead" {
			return
		}
		line, err := Read(client.reader)
		if err == nil {

			go client.ProcessCommand(line)
		} else {
			log.Log.Info(client.data.Name+" Read Bug: ", err)
			break
		}
	}
	client.Dispose()
}

func Write(writer *bufio.Writer, data string) {
	data = data + "<EOF>"
	writer.WriteString(data)
	writer.Flush()
}
func (client *Client) Dispose() {
	client.conn.Close()
	log.Log.Println(client.data.Name + " Disconnected")
	client.status = "dead"
	delete(AllClients.client, client.data.Name)
	client.outgoing <- "\x00\x00\x00"
}
func (client *Client) Write() {
	for data := range client.outgoing {
		if data == "\x00\x00\x00" {
			break
		}
		Write(client.writer, data)

	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}
