package proto

import (
	"encoding/json"
	"fmt"
)

const (
	MsgTypeJoinRoom = 1
	MsgTypeQuitRoom = 2
	MsgTypeChat     = 3
)

type JoinRoomData struct {
	Uid  string `json:"uid"`
	Room string `json:"room"`
	Name string `json:"name"`
}

func (d *JoinRoomData) String() string {
	return fmt.Sprintf("[join] from: %s", d.Name)
}

func (d *JoinRoomData) Decode(bs json.RawMessage) error {
	return json.Unmarshal(bs, d)
}

type QuitRoomData struct {
	Uid  string `json:"uid"`
	Room string `json:"room"`
	Name string `json:"name"`
}

func (d *QuitRoomData) String() string {
	return fmt.Sprintf("[quit] from: %s", d.Name)
}

func (d *QuitRoomData) Decode(bs json.RawMessage) error {
	return json.Unmarshal(bs, d)
}

type ChatData struct {
	Uid  string `json:"uid"`
	Room string `json:"room"`
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

func (d *ChatData) String() string {
	return fmt.Sprintf("[chat] room:%s, from: %s, msg: %s", d.Room, d.Name, d.Msg)
}

func (d *ChatData) Decode(bs json.RawMessage) error {
	return json.Unmarshal(bs, d)
}

type ChatMsgRecv struct {
	Type int             `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (d *ChatMsgRecv) Decode(bs json.RawMessage) {
	json.Unmarshal(bs, d)
}

type ChatMsgSend struct {
	Type int         `json:"type"`
	Data interface{} `json:"data"`
}
