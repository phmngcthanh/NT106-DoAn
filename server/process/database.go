package process

import (
	"Tiamat/Model"
	"Tiamat/prisma"
)

func LoadRoomOfUser(username string) []Model.Room {
	rooms := prisma.LoadRoomOfUser(username)
	AllRooms.mux.Lock()
	for _, tmp := range rooms {
		if _, ok := AllRooms.room[tmp.ID]; !ok {
			AllRooms.room[tmp.ID] = &Rooms{
				RoomID:   tmp.ID,
				RoomName: tmp.Name,
				Members:  make(map[string]*Model.User),
			}
		}
		for _, user := range tmp.Members {
			AllRooms.room[tmp.ID].Members[user.Name] = &user
		}
	}
	AllRooms.mux.Unlock()
	return rooms
}
func LoadRoom(roomid string) (Model.Room, int) {
	AllRooms.mux.Lock()
	defer AllRooms.mux.Unlock()
	room, oke := prisma.LoadRoom(roomid)
	if oke == 0 {
		AllRooms.room[roomid] = &Rooms{
			RoomID:   room.ID,
			RoomName: room.Name,
			Members:  make(map[string]*Model.User),
		}
		for _, user := range room.Members {
			AllRooms.room[roomid].Members[user.Name] = &user
		}
	}
	return room, oke
}
func LoadUser(username string) Model.User {
	return prisma.LoadUser(username)
}
func EditPass(username string, pass string) int {
	return prisma.EditPass(username, pass)
}
func EditProfile(username string, alias string, email string, ava string) int {
	return prisma.EditProfile(username, alias, email, ava)
}
