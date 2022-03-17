package main

import (
	"bati-chat/proto"
	"github.com/batigo/bati-go"
	"log"
)

func main() {
	server := ChatServer{
		users:  map[string]string{},
		rooms:  map[string]string{},
		stopch: make(chan interface{}),
	}

	conf := bati.PostmanConf{
		Servcie: "chat",
		Kafka: &bati.KafkaSt{
			Hostports: "192.168.2.107:9092",
			GroupId:   "chat",
			Readers:   1,
			Writers:   1,
		},
	}
	postman, err := bati.NewPostman(conf, server.msgHandler)
	if err != nil {
		panic(err.Error())
	}

	server.postman = postman
	err = postman.Run()
	if err != nil {
		panic(err.Error())
	}

	<-server.stopch
}

type ChatServer struct {
	users   map[string]string
	rooms   map[string]string
	stopch  chan interface{}
	postman *bati.Postman
}

func (cs *ChatServer) msgHandler(msg bati.BatiMsg, service string) (err error) {
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	switch msg.Type {
	case bati.BatiMsgTypeBiz:
		log.Printf("recv biz msg: %s- %s", msg.Id, msg.Data)
		chatMsg := &proto.ChatMsgRecv{}
		err = chatMsg.Decode(msg.Data)
		if err != nil {
			return
		}
		switch chatMsg.Type {
		case proto.MsgTypeJoinRoom:
			var data = &proto.JoinRoomData{}
			err = data.Decode(chatMsg.Data)
			if err != nil {
				return
			}
			cs.users[msg.Cid] = data.Name
			cs.rooms[msg.Cid] = data.Room
			err = cs.postman.SendConnJoinMsg(msg.Cid, []string{data.Room}, true)
			if err != nil {
				return
			}
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeJoinRoom,
				Data: proto.JoinRoomData{
					Room: data.Room,
					Name: data.Name,
				},
			}, 100, nil, []string{msg.Uid})

		case proto.MsgTypeQuitRoom:
			var data = &proto.QuitRoomData{}
			err = data.Decode(chatMsg.Data)
			if err != nil {
				return
			}
			defer delete(cs.rooms, msg.Cid)
			defer delete(cs.users, msg.Cid)
			err = cs.postman.SendConnQuitMsg(msg.Cid, []string{data.Room}, true)
			if err != nil {
				return
			}
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeQuitRoom,
				Data: proto.QuitRoomData{
					Room: data.Room,
					Name: data.Name,
				},
			}, 100, nil, []string{msg.Uid})

		case proto.MsgTypeChat:
			var data = &proto.ChatData{}
			err = data.Decode(chatMsg.Data)
			if err != nil {
				return
			}
			if data.Name == "" {
				data.Name = cs.users[msg.Cid]
			}
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeChat,
				// chat-server广播消息
				Data: data,
			}, 100, nil, []string{msg.Uid})
		}

	case bati.BatiMsgTypeConnQuit:
		log.Printf("recv conn-quit msg: %s", msg.Id)
		cid := msg.Cid

		defer delete(cs.rooms, cid)
		defer delete(cs.users, cid)
		room, ok := cs.rooms[cid]
		if !ok {
			return
		}
		name, ok := cs.users[cid]
		if !ok {
			return
		}
		return cs.postman.SendRoomBizMsg(room, proto.ChatMsgSend{
			Type: proto.MsgTypeQuitRoom,
			Data: proto.QuitRoomData{
				Room: room,
				Name: name,
			},
		})
	}

	return
}
