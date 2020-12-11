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
	"uberMessenger/src/storage"
	"uberMessenger/src/users"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/websocket"
)

type Endpoints struct {
	UserDAO       *users.DAO
	ChatDAO       *chats.DAO
	MessageDAO    *messages.DAO
	AttachmentDAO *storage.DAO

	msgSockets  map[primitive.ObjectID]*websocket.Conn
	msgUpgrader websocket.Upgrader
	msgChannel  chan *messages.Message

	chatSockets  map[primitive.ObjectID]*websocket.Conn
	chatUpgrader websocket.Upgrader
	chatChannel  chan *chats.Chat
}

func NewEndpoints(
	UserDAO *users.DAO,
	ChatDAO *chats.DAO,
	MessageDAO *messages.DAO,
	AttachmentDAO *storage.DAO,
) *Endpoints {
	endpoints := &Endpoints{
		UserDAO:       UserDAO,
		ChatDAO:       ChatDAO,
		MessageDAO:    MessageDAO,
		AttachmentDAO: AttachmentDAO,

		msgSockets: make(map[primitive.ObjectID]*websocket.Conn),
		msgUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		msgChannel: make(chan *messages.Message, 100),

		chatSockets: make(map[primitive.ObjectID]*websocket.Conn),
		chatUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		chatChannel: make(chan *chats.Chat, 100),
	}

	go endpoints.processMessages()
	go endpoints.processChats()

	return endpoints
}

type TokenParams struct {
	Token string `json:"token"`
}

func (e *Endpoints) processMessages() {
	ctx := context.Background()
	for {
		msg := <-e.msgChannel
		chatID := msg.ChatID
		chat, err := e.ChatDAO.GetChatByID(ctx, chatID)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			continue
		}

		for _, userID := range chat.Users {
			socket, ok := e.msgSockets[userID]
			if !ok {
				continue
			}
			err := socket.WriteJSON(&msg)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				socket.Close()
				delete(e.msgSockets, userID)
			}
		}
	}
}

func (e *Endpoints) processChats() {
	for {
		chat := <-e.chatChannel
		for _, userID := range chat.Users {
			socket, ok := e.chatSockets[userID]
			if !ok {
				continue
			}
			err := socket.WriteJSON(&chat)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				socket.Close()
				delete(e.chatSockets, userID)
			}
		}
	}
}

func (e *Endpoints) GetChatSocketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		e.handleError(w, err)
		return
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		e.handleError(w, err)
		return
	}

	e.chatSockets[userID] = ws
}

func (e *Endpoints) GetMessageSocketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
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

func (e *Endpoints) GetUsersByChatHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	chatIDParam := r.URL.Query().Get("chatId")

	chatID, err := primitive.ObjectIDFromHex(chatIDParam)
	if err != nil {
		e.handleError(w, err)
		return
	}

	chat, err := e.ChatDAO.GetChatByID(ctx, chatID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	var users []*users.User

	for _, userID := range chat.Users {
		user, err := e.UserDAO.GetUserByID(ctx, userID)
		if err != nil {
			e.handleError(w, err)
			return
		}

		users = append(users, user)
	}

	bytes, err := json.Marshal(users)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}
func (e *Endpoints) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := e.getUserIDFromToken(r)
	if err != nil {
		e.handleError(w, errors.New("unauthorized"))
		return
	}

	user, err := e.UserDAO.GetUserByID(context.Background(), userID)
	bytes, err := json.Marshal(user)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)

}
func (e *Endpoints) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)

	nickname := r.URL.Query().Get("nickname")
	password := r.URL.Query().Get("password")

	user, err := e.UserDAO.GetUserByNickname(context.TODO(), nickname)
	if err != nil {
		e.handleError(w, err)
		return
	}

	if strings.Compare(password, user.Password) != 0 {
		e.handleError(w, errors.New("unauthorized"))
		return
	}

	token, err := auth.CreateToken(user.ID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	tokenStruct := &TokenParams{Token: token}

	bytes, err := json.Marshal(tokenStruct)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.Write(bytes)
}

func (e *Endpoints) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.writeHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
		}

		_, err := e.getUserIDFromToken(r)
		if err != nil {
			e.handleError(w, err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (e *Endpoints) getUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return primitive.ObjectID{}, errors.New("unauthorized")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return primitive.ObjectID{}, errors.New("unauthorized")
	}

	return auth.CheckToken(headerParts[1])
}

func (e *Endpoints) writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}

type AddChatParams struct {
	Users []string `json:"users"`
	Name  string   `json:"name"`
}

type AddMessageParams struct {
	FromID         string                   `json:"fromID"`
	ChatID         string                   `json:"chatID"`
	Text           string                   `json:"text"`
	AttachmentLink *messages.AttachmentLink `json:"attachmentLink,omitempty"`
}

func (e *Endpoints) AddMessageHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params AddMessageParams
	err := decoder.Decode(&params)
	if err != nil {
		e.handleError(w, err)
		return
	}

	fromID, err := primitive.ObjectIDFromHex(params.FromID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	chatID, err := primitive.ObjectIDFromHex(params.ChatID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	msg := &messages.Message{
		ID:             primitive.NewObjectID(),
		From:           fromID,
		ChatID:         chatID,
		Text:           params.Text,
		Time:           time.Now().UnixNano(),
		AttachmentLink: params.AttachmentLink,
	}

	err = e.MessageDAO.AddMessage(context.Background(), msg)
	if err != nil {
		e.handleError(w, err)
		return
	}

	e.msgChannel <- msg
}

func (e *Endpoints) AddChatHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params AddChatParams
	err := decoder.Decode(&params)
	if err != nil {
		e.handleError(w, err)
		return
	}
	var userIDs []primitive.ObjectID
	for _, id := range params.Users {
		idBSON, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			e.handleError(w, err)
			return
		}
		userIDs = append(userIDs, idBSON)
	}

	chat := &chats.Chat{
		ID:              primitive.NewObjectID(),
		LastMessageTime: time.Now().UnixNano(),
		Users:           userIDs,
		Name:            params.Name,
	}

	err = e.ChatDAO.AddChat(context.Background(), chat)
	if err != nil {
		e.handleError(w, err)
		return
	}

	e.chatChannel <- chat
}

type AddAttachmentResponse struct {
	ID primitive.ObjectID `json:"id"`
}

func (e *Endpoints) UploadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.handleError(w, err)
		return
	}

	att := &storage.Attachment{
		ID:      primitive.NewObjectID(),
		Content: bytes,
	}

	err = e.AttachmentDAO.InsertAttachment(ctx, att)
	if err != nil {
		e.handleError(w, err)
		return
	}

	resp := &AddAttachmentResponse{ID: att.ID}

	bytes, err = json.Marshal(resp)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id := r.URL.Query().Get("id")
	attID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		e.handleError(w, err)
		return
	}

	att, err := e.AttachmentDAO.GetAttachmentByID(ctx, attID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(att.Content)
}

func (e *Endpoints) GetUserByNicknameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	nickname := r.URL.Query().Get("nickname")

	user, err := e.UserDAO.GetUserByNickname(ctx, nickname)
	if err != nil {
		e.handleError(w, err)
		return
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx := context.Background()
	id := r.URL.Query().Get("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		e.handleError(w, err)
		return
	}

	user, err := e.UserDAO.GetUserByID(ctx, userID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetChatsByUser(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx := context.TODO()
	id := r.URL.Query().Get("userId")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		e.handleError(w, err)
		return
	}

	chats, err := e.ChatDAO.GetChatsByUser(ctx, userID)
	if err != nil {
		e.handleError(w, err)
		return
	}

	for i := 0; i < len(chats); i++ {
		chats[i], err = e.enrichChat(ctx, chats[i])
		if err != nil {
			e.handleError(w, err)
			return
		}
	}

	bytes, err := json.Marshal(chats)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetMessages(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx := context.TODO()

	chatID, err := primitive.ObjectIDFromHex(r.URL.Query().Get("chatId"))
	if err != nil {
		spew.Dump(err)
		e.handleError(w, err)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		e.handleError(w, err)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		e.handleError(w, err)
		return
	}

	msgs, err := e.MessageDAO.GetMessagesByChat(ctx, chatID, limit, offset)
	if err != nil {
		e.handleError(w, err)
		return
	}

	bytes, err := json.Marshal(msgs)
	if err != nil {
		e.handleError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

type RegisterParams struct {
	FirstName  string `json:"firstName"`
	SecondName string `json:"secondName"`
	NickName   string `json:"nickName"`
	Password   string ` json:"password"`
}

func (e *Endpoints) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("recieve register query")
	e.writeHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
	}
	decoder := json.NewDecoder(r.Body)
	var params RegisterParams
	err := decoder.Decode(&params)
	if err != nil {
		e.handleError(w, err)
		return
	}
	ctx := context.Background()

	exists, err := e.UserDAO.NickNameExists(ctx, params.NickName)
	if err != nil {
		e.handleError(w, err)
		return
	}

	if exists {
		http.Error(w, "nickname already exists", http.StatusBadRequest)
		return
	}

	newUser := &users.User{
		ID:         primitive.NewObjectID(),
		FirstName:  params.FirstName,
		SecondName: params.SecondName,
		NickName:   params.NickName,
		Password:   params.Password,
	}

	err = e.UserDAO.InsertUser(ctx, newUser)
	if err != nil {
		e.handleError(w, err)
		return
	}
	bytes, err := json.Marshal(newUser)
	if err != nil {
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

func (e *Endpoints) enrichChat(ctx context.Context, chat *chats.Chat) (*chats.Chat, error) {
	msg, err := e.MessageDAO.GetMessagesByChat(ctx, chat.ID, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(msg) == 0 {
		return chat, nil
	}

	if msg[0].AttachmentLink != nil && msg[0].Text == "" {
		chat.LastMessage = "attachment"
	} else {
		chat.LastMessage = msg[0].Text
	}

	chat.LastMessageTime = msg[0].Time

	return chat, nil
}

func main() {
	ctx := context.TODO()

	client, err := common.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	userDAO, err := users.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	chatDAO, err := chats.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	messageDAO, err := messages.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	attDAO, err := storage.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	e := NewEndpoints(userDAO, chatDAO, messageDAO, attDAO)

	router := mux.NewRouter()
	router.Handle("/register/", http.HandlerFunc(e.RegisterHandler)).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/getToken/", http.HandlerFunc(e.GetTokenHandler)).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/users/", e.Middleware(http.HandlerFunc(e.GetUserByIDHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/usersByChat/", e.Middleware(http.HandlerFunc(e.GetUsersByChatHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/me/", e.Middleware(http.HandlerFunc(e.GetMe))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/usersByNickname/", e.Middleware(http.HandlerFunc(e.GetUserByNicknameHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/chats/", e.Middleware(http.HandlerFunc(e.GetChatsByUser))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/messages/", e.Middleware(http.HandlerFunc(e.GetMessages))).Methods(http.MethodGet, http.MethodOptions)

	router.Handle("/addChat", e.Middleware(http.HandlerFunc(e.AddChatHandler))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/addMessage", e.Middleware(http.HandlerFunc(e.AddMessageHandler))).Methods(http.MethodPost, http.MethodOptions)

	router.Handle("/addAttachment", e.Middleware(http.HandlerFunc(e.UploadAttachmentHandler))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/attachments/", e.Middleware(http.HandlerFunc(e.GetAttachmentHandler))).Methods(http.MethodGet, http.MethodOptions)

	router.Handle("/messageWs/", http.HandlerFunc(e.GetMessageSocketHandler))
	router.Handle("/chatWs/", http.HandlerFunc(e.GetChatSocketHandler))

	http.Handle("/", router)

	fmt.Println("Server is listening... on port 8181")
	log.Fatal(http.ListenAndServe(":8181", nil))
}
