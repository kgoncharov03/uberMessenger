package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
)

type Endpoints struct {
	UserDAO *users.DAO
	ChatDAO *chats.DAO
	MessageDAO *messages.DAO
}

type TokenParams struct {
	Token string `json:"token"`
}

func (e* Endpoints) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	spew.Dump(r)
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

func (e *Endpoints) AddMessageParams (w http.ResponseWriter, r *http.Request)  {
	decoder := json.NewDecoder(r.Body)
	var params AddChatParams
	err := decoder.Decode(&params)
	if err!=nil {
		e.handleError(w, err)
		return
	}


}

func (e *Endpoints) AddChatHandler(w http.ResponseWriter, r *http.Request) {
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

	err=e.MessageDAO.AddMessage(context.TODO(), msg)
	if err!=nil {
		e.handleError(w, err)
		return
	}
}


func (e *Endpoints) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()
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

	chats,err:= e.ChatDAO.GetChatsByUser(ctx, userID)
	if err!=nil {
		e.handleError(w, err)
		return
	}


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
		http.Error(w, err.Error(), 500)
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

	e:=&Endpoints{
		UserDAO:userDAO,
		ChatDAO:chatDAO,
		MessageDAO:messageDAO,
	}

	router := mux.NewRouter()
	router.Handle("/getToken/", http.HandlerFunc(e.GetTokenHandler)).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/users/", e.Middleware(http.HandlerFunc(e.GetUserByIDHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/chats/", e.Middleware(http.HandlerFunc(e.GetChatsByUser))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/messages/", e.Middleware(http.HandlerFunc(e.GetMessages))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/addChat", e.Middleware(http.HandlerFunc(e.AddChatHandler))).Methods(http.MethodPost, http.MethodOptions)


	http.Handle("/",router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}