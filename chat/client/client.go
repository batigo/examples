package main

import (
	"bati-chat/proto"
	"context"
	"encoding/json"
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

var service = "chat"

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
	flag.StringVar(&name, "name", "name1222", "user name")
	flag.IntVar(&dt, "dt", 1, "device type ")
	flag.Parse()

	com := baticli.CompressorType_Null
	if compressor == "deflate" {
		com = baticli.CompressorType_Deflate
	}

	conf := baticli.ConnConfig{
		Url:        url,
		Uid:        uid,
		Did:        did,
		Dt:         baticli.DeviceType(dt),
		Timeout:    time.Second * 5,
		HeartBeat:  time.Second * 60,
		Compressor: com,
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

	cli.SetRecvMsgHandler(func(msg *baticli.ClientMsg) {
		switch msg.Type {
		case baticli.ClientMsgType_Ack:
			log.Printf("=== recv ack msg, id: %s\n", msg.Id)
		case baticli.ClientMsgType_Biz:
			chatMsg := &proto.ChatMsgRecv{}
			chatMsg.Decode(msg.BizData)
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
		Data: proto.JoinRoomData{
			Room: room,
			Name: name,
			Uid:  uid,
		},
	}
	bs, _ := json.Marshal(joinMsg)
	sendmsgFunc(&baticli.ClientMsg{
		Id:        baticli.Genmsgid(),
		Type:      baticli.ClientMsgType_Biz,
		Ack:       1,
		ServiceId: &service,
		BizData:   bs,
	})

	for i := 0; i < 30+randd.Intn(20); i++ {
		bs, _ = json.Marshal(proto.ChatMsgSend{
			Type: proto.MsgTypeChat,
			Data: proto.ChatData{
				Uid:  uid,
				Room: room,
				Msg:  fmt.Sprintf("msg-%d", i),
			},
		})
		sendmsgFunc(&baticli.ClientMsg{
			Id:        baticli.Genmsgid(),
			Type:      baticli.ClientMsgType_Biz,
			ServiceId: &service,
			BizData:   bs,
		})
		time.Sleep(time.Second * 5)
	}

	quitmsg := proto.ChatMsgSend{
		Type: proto.MsgTypeQuitRoom,
		Data: proto.QuitRoomData{Room: room, Name: name, Uid: uid},
	}
	bs, _ = json.Marshal(quitmsg)
	sendmsgFunc(&baticli.ClientMsg{
		Id:        baticli.Genmsgid(),
		Type:      baticli.ClientMsgType_Biz,
		Ack:       1,
		ServiceId: &service,
		BizData:   bs,
	})
	cli.Close()
}
