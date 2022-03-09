package main

import (
	"context"
	"flag"
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
	_, err := baticli.NewConn(context.Background(), conf)
	if err != nil {
		panic(err)
	}

}
