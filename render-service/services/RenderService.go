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
)

var (
	lock = sync.Mutex{}
)

type RenderService struct {
}

func (s *RenderService) Start(port int) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.Static("static"))

	e.GET("/", MainPage)
	e.GET("/box", MainPage)
	e.GET("/login", LogInPage)
	e.GET("/lobby", LobbyPage)
	e.GET("/movie", GetMovies)
	e.GET("/movie/:movie_id", GetMovie)
	e.GET("/clientBoxData", ClientBoxData)
	e.GET("/ws", MessageHandler)
	e.GET("/box/:box_id/exists", ContainsUser)

	e.POST("/join", JoinBox)
	e.POST("/login", LogIn)
	e.POST("/signup", SignUp)
	e.POST("/create", CreateBox)
	e.POST("/delete", DeleteBox)
	e.POST("/leave", LeaveBox)
	e.POST("/kick/:user_id", KickFromBox)

	log.Println("Starting Render Service on port ", port, "...")
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
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
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

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		if internal.MsgBoxes[boxRes.MsgBoxID].Clients == nil {
			internal.AppendNewMsgBox(boxRes.MsgBoxID, auth.UserID, boxRes.ID)
		}
		internal.MsgBoxes[boxRes.MsgBoxID].AppendNew(auth.UserID, auth.Username, ws)
		for {
			var data internal.ClientData
			err = websocket.JSON.Receive(ws, &data)
			if err != nil {
				break
			}

			err = internal.MsgBoxes[boxRes.MsgBoxID].Broadcast(auth.UserID, &data)
			if err != nil {
				break
			}
		}
		if auth.UserID == boxRes.OwnerID {
			c.Logger().Print(fmt.Sprintf("deleted box %v", boxOfUserResponse.ID))
			err = UncheckDeleteBox(c, boxRes.ID)
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					if he.Code == http.StatusNotFound {
						err = nil
					}
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return err
}

func GetMovie(c echo.Context) error {
	movieId, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	url := fmt.Sprintf("%s/api/movie/%d", movieServiceAddr, movieId)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to get movie by id")
	}

	var mRes internal.MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&mRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie")
	}

	return c.JSON(http.StatusOK, mRes)
}

func GetMovies(c echo.Context) error {
	url := fmt.Sprintf("%s/api/movie", movieServiceAddr)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to get movies")
	}

	var mRes []internal.MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&mRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movies")
	}

	return c.JSON(http.StatusOK, mRes)
}

func ContainsUser(c echo.Context) error {
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

	boxId, err := strconv.Atoi(c.Param("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	containsURL := fmt.Sprintf("%s/api/box/%d/exists/%d", movieServiceAddr, boxId, auth.UserID)
	resp, err := http.Get(containsURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed checking user existence in box")
	}

	var bRes internal.BooleanResponse
	if err := json.NewDecoder(resp.Body).Decode(&bRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user existence in box response")
	}

	return c.JSON(http.StatusOK, bRes)
}

func JoinBox(c echo.Context) error {
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
		return c.String(http.StatusConflict, "User already in a box")
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
	lock.Lock()
	defer lock.Unlock()

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

	// Retrieve the box details
	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve movie box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse movie box response")
	}

	if boxRes.OwnerID != auth.UserID {
		return c.String(http.StatusNotFound, "Can't find target movie box")
	}

	// Delete box
	deleteBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
	req, _ := http.NewRequest(http.MethodDelete, deleteBoxURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.String(resp.StatusCode, fmt.Sprint("Failed to delete movie box ", string(bodyBytes)))
	}

	// Delete the its message box
	deleteMessageBoxURL := fmt.Sprintf("%s/api/msg/box/%d", messageServiceAddr, boxRes.MsgBoxID)
	req, _ = http.NewRequest(http.MethodDelete, deleteMessageBoxURL, nil)
	resp, err = client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, "Failed to delete message box")
	}

	// Remove the message box websocket client
	internal.MsgBoxes[boxRes.MsgBoxID].Close()
	delete(internal.MsgBoxes, boxRes.MsgBoxID)

	return c.NoContent(http.StatusOK)
}

func UncheckDeleteBox(c echo.Context, boxId int) error {
	lock.Lock()
	defer lock.Unlock()

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

	// Delete Movie Box
	deleteBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxId)
	req, _ := http.NewRequest(http.MethodDelete, deleteBoxURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	// Delete Message Box
	deleteMsgBoxURL := fmt.Sprintf("%s/api/msg/box/%d", messageServiceAddr, boxRes.MsgBoxID)
	req, _ = http.NewRequest(http.MethodDelete, deleteMsgBoxURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	// Remove the message box websocket client
	internal.MsgBoxes[boxRes.MsgBoxID].Close()
	delete(internal.MsgBoxes, boxRes.MsgBoxID)

	return nil
}

func LeaveBox(c echo.Context) error {
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

	// Get Box of User
	boxOfUserURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(boxOfUserURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusBadRequest, "Failed to retrieve user's box")
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
	}

	// Get Box Details
	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse box details")
	}

	if boxRes.OwnerID == auth.UserID {
		return c.String(http.StatusBadRequest, "You are the owner")
	}

	// Remove User from Box
	removeUserURL := fmt.Sprintf("%s/api/box/%d/remove/%d", movieServiceAddr, boxOfUserResponse.ID, auth.UserID)
	req, _ := http.NewRequest(http.MethodDelete, removeUserURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to remove user from box")
	}

	// Remove User from Message Box
	removeMsgUserURL := fmt.Sprintf("%s/api/msg/box/%d/remove/%d", messageServiceAddr, boxRes.MsgBoxID, auth.UserID)
	req, _ = http.NewRequest(http.MethodDelete, removeMsgUserURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to remove user from message box")
	}

	internal.MsgBoxes[boxRes.MsgBoxID].Remove(auth.UserID)

	return c.NoContent(http.StatusOK)
}

func KickFromBox(c echo.Context) error {
	kickId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse user id to be kicked")
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

	if kickId == auth.UserID {
		return c.String(http.StatusBadRequest, "Can not kick yourself")
	}

	// Get Box of User
	boxOfUserURL := fmt.Sprintf("%s/api/box/user/%d", movieServiceAddr, auth.UserID)
	resp, err := http.Get(boxOfUserURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusBadRequest, "Failed to retrieve user's box")
	}
	defer resp.Body.Close()

	var boxOfUserResponse internal.IdentifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxOfUserResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user's box response")
	}

	// Get Box Details
	getBoxURL := fmt.Sprintf("%s/api/box/%d", movieServiceAddr, boxOfUserResponse.ID)
	resp, err = http.Get(getBoxURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to retrieve box details")
	}
	defer resp.Body.Close()

	var boxRes internal.BoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxRes); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse box details")
	}

	if boxRes.OwnerID != auth.UserID {
		return c.String(http.StatusUnauthorized, "Kicking is not allowed, you are unauthorized")
	}

	// Remove User from Box
	removeUserURL := fmt.Sprintf("%s/api/box/%d/remove/%d", movieServiceAddr, boxOfUserResponse.ID, kickId)
	req, _ := http.NewRequest(http.MethodDelete, removeUserURL, nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to remove user from box")
	}

	// Remove User from Message Box
	removeMsgUserURL := fmt.Sprintf("%s/api/msg/box/%d/remove/%d", messageServiceAddr, boxRes.MsgBoxID, kickId)
	req, _ = http.NewRequest(http.MethodDelete, removeMsgUserURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.String(http.StatusInternalServerError, "Failed to remove user from message box")
	}

	internal.MsgBoxes[boxRes.MsgBoxID].Remove(kickId)

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

	internal.AppendNewMsgBox(msgBoxResponse.ID, auth.UserID, boxResponse.ID)
	return c.JSON(http.StatusOK, boxResponse)
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

	data := &internal.ClientBoxData{
		BoxId:   boxOfUserResponse.ID,
		IsOwner: auth.UserID == internal.MsgBoxes[boxRes.MsgBoxID].OwnerId,
	}

	return c.JSON(http.StatusOK, data)
}

func Authenticate(c echo.Context) (*internal.AuthReponse, error) {
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, echo.NewHTTPError(resp.StatusCode, body)
	}

	var authResponse internal.AuthReponse
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		return nil, err
	}

	return &authResponse, nil
}
