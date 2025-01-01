package messages

type InitMessage struct {
	Username string `json:"username"`
}

type Message_ struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

type InitRoomMessage struct {
	RoomName string `json:"room_name"`
	Username string `json:"username"`
}

type RoomMessage struct {
	RoomName string `json:"room_name"`
	Username string `json:"username"`
	Message  string `json:"message"`
}
