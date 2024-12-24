package internal

import (
	"golang.org/x/net/websocket"
)

var MsgBoxes = make(map[int]*MsgBox)

type Client struct {
	UserId   int
	Username string
	Conn     *websocket.Conn
	Box      *MsgBox
}

type MsgBox struct {
	MvBoxId int
	Clients []*Client
	OwnerId int
}

type ClientData struct {
	Datatype     int      `json:"datatype"`
	Username     string   `json:"username"`
	Content      string   `json:"content"`
	BoxUserNum   int      `json:"box_user_num"`
	MovieUrl     string   `json:"movie_url"`
	Elapsed      float64  `json:"elapsed"`
	IsPause      bool     `json:"is_pause"`
	IsOwner      bool     `json:"is_owner"`
	BoxId        int      `json:"box_id"`
	MemUsernames []string `json:"mem_usernames"`
	MemIds       []int    `json:"mem_ids"`
}

type ClientBoxData struct {
	BoxId   int  `json:"box_id"`
	IsOwner bool `json:"is_owner"`
}

func (s *MsgBox) Broadcast(fromUserId int, data *ClientData) error {
	if data.Datatype == 2 {
		for _, client := range s.Clients {
			data.MemUsernames = append(data.MemUsernames, client.Username)
			data.MemIds = append(data.MemIds, client.UserId)
		}
	}

	for _, client := range s.Clients {
		if client.UserId == fromUserId && data.Datatype != 2 {
			continue
		}
		data.IsOwner = client.UserId == s.OwnerId
		data.Username = client.Username
		data.BoxUserNum = len(s.Clients)
		data.BoxId = s.MvBoxId

		err := websocket.JSON.Send(client.Conn, &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MsgBox) AppendNew(userId int, username string, conn *websocket.Conn) {
	client := &Client{
		UserId:   userId,
		Username: username,
		Conn:     conn,
		Box:      s,
	}

	for idx, cli := range s.Clients {
		if cli.UserId == userId {
			s.Clients[idx].Conn = conn
			return
		}
	}

	s.Clients = append(s.Clients, client)
}

func AppendNewMsgBox(boxId int, ownerId int, mvBoxId int) {
	MsgBoxes[boxId] = &MsgBox{OwnerId: ownerId, MvBoxId: mvBoxId}
	MsgBoxes[boxId].Clients = make([]*Client, 0)
}

func (s *MsgBox) Close() {
	for _, client := range s.Clients {
		client.Conn.Close()
	}
}

func (s *MsgBox) Remove(userId int) {
	for idx, client := range s.Clients {
		if client.UserId == userId {
			client.Conn.Close()
			s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
			return
		}
	}
}

type AuthReponse struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

type MovieResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	PosterURL string `json:"poster_url"`
}

type BoxResponse struct {
	ID       int     `json:"id"`
	OwnerID  int     `json:"owner_id"`
	MsgBoxID int     `json:"msg_box_id"`
	Elapsed  float32 `json:"elapsed"`
	MovieURL string  `json:"movie_url"`
	Password string  `json:"password"`
	UserIDs  []int   `json:"user_ids"`
}

type BoxCreateRequest struct {
	OwnerId  int    `json:"owner_id"`
	MsgBoxId int    `json:"msg_box_id"`
	Password string `json:"password"`
}

type BoxAddUserRequest struct {
	UserId int `json:"user_id"`
	BoxId  int `json:"box_id"`
}

type IdentifierResponse struct {
	ID int `json:"id"`
}

type BooleanResponse struct {
	Value bool `json:"value"`
}

type LogInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogInResponse struct {
	JwtToken string `json:"jwt_token"`
}

type SignUpRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}
