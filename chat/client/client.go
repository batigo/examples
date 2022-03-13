package main

import (
	"bati-chat/proto"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/batigo/baticli-go"
)

var randd *rand.Rand

func init() {
	randd = rand.New(rand.NewSource(int64(os.Getpid())))
}

func main() {
	var did string
	var uid string
	var dt int
	var compressor string
	var url string
	var room string
	var name string

	flag.StringVar(&url, "url", "", "remote ws url")
	flag.StringVar(&did, "did", "", "device id")
	flag.StringVar(&uid, "uid", "", "user id")
	flag.StringVar(&compressor, "compressor", "null", "compressor type: null | deflate ")
	flag.StringVar(&room, "room", "room1", "join room id")
	flag.StringVar(&name, "name", "name1", "user name")
	flag.IntVar(&dt, "dt", 1, "device type ")
	flag.Parse()

	conf := baticli.ConnConfig{
		Url:        url,
		Uid:        uid,
		Did:        did,
		Dt:         baticli.DeviceType(dt),
		Timeout:    time.Second * 5,
		HeartBeat:  time.Second * 60,
		Compressor: baticli.CompressorType(compressor),
		BinaryMsg:  false,
	}
	cli, sendmsgFunc, err := baticli.NewConn(context.Background(), conf)
	if err != nil {
		panic(err)
	}

	stopch := make(chan interface{})

	cli.SetConnCloseHanler(func() {
		close(stopch)
		cli.Close()
	})

	cli.SetRecvMsgHandler(func(msg baticli.ClientMsgRecv) {
		switch msg.Type {
		case baticli.ClientMsgTypeAck:
			log.Printf("=== recv ack msg, id: %s\n", msg.Id)
		case baticli.ClientMsgTypeBiz:
			chatMsg := &proto.ChatMsgRecv{}
			chatMsg.Decode(msg.Data)
			switch chatMsg.Type {
			case proto.MsgTypeJoinRoom:
				var join proto.JoinRoomData
				join.Decode(chatMsg.Data)
				log.Printf("=== %s\n", join.String())
			case proto.MsgTypeQuitRoom:
				var quit proto.QuitRoomData
				quit.Decode(chatMsg.Data)
				log.Printf("=== %s\n", quit.String())
			case proto.MsgTypeChat:
				var chat proto.ChatData
				chat.Decode(chatMsg.Data)
				log.Printf("=== %s\n", chat.String())
			}
		default:
			log.Printf("recv unknown msg, type: %d\n", msg.Type)
		}
	})

	err = cli.Init()
	if err != nil {
		panic(err)
	}

	joinMsg := proto.ChatMsgSend{
		Type: proto.MsgTypeJoinRoom,
		Data: proto.JoinRoomData{Room: room, Name: name, Uid: uid},
	}
	sendmsgFunc(baticli.ClientMsgSend{
		Id:        baticli.Genmsgid(),
		Type:      baticli.ClientMsgTypeBiz,
		Ack:       1,
		ServiceId: "chat",
		Data:      joinMsg,
	})

	for i := 0; i < 30+randd.Intn(20); i++ {
		sendmsgFunc(baticli.ClientMsgSend{
			Id:        baticli.Genmsgid(),
			Type:      baticli.ClientMsgTypeBiz,
			Ack:       1,
			ServiceId: "chat",
			Data: proto.ChatMsgSend{
				Type: proto.MsgTypeChat,
				Data: proto.ChatData{
					Uid:  uid,
					Room: room,
					Msg:  fmt.Sprintf("msg-%d", i),
				},
			},
		})
	}

	quitmsg := proto.ChatMsgSend{
		Type: proto.MsgTypeQuitRoom,
		Data: proto.QuitRoomData{Room: room, Name: name, Uid: uid},
	}
	sendmsgFunc(baticli.ClientMsgSend{
		Id:        baticli.Genmsgid(),
		Type:      baticli.ClientMsgTypeBiz,
		Ack:       1,
		ServiceId: "chat",
		Data:      quitmsg,
	})
	cli.Close()
}
