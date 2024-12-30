package internal

import (
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

var boxesLock = sync.RWMutex{}
var clientsLock = sync.RWMutex{}
var Logger echo.Logger

var mvBoxes = make(map[int]*Box)
var clients = make(map[int]*Client)

type Client struct {
	UserId      int
	DisplayName string
	Conn        *websocket.Conn
	Mu          *sync.Mutex
	ElapsedSec  int
}

type Box struct {
	Id        int
	ClientIds []int
	OwnerId   int
	Mu        *sync.RWMutex
}

type BoxSocketData struct {
	Datatype     int      `json:"datatype"`
	SenderID     int      `json:"sender_id"`
	DisplayName  string   `json:"display_name"`
	Content      string   `json:"content"`
	BoxUserNum   int      `json:"box_user_num"`
	MovieUrl     string   `json:"movie_url"`
	MovieTitle   string   `json:"movie_title"`
	Elapsed      float64  `json:"elapsed"`
	IsPaused     bool     `json:"is_paused"`
	IsOwner      bool     `json:"is_owner"`
	BoxId        int      `json:"box_id"`
	MemUsernames []string `json:"mem_usernames"`
	MemIds       []int    `json:"mem_ids"`
}

type ForwardSocketData struct {
	Datatype    int    `json:"datatype"`
	SenderID    int    `json:"sender_id"`
	ReceiverID  int    `json:"receiver_id"`
	DisplayName string `json:"display_name"`
	Content     string `json:"content"`
}

type ClientBoxData struct {
	BoxId       int     `json:"box_id"`
	MovieId     int     `json:"movie_id"`
	Elapsed     float64 `json:"elapsed"`
	UserId      int     `json:"user_id"`
	DisplayName string  `json:"display_name"`
	IsOwner     bool    `json:"is_owner"`
}

func Forward(senderId int, receiverId int, data *ForwardSocketData) error {
	clientsLock.RLock()
	defer clientsLock.RUnlock()

	data.SenderID = senderId
	data.DisplayName = clients[senderId].DisplayName

	if receiver, ok := clients[receiverId]; ok {
		err := websocket.JSON.Send(receiver.Conn, &data)
		return err
	}
	return nil
}

func (s *Box) Broadcast(senderId int, data *BoxSocketData) error {
	if data.Datatype < 0 {
		return nil
	}
	data.SenderID = senderId

	clientsLock.RLock()
	defer clientsLock.RUnlock()

	if data.Datatype == 2 {
		for _, cid := range s.ClientIds {
			data.MemUsernames = append(data.MemUsernames, clients[cid].DisplayName)
			data.MemIds = append(data.MemIds, clients[cid].UserId)
		}
	}

	if senderId >= 0 {
		if c, ok := clients[senderId]; ok {
			data.DisplayName = c.DisplayName
		}
	}
	for _, cid := range s.ClientIds {
		if clients[cid].UserId == senderId && data.Datatype != 2 {
			continue
		}
		data.IsOwner = clients[cid].UserId == s.OwnerId
		data.BoxUserNum = len(s.ClientIds)
		data.BoxId = s.Id

		err := websocket.JSON.Send(clients[cid].Conn, &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetClient(userId int) (*Client, bool) {
	clientsLock.RLock()
	defer clientsLock.RUnlock()
	c, ok := clients[userId]
	return c, ok
}

func GetBox(boxId int) (*Box, bool) {
	boxesLock.RLock()
	defer boxesLock.RUnlock()
	b, ok := mvBoxes[boxId]
	return b, ok
}

func GetBoxOfClient(userId int) *Box {
	boxesLock.RLock()
	var tempBoxes []*Box
	for _, box := range mvBoxes {
		tempBoxes = append(tempBoxes, box)
	}
	boxesLock.RUnlock()

	for _, b := range tempBoxes {
		b.Mu.RLock()
		for _, cid := range b.ClientIds {
			if cid == userId {
				b.Mu.RUnlock()
				return b
			}
		}
		b.Mu.RUnlock()
	}
	return nil
}

func Add(userId int, displayName string, conn *websocket.Conn) (*Client, bool) {
	clientsLock.Lock()
	if c, ok := clients[userId]; ok {
		clientsLock.Unlock()
		c.Mu.Lock()
		defer c.Mu.Unlock()
		if c.Conn != nil && c.Conn != conn {
			c.Conn.Close()
		}
		c.Conn = conn
		return c, false
	}
	defer clientsLock.Unlock()
	client := &Client{
		UserId:      userId,
		DisplayName: displayName,
		Conn:        conn,
		Mu:          &sync.Mutex{},
		ElapsedSec:  0,
	}
	clients[userId] = client
	return client, true
}

func Remove(userId int) {
	clientsLock.Lock()
	delete(clients, userId)
	clientsLock.Unlock()

	boxesLock.RLock()
	var tempBoxes []*Box
	for _, box := range mvBoxes {
		tempBoxes = append(tempBoxes, box)
	}
	boxesLock.RUnlock()

	for _, box := range tempBoxes {
		box.Remove(userId)
	}
}

func AddBox(boxId int, ownerId int) *Box {
	boxesLock.Lock()
	defer boxesLock.Unlock()

	mvBoxes[boxId] = &Box{
		OwnerId: ownerId,
		Id:      boxId,
		Mu:      &sync.RWMutex{},
	}
	mvBoxes[boxId].ClientIds = make([]int, 0)
	return mvBoxes[boxId]
}

func RemoveBox(boxId int) {
	boxesLock.Lock()
	defer boxesLock.Unlock()
	delete(mvBoxes, boxId)
}

func (s *Box) Add(userId int) {
	clientsLock.RLock()
	_, ok := clients[userId]
	clientsLock.RUnlock()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	for _, cid := range s.ClientIds {
		if cid == userId {
			return
		}
	}
	if ok {
		s.ClientIds = append(s.ClientIds, userId)
	}
}

func (s *Box) Remove(userId int) {
	s.Mu.Lock()

	if s.OwnerId == userId {
		s.Mu.Unlock()
		RemoveBox(s.Id)
		return
	}
	defer s.Mu.Unlock()

	for idx, cid := range s.ClientIds {
		if cid == userId {
			s.ClientIds = append(s.ClientIds[:idx], s.ClientIds[idx+1:]...)
			return
		}
	}
}

func SetClientElapsed(userId int, elapsed int) {
	if c, ok := GetClient(userId); ok {
		c.Mu.Lock()
		defer c.Mu.Unlock()
		c.ElapsedSec = elapsed
	}
}

type UserResponse struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsOnline    bool   `json:"is_online"`
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
	Elapsed  float64 `json:"elapsed"`
	MovieID  int     `json:"movie_id"`
	Password string  `json:"password"`
	UserIDs  []int   `json:"user_ids"`
}

type BoxCreateRequest struct {
	OwnerId  int    `json:"owner_id"`
	MsgBoxId int    `json:"msg_box_id"`
	Password string `json:"password"`
}

type BoxMovieUpdateClientRequest struct {
	MovieId  int     `json:"movie_id"`
	Elapsed  float64 `json:"elapsed"`
	IsPaused bool    `json:"is_paused"`
}

type BoxMovieUpdateRequest struct {
	MovieId int     `json:"movie_id"`
	Elapsed float64 `json:"elapsed"`
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

type MessageResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type DisplayNameMessageResponse struct {
	UserId      int    `json:"user_id"`
	DisplayName string `json:"display_name"`
	Content     string `json:"content"`
}

type BoxMessageRequest struct {
	UserID  int    `json:"user_id"`
	BoxID   int    `json:"box_id"`
	Content string `json:"content"`
}

type DirectMessageRequest struct {
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

type BoxPreviewResponse struct {
	BoxID            int     `json:"box_id"`
	OwnerID          int     `json:"owner_id"`
	OwnerDisplayName string  `json:"owner_display_name"`
	Elapsed          float64 `json:"elapsed"`
	MovieID          int     `json:"movie_id"`
	MovieTitle       string  `json:"movie_title"`
	MoviePosterUrl   string  `json:"movie_poster_url"`
	NumberOfMember   int     `json:"member_num"`
}
