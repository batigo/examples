package main

import (
	"bati-chat/proto"
	"github.com/batigo/bati-go"
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
	switch msg.Type {
	case bati.BatiMsgTypeBiz:
		chatMsg := &proto.ChatMsgRecv{}
		chatMsg.Decode(msg.Data)
		switch chatMsg.Type {
		case proto.MsgTypeJoinRoom:
			var data = &proto.JoinRoomData{}
			data.Decode(chatMsg.Data)
			cs.users[msg.Cid] = data.Name
			cs.rooms[msg.Cid] = data.Room
			cs.postman.SendConnJoinMsg(msg.Cid, []string{data.Room}, true)
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeJoinRoom,
				Data: proto.JoinRoomData{
					Room: data.Room,
					Name: data.Name,
				},
			}, 100, nil, nil)

		case proto.MsgTypeQuitRoom:
			var data = &proto.QuitRoomData{}
			data.Decode(chatMsg.Data)
			defer delete(cs.rooms, msg.Cid)
			defer delete(cs.users, msg.Cid)
			cs.postman.SendConnQuitMsg(msg.Cid, []string{data.Room}, true)
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeQuitRoom,
				Data: proto.QuitRoomData{
					Room: data.Room,
					Name: data.Name,
				},
			}, 100, nil, nil)

		case proto.MsgTypeChat:
			var data = &proto.ChatData{}
			data.Decode(msg.Data)
			if data.Name == "" {
				data.Name = cs.users[msg.Cid]
			}
			return cs.postman.SendRoomBizMsgCond(data.Room, proto.ChatMsgSend{
				Type: proto.MsgTypeChat,
				// chat-server广播消息
				Data: data,
			}, 100, nil, nil)
		}

	case bati.BatiMsgTypeConnQuit:
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
