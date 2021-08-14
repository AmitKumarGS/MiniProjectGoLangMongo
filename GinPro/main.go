package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var session, _ = mgo.Dial("127.0.0.1")
var c = session.DB("Bookstore").C("Books")

type ToDoItem struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	Date   time.Time
	Book   string
	Author string
	Done   bool
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	_ = c.Insert(ToDoItem{
		bson.NewObjectId(),
		time.Now(),
		r.FormValue("Book"),
		r.FormValue("Author"),
		false,
	})

	result := ToDoItem{}
	_ = c.Find(bson.M{"book": r.FormValue("Book")}).One(&result)
	json.NewEncoder(w).Encode(result)
}

func GetBooks(w http.ResponseWriter, r *http.Request) {
	var res []ToDoItem

	vars := mux.Vars(r)
	id := vars["id"]
	if id != "" {
		res = GetByID(id)
	} else {
		_ = c.Find(nil).All(&res)
	}

	json.NewEncoder(w).Encode(res)
}

func GetByID(id string) []ToDoItem {
	var result ToDoItem
	var res []ToDoItem
	_ = c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	res = append(res, result)
	return res
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	err := c.Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"done": true}})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": `+err.Error()+`}`)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := c.RemoveId(bson.ObjectIdHex(id))
	if err == mgo.ErrNotFound {
		json.NewEncoder(w).Encode(err.Error())
	} else {
		io.WriteString(w, "{result: 'OK'}")
	}
}

func main() {
	session.SetMode(mgo.Monotonic, true)
	fmt.Println("8000")
	defer session.Close()
	router := mux.NewRouter()
	router.HandleFunc("/todo", GetBooks).Methods("GET")
	router.HandleFunc("/todo/{id}", GetBooks).Methods("GET")
	router.HandleFunc("/todo", CreateBook).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateBook).Methods("PATCH")
	router.HandleFunc("/todo/{id}", DeleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", router))
}
