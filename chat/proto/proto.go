package proto

const (
	MsgTypeJoinRoom = 1
	MsgTypeQuitRoom = 2
	MsgTypeChat     = 3
)

type JoinRoomData struct {
	Id string `json:"id"`
}

type QuitRoomData struct {
	Id string `json:"id"`
}

type ChatData struct {
	Msg string `json:"msg"`
}
