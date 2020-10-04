package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"uberMessenger/src/chats"
	"uberMessenger/src/common"
	"uberMessenger/src/messages"
	"uberMessenger/src/users"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Endpoints struct {
	UserDAO *users.DAO
	ChatDAO *chats.DAO
	MessageDAO *messages.DAO
}

func(e *Endpoints) writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func (e *Endpoints) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()
	id := r.URL.Query().Get("id")
	userID,err:=primitive.ObjectIDFromHex(id)
	e.handleError(w, err)

	user,err:= e.UserDAO.GetUserByID(ctx, userID)
	e.handleError(w, err)


	bytes, err:=json.Marshal(user)
	e.handleError(w, err)

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetChatsByUser(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()
	id := r.URL.Query().Get("userId")
	userID,err:=primitive.ObjectIDFromHex(id)
	e.handleError(w, err)

	chats,err:= e.ChatDAO.GetChatsByUser(ctx, userID)
	e.handleError(w, err)


	bytes, err:=json.Marshal(chats)
	e.handleError(w, err)

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) GetMessages(w http.ResponseWriter, r *http.Request) {
	e.writeHeaders(w)
	ctx:=context.TODO()

	chatID,err :=primitive.ObjectIDFromHex( r.URL.Query().Get("chatId"))
	e.handleError(w, err)

	limit,err := strconv.Atoi(r.URL.Query().Get("limit"))
	e.handleError(w, err)

	offset,err := strconv.Atoi(r.URL.Query().Get("offset"))
	e.handleError(w, err)

	msgs, err:=e.MessageDAO.GetMessagesByChat(ctx, chatID, limit, offset)
	e.handleError(w, err)

	bytes, err:=json.Marshal(msgs)
	e.handleError(w, err)

	w.WriteHeader(200)
	w.Write(bytes)
}

func (e *Endpoints) handleError(w http.ResponseWriter, err error) {
	if err != nil {
		w.WriteHeader(500)
		e:= struct {
			err string
		}{
			err:err.Error(),
		}

		bytes,_:=json.Marshal(e)

		w.WriteHeader(500)
		w.Write(bytes)
	}
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
	router.Path("/users/").Queries("id", "{id}").HandlerFunc(e.GetUserByIDHandler)
	router.Path("/chats/").Queries("userId", "{userId}").HandlerFunc(e.GetChatsByUser)
	router.Path("/messages/").HandlerFunc(e.GetMessages)
	http.Handle("/",router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}