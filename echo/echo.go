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
	cli, sendMsgFunc, err := baticli.NewConn(context.Background(), conf)
	if err != nil {
		panic(err)
	}

	cli.SetConnCloseHanler(func() {
		cli.Close()
	})

	cli.SetRecvMsgHandler(func(msg *baticli.ClientMsg) {
		switch msg.Type {
		case baticli.ClientMsgType_Ack:
			log.Printf("=== recv ack msg, id: %s\n", msg.Id)
		case baticli.ClientMsgType_Echo:
			log.Printf("=== recv echo, id: %s, data: %s", msg.Id, msg.BizData)
		default:
			log.Printf("=== recv unknown msg, type: %d\n", msg.Type)
		}
	})

	err = cli.Init()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		sendMsgFunc(&baticli.ClientMsg{
			Id:      baticli.Genmsgid(),
			Type:    baticli.ClientMsgType_Echo,
			Ack:     1,
			BizData: []byte(fmt.Sprintf("msg-%d", i)),
		})
		time.Sleep(time.Second)
	}

	cli.Close()
}
