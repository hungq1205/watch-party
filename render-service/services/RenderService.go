package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"render-service/internal"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

const (
	userServiceAddr    = "http://localhost:3001"
	messageServiceAddr = "http://localhost:3002"
	movieServiceAddr   = "http://localhost:3003"

	CONN_TIMEOUT = 33
)

var (
	delLock = sync.Mutex{}
)

type RenderService struct {
}

func (s *RenderService) Start(port int) {
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.Static("static"))

	e.GET("/", MainPage)
	e.GET("/box", MainPage)
	e.GET("/login", LogInPage)
	e.GET("/lobby", LobbyPage)

	e.GET("/clientBoxData", ClientBoxData)
	e.GET("/ws", MessageHandler)

	e.POST("/login", LogIn)
	e.POST("/signup", SignUp)
	e.POST("/join", Join)

	e.GET("/api/user/:user_id", GetUser)
	e.GET("/api/user/:user_id/friends", GetFriends)
	e.GET("/api/movie/search", SearchMovies)
	e.GET("/api/movie/:movie_id", GetMovie)
	e.GET("/api/box/:box_id/exists/:user_id", ContainsUser)
	e.GET("/api/box/:box_id/msg", GetBoxMessages)
	e.GET("/api/user/:sender_id/msg/:receiver_id", GetDirectMessages)

	e.POST("/api/user/:sender_id/add/:receiver_id", AddFriend)
	e.POST("/api/box", CreateBox)
	e.POST("/api/box/:box_id/msg/:user_id", CreateBoxMessage)
	e.POST("/api/user/:sender_id/msg/:receiver_id", CreateDirectMessage)
	e.POST("/api/box/:box_id/movie/update", BroadcastMovieOfBox)

	e.PATCH("/api/box/:box_id/movie", UpdateMovieOfBox)

	e.DELETE("/api/user/:user_id/friends/:friend_id", Unfriend)
	e.DELETE("/api/box/:box_id/user/:user_id", RemoveFromBox)
	e.DELETE("/api/box/:box_id", DeleteBox)

	log.Println("Starting Render Service on port ", port, "...")
	internal.Logger = e.Logger
	if err := e.Start(fmt.Sprint(":", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func LogInPage(c echo.Context) error {
	_, err := Authenticate(c)
	if err == nil {
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().Header().Set("Expires", "0")

		return c.Redirect(http.StatusTemporaryRedirect, "/lobby")
	}

	return c.File("static/views/login.html")
}

func LobbyPage(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return c.Redirect(http.StatusTemporaryRedirect, "/box")
	} else if resp.StatusCode != http.StatusNotFound {
		rbytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(resp.StatusCode, string(rbytes))
	}

	return c.File("static/views/lobby.html")
}

func MainPage(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				c.Logger().Debug("back 1")
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		c.Logger().Debug("back 2")
		return c.Redirect(http.StatusMovedPermanently, "/login")
	}

	return c.File("static/views/index.html")
}

func MessageHandler(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var box *internal.Box
	box = nil
	if resp.StatusCode == http.StatusOK {
		var boxOfUserResponse internal.IdentifierResponse
		if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
		}
		resp.Body.Close()

		getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
		resp, err = http.Get(getBoxURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}

		var boxRes internal.BoxResponse
		if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
		}
		resp.Body.Close()

		var ok bool
		box, ok = internal.GetBox(boxRes.ID)
		if !ok {
			internal.AddBox(boxRes.ID, auth.UserID)
			box, _ = internal.GetBox(boxRes.ID)
		}
	} else if resp.StatusCode != http.StatusNotFound {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}

	websocket.Handler(func(ws *websocket.Conn) {
		if _, isCreated := internal.Add(auth.UserID, auth.DisplayName, ws); isCreated {
			go TrackSocket(auth.UserID)
		}
		if err := RawUpdateUserState(auth.UserID, true); err != nil {
			c.Logger().Printf("failed to update user state for user %d: \n%v", auth.UserID, err)
		}
		c.Logger().Printf("User %s(#%d) went online", auth.Username, auth.UserID)

		if box != nil {
			box.Add(auth.UserID)
			for {
				var data internal.BoxSocketData
				err = websocket.JSON.Receive(ws, &data)
				internal.SetClientElapsed(auth.UserID, 0)
				if err != nil {
					break
				}

				if box, ok := internal.GetBox(box.Id); ok {
					if err = box.Broadcast(auth.UserID, &data); err != nil {
						internal.Logger.Printf("Broadcasting err in box %d: %v", box.Id, err.Error())
						break
					}
				}
			}
		} else {
			for {
				var data internal.ForwardSocketData
				err = websocket.JSON.Receive(ws, &data)
				internal.SetClientElapsed(auth.UserID, 0)
				if err != nil {
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return err
}

func TrackSocket(userId int) {
	for {
		time.Sleep(time.Duration(11) * time.Second)
		client, ok := internal.GetClient(userId)
		if !ok {
			continue
		}

		client.Mu.Lock()
		client.ElapsedSec += 11
		elapsed := client.ElapsedSec
		displayName := client.DisplayName
		client.Mu.Unlock()

		if elapsed >= CONN_TIMEOUT {
			if err := CloseSocket(userId); err != nil {
				internal.Logger.Printf(err.Error())
				return
			}
			internal.Logger.Printf("User %s(#%d) went offline", displayName, userId)
			return
		}
	}
}

func CloseSocket(userId int) error {
	if client, ok := internal.GetClient(userId); ok {
		if box := internal.GetBoxOfClient(userId); box != nil {
			RawRemoveFromBox(userId, box.Id)
			box.Broadcast(-1, &internal.BoxSocketData{Datatype: 2})
		}
		client.Mu.Lock()
		client.Conn.Close()
		client.Mu.Unlock()
		if err := RawUpdateUserState(userId, false); err != nil {
			return fmt.Errorf("failed to update user state for user %d: \n%v", userId, err)
		}
	}
	return nil
}

func UpdateMovieOfBox(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var payload internal.BoxMovieUpdateClientRequest
	if err := c.Bind(&payload); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusUnauthorized, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get the box id of owner
	boxOfUserURL := fmt.Sprintf("%s/api/box/owner/%d", movieServiceAddr, auth.UserID)

	resp, err := http.Get(boxOfUserURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return c.String(http.StatusNotFound, "Box not found for the user")
		}
		return c.String(http.StatusBadRequest, "Failed to retrieve users box")
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse users box ID response")
	}

	if boxId != boxOfUserResponse.ID {
		return c.NoContent(http.StatusUnauthorized)
	}

	payloadJson, err := json.Marshal(&internal.BoxMovieUpdateRequest{
		MovieId: payload.MovieId,
		Elapsed: payload.Elapsed,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse update box movie request to json")
	}

	url := fmt.Sprintf("%s/api/box/%d/movie", movieServiceAddr, boxId)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(payloadJson))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to construct patch request")
	}
	client := http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return c.String(resp.StatusCode, "Failed to update box movie")
	}

	var socketData *internal.BoxSocketData
	if payload.MovieId >= 0 {
		movieRes, err := RawGetMovie(payload.MovieId)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		socketData = &internal.BoxSocketData{
			Datatype:   1,
			IsPaused:   payload.IsPaused,
			Elapsed:    payload.Elapsed,
			MovieUrl:   movieRes.URL,
			MovieTitle: movieRes.Title,
		}
	} else {
		socketData = &internal.BoxSocketData{
			Datatype:   1,
			IsPaused:   true,
			Elapsed:    0,
			MovieUrl:   "",
			MovieTitle: "",
		}
	}

	if box, ok := internal.GetBox(boxId); ok {
		box.Broadcast(auth.UserID, socketData)
	}

	return c.NoContent(http.StatusOK)
}

func BroadcastMovieOfBox(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	url := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to retrieve box")
	}

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse users box response")
	}

	var payload *internal.BoxSocketData
	if boxRes.MovieID >= 0 {
		movieRes, err := RawGetMovie(boxRes.MovieID)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		payload = &internal.BoxSocketData{
			Datatype:   1,
			IsPaused:   false,
			Elapsed:    boxRes.Elapsed,
			MovieUrl:   movieRes.URL,
			MovieTitle: movieRes.Title,
		}
	} else {
		payload = &internal.BoxSocketData{
			Datatype:   1,
			IsPaused:   true,
			Elapsed:    0,
			MovieUrl:   "",
			MovieTitle: "",
		}
	}

	if box, ok := internal.GetBox(boxId); ok {
		box.Broadcast(box.OwnerId, payload)
	}

	return c.NoContent(http.StatusOK)
}

func CreateBoxMessage(c echo.Context) error {
	var payload internal.BoxMessageRequest
	if err := c.Bind(&payload); err != nil {
		return c.String(http.StatusBadRequest, "Invalid message creating request body")
	}

	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user_id value")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if auth.UserID != userId {
		return c.NoContent(http.StatusUnauthorized)
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
	}
	if boxId != boxOfUserResponse.ID {
		return c.NoContent(http.StatusUnauthorized)
	}

	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve movie box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	payload.BoxID = boxRes.MsgBoxID
	payload.UserID = userId
	payloadStr, err := json.Marshal(payload)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create message payload")
	}

	url := fmt.Sprintf("%s/api/msg", messageServiceAddr)
	resp, err = http.Post(url, "application/json", bytes.NewReader(payloadStr))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return c.String(resp.StatusCode, "Failed to create message")
	}

	data := internal.BoxSocketData{
		Datatype: 0,
		Content:  payload.Content,
	}
	internal.SetClientElapsed(auth.UserID, 0)
	if box, ok := internal.GetBox(boxRes.ID); ok {
		if err = box.Broadcast(auth.UserID, &data); err != nil {
			internal.Logger.Printf("Broadcasting err in box %d: %v", box.Id, err.Error())
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return c.NoContent(http.StatusCreated)
}

func CreateDirectMessage(c echo.Context) error {
	senderId, err := strconv.Atoi(c.Param("sender_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid sender_id value")
	}

	receiverId, err := strconv.Atoi(c.Param("receiver_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid receiver_id value")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if senderId != auth.UserID {
		return c.NoContent(http.StatusUnauthorized)
	}

	var payload internal.DirectMessageRequest
	if err = c.Bind(&payload); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	payload.SenderID = senderId
	payload.ReceiverID = receiverId

	payloadStr, err := json.Marshal(payload)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create message payload")
	}

	url := fmt.Sprintf("%s/api/msg", messageServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payloadStr))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return c.String(resp.StatusCode, "Failed to create message")
	}

	internal.SetClientElapsed(auth.UserID, 0)
	if _, ok := internal.GetClient(receiverId); ok {
		data := internal.ForwardSocketData{
			Datatype: 3,
			Content:  payload.Content,
		}
		internal.Forward(auth.UserID, receiverId, &data)
	}

	return c.NoContent(http.StatusCreated)
}

func GetBoxMessages(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid box_id value")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
	}

	if boxOfUserResponse.ID != boxId {
		return c.NoContent(http.StatusUnauthorized)
	}

	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve movie box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	url := fmt.Sprintf("%s/api/msg?box_id=%d", messageServiceAddr, boxRes.MsgBoxID)
	resp, err = http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to get messages by box id")
	}

	var messages []internal.DisplayNameMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse messages")
	}

	for i := range messages {
		user, err := RawGetUser(messages[i].UserId)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch user for user_id %d: %v", messages[i].UserId, err))
		}
		messages[i].DisplayName = user.DisplayName
	}

	return c.JSON(http.StatusOK, messages)
}

func GetDirectMessages(c echo.Context) error {
	senderId, err := strconv.Atoi(c.Param("sender_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid sender_id value")
	}

	receiverId, err := strconv.Atoi(c.Param("receiver_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid receiver_id value")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if auth.UserID != senderId {
		return c.NoContent(http.StatusUnauthorized)
	}

	url := fmt.Sprintf("%s/api/msg/dm?user_id1=%d&user_id2=%d", messageServiceAddr, senderId, receiverId)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to get direct messages with user_id")
	}

	var messages []internal.MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse messages")
	}

	return c.JSON(http.StatusOK, messages)
}

func GetMovie(c echo.Context) error {
	movieId, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	mRes, err := RawGetMovie(movieId)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mRes)
}

func SearchMovies(c echo.Context) error {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	url := fmt.Sprintf("%s/api/movie/search?query=%s&page=%d&size=15", movieServiceAddr, c.QueryParam("query"), page)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to get movies of query "+c.QueryParam("page"))
	}

	var mRes []internal.MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&mRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movies")
	}

	return c.JSON(http.StatusOK, mRes)
}

func ContainsUser(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	res, err := RawExistsUserInBox(boxId, userId)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func Join(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	password := c.FormValue("password")
	rawBoxId, err := strconv.Atoi(c.FormValue("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	boxId := int(rawBoxId)

	userBoxURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userBoxURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return MainPage(c)
	}

	boxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err = http.Get(boxURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to connect: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return c.String(http.StatusNotFound, "Movie box doesn't exist")
	} else if resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve movie box information")
	}

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	if boxRes.Password != password {
		return c.String(http.StatusBadRequest, "Incorrect password or box ID for movie box")
	}

	addToBoxURL := fmt.Sprintf("%s/api/box/%d/add/%d", movieServiceAddr, boxId, auth.UserID)
	resp, err = http.Post(addToBoxURL, "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to add user to movie box")
	}
	defer resp.Body.Close()

	addToMsgBoxURL := fmt.Sprintf("%s/api/msg/box/%d/add/%d", messageServiceAddr, boxRes.MsgBoxID, auth.UserID)
	resp, err = http.Post(addToMsgBoxURL, "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to add user to message box")
	}
	defer resp.Body.Close()

	return c.NoContent(http.StatusOK)
}

func DeleteBox(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse box id")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get the box id of user
	boxOfUserURL := fmt.Sprintf("%s/api/box/owner/%d", movieServiceAddr, auth.UserID)

	resp, err := http.Get(boxOfUserURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return c.String(http.StatusNotFound, "Box not found for the user")
		}
		return c.String(http.StatusBadRequest, "Failed to retrieve users box")
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse users box response")
	}

	if boxId != boxOfUserResponse.ID {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err := RawDeleteBox(boxId); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func RemoveFromBox(c echo.Context) error {
	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse box id")
	}

	kickId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse user id to be removed")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
	}

	if boxOfUserResponse.ID != boxId {
		return c.NoContent(http.StatusUnauthorized)
	}

	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve movie box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	if kickId == boxRes.OwnerID {
		return DeleteBox(c)
	}

	if kickId != auth.UserID && boxRes.OwnerID != auth.UserID {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err := RawRemoveFromBox(kickId, boxId); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func CreateBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	userBoxURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userBoxURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return c.String(http.StatusConflict, "User already in a box")
	}

	password := c.FormValue("password")

	// Create Message Box
	createMsgBoxURL := fmt.Sprintf("%s/api/msg/box", messageServiceAddr)
	resp, err = http.Post(createMsgBoxURL, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf("{\"user_id\":%d}", auth.UserID))))
	if err != nil {
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Creating message box: %v", err))
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.String(resp.StatusCode, "Failed to create message box")
	}

	var msgBoxResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgBoxResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse message box response")
	}

	// Create Movie Box
	createBoxURL := fmt.Sprintf("%s/api/box", movieServiceAddr)
	boxRequest := internal.BoxCreateRequest{
		OwnerId:  auth.UserID,
		MsgBoxId: msgBoxResponse.ID,
		Password: password,
	}
	boxJSON, _ := json.Marshal(boxRequest)
	resp, err = http.Post(createBoxURL, "application/json", bytes.NewBuffer(boxJSON))
	if err != nil {
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Creating movie box: %v", err))
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.String(resp.StatusCode, "Failed to create movie box")
	}

	var boxResponse internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	internal.AddBox(boxResponse.ID, auth.UserID)

	return c.JSON(http.StatusOK, boxResponse)
}

func GetUser(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	res, err := RawGetUser(userId)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}

func GetFriends(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	url := fmt.Sprintf("%s/api/user/%d/friends", userServiceAddr, userId)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("Failed to get friends of user %d", userId))
	}

	var fRes []internal.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&fRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse friends")
	}

	if len(fRes) == 0 {
		fRes = []internal.UserResponse{}
	}

	return c.JSON(http.StatusOK, fRes)
}

func AddFriend(c echo.Context) error {
	senderId, err := strconv.Atoi(c.Param("sender_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse sender id")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if auth.UserID != senderId {
		return c.NoContent(http.StatusUnauthorized)
	}

	url := fmt.Sprintf("%s/api/user/%d/add/%s", userServiceAddr, senderId, c.Param("receiver_id"))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return c.String(resp.StatusCode, err.Error())
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return c.String(resp.StatusCode, fmt.Sprintf("Failed to send friend from %d to %s", senderId, c.Param("receiver_id")))
	}

	return c.NoContent(http.StatusCreated)
}

func Unfriend(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse user id")
	}

	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if auth.UserID != userId {
		return c.NoContent(http.StatusUnauthorized)
	}

	url := fmt.Sprintf("%s/api/user/%d/friends/%s", userServiceAddr, userId, c.Param("friend_id"))
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to unfriend")
	}

	return c.NoContent(http.StatusOK)
}

func LogIn(c echo.Context) error {
	req := internal.LogInRequest{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	loginJSON, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/user/login", userServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to login")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return c.String(http.StatusUnauthorized, "Incorrect username or password")
	}

	var res internal.LogInResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse login response")
	}

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Internal server error")
	}

	token := &http.Cookie{
		Name:    "jwtcookie",
		Value:   res.JwtToken,
		Expires: time.Now().Add(time.Minute * 30),
	}
	c.SetCookie(token)

	return c.NoContent(http.StatusOK)
}

func SignUp(c echo.Context) error {
	req := internal.SignUpRequest{
		Username:    c.FormValue("username"),
		Password:    c.FormValue("password"),
		DisplayName: c.FormValue("display_name"),
	}

	signupJSON, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/user/signup", userServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(signupJSON))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to signup")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return c.String(http.StatusConflict, "Username has already exist")
	}

	if resp.StatusCode != http.StatusCreated {
		return c.String(resp.StatusCode, "Internal server error")
	}

	return c.NoContent(http.StatusCreated)
}

func ClientBoxData(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusUnauthorized {
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			return c.String(he.Code, fmt.Sprintf("%v", he.Message))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var data *internal.ClientBoxData

	userURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(userURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var boxOfUserResponse internal.IdentifierResponse
		if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to parse users box response")
		}

		getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
		resp, err = http.Get(getBoxURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
		defer resp.Body.Close()

		var boxRes internal.BoxResponse
		if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
		}

		data = &internal.ClientBoxData{
			BoxId:       boxOfUserResponse.ID,
			MovieId:     boxRes.MovieID,
			Elapsed:     boxRes.Elapsed,
			UserId:      auth.UserID,
			DisplayName: auth.DisplayName,
			IsOwner:     auth.UserID == boxRes.OwnerID,
		}
	} else if resp.StatusCode == http.StatusNotFound {
		data = &internal.ClientBoxData{
			UserId:      auth.UserID,
			DisplayName: auth.DisplayName,
		}
	} else {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}

	return c.JSON(http.StatusOK, data)
}

func Authenticate(c echo.Context) (*internal.UserResponse, error) {
	jwtcookie, err := c.Cookie("jwtcookie")
	if err != nil || jwtcookie == nil {
		if err == http.ErrNoCookie {
			return nil, echo.NewHTTPError(http.StatusUnauthorized)
		}
		return nil, err
	}

	value := jwtcookie.Value
	payload := map[string]string{
		"jwt_token": value,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := userServiceAddr + "/api/user/authenticate"
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	var user internal.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func RawRemoveFromBox(userId int, boxId int) error {
	if box, ok := internal.GetBox(boxId); ok {
		box.Mu.RLock()
		if userId == box.OwnerId {
			box.Mu.RUnlock()
			return RawDeleteBox(boxId)
		}
		box.Mu.RUnlock()
	} else {
		return fmt.Errorf("failed to retrieve box + %d", boxId)
	}

	// Get Box Details
	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err := http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retrieve box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return fmt.Errorf("failed to parse box details")
	}

	// Remove User from Box
	removeUserURL := fmt.Sprintf("%s/api/box/%d/remove/%d", movieServiceAddr, boxId, userId)
	req, _ := http.NewRequest(http.MethodDelete, removeUserURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove user from box")
	}
	resp.Body.Close()

	// Remove User from Message Box
	removeMsgUserURL := fmt.Sprintf("%s/api/msg/box/%d/remove/%d", messageServiceAddr, boxRes.MsgBoxID, userId)
	req, _ = http.NewRequest(http.MethodDelete, removeMsgUserURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove user from message box")
	}
	resp.Body.Close()

	if box, ok := internal.GetBox(boxRes.ID); ok {
		box.Broadcast(-1, &internal.BoxSocketData{Datatype: 2})
		box.Remove(userId)
		box.Broadcast(-1, &internal.BoxSocketData{Datatype: 2})
	} else {
		return fmt.Errorf("failed to retrieve box + %d", boxRes.ID)
	}
	return nil
}

func RawDeleteBox(boxId int) error {
	// Get Box
	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	resp, err := http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(http.StatusNotFound, "Box not found")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return err
	}

	delLock.Lock()

	// Delete Movie Box
	deleteBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	req, _ := http.NewRequest(http.MethodDelete, deleteBoxURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		delLock.Unlock()
		return err
	}

	// Delete Message Box
	deleteMsgBoxURL := fmt.Sprintf("%s/api/msg/box/%d", messageServiceAddr, boxRes.MsgBoxID)
	req, _ = http.NewRequest(http.MethodDelete, deleteMsgBoxURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		delLock.Unlock()
		return err
	}

	delLock.Unlock()

	if box, ok := internal.GetBox(boxId); ok {
		box.Broadcast(-1, &internal.BoxSocketData{Datatype: 2})
		internal.RemoveBox(boxId)
		box.Broadcast(-1, &internal.BoxSocketData{Datatype: 2})
	} else {
		internal.RemoveBox(boxId)
	}

	return nil
}

func RawUpdateUserState(userId int, isOnline bool) error {
	url := fmt.Sprintf("%s/api/user/%d/state", userServiceAddr, userId)
	requestBody := map[string]bool{
		"is_online": isOnline,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal("Error marshalling JSON:", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending PUT request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update user state. Status code: %d", resp.StatusCode)
	}

	return nil
}

func RawGetMovie(movieId int) (*internal.MovieResponse, error) {
	url := fmt.Sprintf("%s/api/movie/%d", movieServiceAddr, movieId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for movie_id %d", resp.StatusCode, movieId)
	}

	var mRes internal.MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&mRes); err != nil {
		return nil, err
	}

	return &mRes, nil
}

func RawGetUser(userId int) (*internal.UserResponse, error) {
	url := fmt.Sprintf("%s/api/user/%d", userServiceAddr, userId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for user_id %d: %w", userId, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for user_id %d", resp.StatusCode, userId)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for user_id %d: %w", userId, err)
	}

	var userAPIResponse internal.UserResponse
	if err := json.Unmarshal(body, &userAPIResponse); err != nil {
		return nil, fmt.Errorf("failed to parse user response for user_id %d: %w", userId, err)
	}

	return &userAPIResponse, nil
}

func RawExistsUserInBox(boxId int, userId int) (*internal.BooleanResponse, error) {
	containsURL := fmt.Sprintf("%s/api/box/%d/exists/%d", movieServiceAddr, boxId, userId)
	resp, err := http.Get(containsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed checking user existence in box")
	}

	var bRes internal.BooleanResponse
	if err := json.NewDecoder(resp.Body).Decode(&bRes); err != nil {
		return nil, fmt.Errorf("failed to parse user existence in box response")
	}

	return &bRes, nil
}
