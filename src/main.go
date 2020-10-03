package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"uberMessenger/src/common"
	"uberMessenger/src/users"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Endpoints struct {
	UserDAO *users.DAO
}




func (e *Endpoints) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

func (e *Endpoints) handleError(w http.ResponseWriter, err error) {
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		panic(err)
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

	e:=&Endpoints{UserDAO:userDAO}

	router := mux.NewRouter()
	router.Path("/users/").Queries("id", "{id}").HandlerFunc(e.GetUserByIDHandler)
	http.Handle("/",router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}