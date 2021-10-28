package prisma

import (
	"Tiamat/Model"
	"Tiamat/log"
	"Tiamat/prisma/prisma-client"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

var (
	Prisclient  *prisma.Client
	Priscontext context.Context
)

func InitPrisma() {
	Prisclient = prisma.New(nil)
	Priscontext = context.TODO()
	return
}
func Hash256(msg string) string {
	tmp := sha256.Sum256([]byte(msg))
	msg256 := hex.EncodeToString(tmp[:])
	return string(msg256)
}
func Login(UserName string, Password string) int {

	user, err := Prisclient.User(prisma.UserWhereUniqueInput{
		UserName: &UserName,
	}).Exec(Priscontext)
	if err != nil || user.PassWord != Password {
		log.Log.Warn(err)
		return 1
	}

	return 0
}
func Register(UserName string, Password string, Email string) int {
	_, err := Prisclient.CreateUser(prisma.UserCreateInput{
		Email:    Email,
		UserName: UserName,
		Alias:    UserName,
		PassWord: Password,
		Ava:      "",
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}

func CreateRoom(RoomName string, RoomID string, HostRoom string) int {

	_, err := Prisclient.CreateRoom(prisma.RoomCreateInput{
		Name:      RoomName,
		Roomid:    RoomID,
		Createday: time.Now().Format(time.RFC3339),
		Users: &prisma.UserCreateManyInput{
			Connect: []prisma.UserWhereUniqueInput{
				prisma.UserWhereUniqueInput{
					UserName: &HostRoom,
				},
			},
		},
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}

func JoinRoom(RoomID string, UserID string) int {

	_, err := Prisclient.UpdateRoom(prisma.RoomUpdateParams{
		Where: prisma.RoomWhereUniqueInput{
			Roomid: &RoomID,
		},
		Data: prisma.RoomUpdateInput{
			Users: &prisma.UserUpdateManyInput{
				Connect: []prisma.UserWhereUniqueInput{
					prisma.UserWhereUniqueInput{
						UserName: &UserID,
					},
				},
			},
		},
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}

func ExitRoom(RoomID string, UserID string) int {

	_, err := Prisclient.UpdateRoom(prisma.RoomUpdateParams{
		Where: prisma.RoomWhereUniqueInput{
			Roomid: &RoomID,
		},
		Data: prisma.RoomUpdateInput{
			Users: &prisma.UserUpdateManyInput{
				Disconnect: []prisma.UserWhereUniqueInput{
					prisma.UserWhereUniqueInput{
						UserName: &UserID,
					},
				},
			},
		},
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}

func LoadRoomOfUser(UserName string) []Model.Room {
	rooms, _ := Prisclient.Rooms(&prisma.RoomsParams{
		Where: &prisma.RoomWhereInput{
			UsersSome: &prisma.UserWhereInput{
				UserName: &UserName,
			},
		},
	}).Exec(Priscontext)
	res := make([]Model.Room, 0)
	for _, room := range rooms {
		users := LoadRoomUser(room.Roomid)
		tmp := Model.Room{
			ID:      room.Roomid,
			Name:    room.Name,
			Members: users,
		}
		res = append(res, tmp)
	}
	return res
}

func LoadRoomUser(RoomID string) []Model.User {
	room, _ := Prisclient.Room(prisma.RoomWhereUniqueInput{
		Roomid: &RoomID,
	}).Users(nil).Exec(Priscontext)

	users := make([]Model.User, 0)
	for _, user := range room {
		tmp := Model.User{
			Name:  user.UserName,
			Email: user.Email,
			Alias: user.Alias,
		}
		users = append(users, tmp)
	}
	return users
}
func LoadUser(UserName string) Model.User {
	user, _ := Prisclient.User(prisma.UserWhereUniqueInput{
		UserName: &UserName,
	}).Exec(Priscontext)
	res := Model.User{
		Name:  user.UserName,
		Alias: user.Alias,
		Email: user.Email,
		Ava:   user.Ava,
	}
	return res
}

func LoadRoom(RoomID string) (Model.Room, int) {
	room, err := Prisclient.Room(prisma.RoomWhereUniqueInput{
		Roomid: &RoomID,
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return Model.Room{}, 1
	}
	return Model.Room{
		ID:      room.Roomid,
		Name:    room.Name,
		Members: LoadRoomUser(RoomID),
	}, 0
}
func EditPass(UserName string, Pass string) int {
	_, err := Prisclient.UpdateUser(prisma.UserUpdateParams{
		Where: prisma.UserWhereUniqueInput{
			UserName: &UserName,
		},
		Data: prisma.UserUpdateInput{
			PassWord: &Pass,
		},
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}
func EditProfile(UserName string, Alias string, Email string, Ava string) int {
	_, err := Prisclient.UpdateUser(prisma.UserUpdateParams{
		Where: prisma.UserWhereUniqueInput{
			UserName: &UserName,
		},
		Data: prisma.UserUpdateInput{
			Alias: &Alias,
			Email: &Email,
			Ava:   &Ava,
		},
	}).Exec(Priscontext)
	if err != nil {
		log.Log.Warn(err)
		return 1
	}
	return 0
}
