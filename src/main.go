package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"uberMessenger/src/auth"
	"uberMessenger/src/chats"
	"uberMessenger/src/common"
	"uberMessenger/src/messages"
	"uberMessenger/src/users"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/websocket"
)

type Endpoints struct {
	UserDAO *users.DAO
	ChatDAO *chats.DAO
	MessageDAO *messages.DAO

	msgSockets  map[primitive.ObjectID]*websocket.Conn
	msgUpgrader websocket.Upgrader
	msgChannel  chan *messages.Message
}

func NewEndpoints(
	UserDAO *users.DAO,
	ChatDAO *chats.DAO,
	MessageDAO *messages.DAO,
	) *Endpoints {
	endpoints:=&Endpoints{
		UserDAO:    UserDAO,
		ChatDAO:    ChatDAO,
		MessageDAO: MessageDAO,
		msgSockets: make(map[primitive.ObjectID]*websocket.Conn),
		msgUpgrader:   websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		msgChannel: make(chan *messages.Message,100),
	}

	go endpoints.processMessages()
	return endpoints
}

type TokenParams struct {
	Token string `json:"token"`
}

func (e *Endpoints) processMessages() {
	ctx:=context.Background()
	for {
		msg:= <-e.msgChannel
		fmt.Println(msg)
		chatID:=msg.ChatID
		chat,err:=e.ChatDAO.GetChatByID(ctx, chatID)
		if err!=nil {
			log.Printf("Websocket error: %s", err)
			continue
		}

		for _, userID:=range chat.Users{
			socket, ok:=e.msgSockets[userID]
			if !ok {
				continue
			}
			err:=socket.WriteJSON(&msg)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				socket.Close()
				delete(e.msgSockets, userID)
			}
		}
	}
}

func (e *Endpoints) GetMessageSocketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userID,err:=primitive.ObjectIDFromHex(id)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		e.handleError(w, err)
		return
	}

	e.msgSockets[userID] = ws
}

func (e *Endpoints) GetMe(w http.ResponseWriter, r *http.Request) {
	userID,err:=e.getUserIDFromToken(r)
	if err!=nil {
		e.handleError(w, errors.New("unauthorized"))
		return
	}

	user,err:=e.UserDAO.GetUserByID(context.Background(), userID)
	bytes, err:=json.Marshal(user)
	if err!=nil {
		e.handleError(w, err)
		return
	}


	w.WriteHeader(200)
	w.Write(bytes)

}
func (e* Endpoints) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)

	nickname:=r.URL.Query().Get("nickname")
	password:=r.URL.Query().Get("password")

	user,err:=e.UserDAO.GetUserByNickname(context.TODO(), nickname)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	if strings.Compare(password, user.Password) !=0 {
		e.handleError(w, errors.New("unauthorized"))
		return
	}

	token,err := auth.CreateToken(user.ID)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	tokenStruct:=&TokenParams{Token:token}

	bytes, err:=json.Marshal(tokenStruct)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	w.Write(bytes)
}

func (e* Endpoints) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spew.Dump(r)
		e.writeHeaders(w)
		if r.Method==http.MethodOptions {
			w.WriteHeader(200)
		}

		_, err:=e.getUserIDFromToken(r)
		if err!=nil {
			e.handleError(w, err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (e *Endpoints) getUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
	authHeader:= r.Header.Get("Authorization")

	if authHeader == "" {
		return primitive.ObjectID{}, errors.New("unauthorized")
	}

	headerParts:=strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return primitive.ObjectID{}, errors.New("unauthorized")
	}

	return auth.CheckToken(headerParts[1])
}

func(e *Endpoints) writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}


type AddChatParams struct {
	Users []string `json:"users"`
}

type AddMessageParams struct {
	FromID string `json:"fromID"`
	ChatID string `json:"chatID"`
	Text string `json:"text"`
}

func (e *Endpoints) AddMessageHandler (w http.ResponseWriter, r *http.Request)  {
	decoder := json.NewDecoder(r.Body)
	var params AddMessageParams
	err := decoder.Decode(&params)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	fromID, err:=primitive.ObjectIDFromHex(params.FromID)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	chatID, err:=primitive.ObjectIDFromHex(params.ChatID)
	if err!=nil {
		e.handleError(w, err)
		return
	}
	
	msg:=&messages.Message{
		ID:     primitive.NewObjectID(),
		From:   fromID,
		ChatID: chatID,
		Text:   params.Text,
		Time:   time.Now(),
	}

	err = e.MessageDAO.AddMessage(context.Background(), msg)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	e.msgChannel <- msg
	fmt.Print("KEK")
}

func (e *Endpoints) AddChatHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params AddChatParams
	err := decoder.Decode(&params)
	if err!=nil {
		e.handleError(w, err)
		return
	}
	var userIDs []primitive.ObjectID
	for _, id:=range params.Users {
		idBSON,err:=primitive.ObjectIDFromHex(id)
		if err!=nil {
			e.handleError(w, err)
			return
		}
		userIDs = append(userIDs, idBSON)
	}

	chat:=&chats.Chat{
		ID:              primitive.NewObjectID(),
		LastMessageTime: time.Now(),
		Users:           userIDs,
		Name:            "",
	}

	err=e.ChatDAO.AddChat(context.Background(), chat)
	if err!=nil {
		e.handleError(w, err)
		return
	}
}

func (e *Endpoints) GetUserByNicknameHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.Background()
	nickname := r.URL.Query().Get("nickname")

	user, err:=e.UserDAO.GetUserByNickname(ctx, nickname)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	bytes, err:=json.Marshal(user)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}


func (e *Endpoints) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.Background()
	id := r.URL.Query().Get("id")
	userID,err:=primitive.ObjectIDFromHex(id)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	user,err:= e.UserDAO.GetUserByID(ctx, userID)
	if err!=nil {
		e.handleError(w, err)
		return
	}


	bytes, err:=json.Marshal(user)
	if err!=nil {
		e.handleError(w, err)
		return
	}


	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetChatsByUser(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()
	id := r.URL.Query().Get("userId")
	userID,err:=primitive.ObjectIDFromHex(id)
	if err!=nil {
		e.handleError(w, err)
		return
	}
	spew.Dump(userID)

	chats,err:= e.ChatDAO.GetChatsByUser(ctx, userID)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	spew.Dump(chats)
	bytes, err:=json.Marshal(chats)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetMessages(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()

	chatID,err :=primitive.ObjectIDFromHex(r.URL.Query().Get("chatId"))
	if err!=nil {
		spew.Dump(err)
		e.handleError(w, err)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err!=nil {
		e.handleError(w, err)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err!=nil {
		e.handleError(w, err)
		return
	}

	msgs, err:=e.MessageDAO.GetMessagesByChat(ctx, chatID, limit, offset)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	bytes, err:=json.Marshal(msgs)
	if err!=nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) handleError(w http.ResponseWriter, err error) {
		log.Print(err)
		http.Error(w, err.Error(), 500)
}

func rootHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadFile("index.html")
	if err != nil {
		http.Error(writer, "Internal error" + err.Error(), 400)
	}
	writer.Write(body)
}

func main() {
	ctx:=context.TODO()

	client, err := common.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	userDAO,err:= users.NewDAO(ctx, client)
	if err!=nil {
		panic(err)
	}

	chatDAO, err:=chats.NewDAO(ctx, client)
	if err!=nil {
		panic(err)
	}

	messageDAO, err:=messages.NewDAO(ctx, client)
	if err!=nil {
		panic(err)
	}

	e:=NewEndpoints(userDAO, chatDAO, messageDAO)

	router := mux.NewRouter()
	router.Handle("/getToken/", http.HandlerFunc(e.GetTokenHandler)).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/users/", e.Middleware(http.HandlerFunc(e.GetUserByIDHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/me/", e.Middleware(http.HandlerFunc(e.GetMe))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/usersByNickname/", e.Middleware(http.HandlerFunc(e.GetUserByNicknameHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/chats/", e.Middleware(http.HandlerFunc(e.GetChatsByUser))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/messages/", e.Middleware(http.HandlerFunc(e.GetMessages))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/addChat", e.Middleware(http.HandlerFunc(e.AddChatHandler))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/addMessage", e.Middleware(http.HandlerFunc(e.AddMessageHandler))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/messageWs/", http.HandlerFunc(e.GetMessageSocketHandler))
	//router.Handle("/chatWs/")
	router.Handle("/test/", http.HandlerFunc(rootHandler))


	http.Handle("/",router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}