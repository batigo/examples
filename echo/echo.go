package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/batigo/baticli-go"
)

func main() {
	var did string
	var uid string
	var dt int
	var compressor string
	var url string

	flag.StringVar(&url, "url", "", "remote ws url")
	flag.StringVar(&did, "did", "", "device id")
	flag.StringVar(&uid, "uid", "", "user id")
	flag.StringVar(&compressor, "compressor", "null", "compressor type: null | deflate ")
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

	cli.SetConnCloseHanler(func() {
		cli.Close()
	})

	cli.SetRecvMsgHandler(func(msg baticli.ClientMsgRecv) {
		switch msg.Type {
		case baticli.ClientMsgTypeAck:
			log.Printf("=== recv ack msg, id: %s\n", msg.Id)
		case baticli.ClientMsgTypeEcho:
			log.Printf("=== recv echo, id: %s, data: %s", msg.Id, msg.Data)
		default:
			log.Printf("recv unknown msg, type: %d\n", msg.Type)
		}
	})

	err = cli.Init()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		sendmsgFunc(baticli.ClientMsgSend{
			Id:        baticli.Genmsgid(),
			Type:      baticli.ClientMsgTypeEcho,
			Ack:       1,
			ServiceId: "chat",
			Data:      fmt.Sprintf("msg-%d", i),
		})
		time.Sleep(time.Second)
	}

	cli.Close()
}
