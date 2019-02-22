package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	cache "github.com/patrickmn/go-cache"

	"github.com/gorilla/websocket"
)

var c *cache.Cache

func main() {
	c = cache.New(10*time.Second, 10*time.Second)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("here")
		fmt.Fprintln(w, "hey")
	}).Methods(http.MethodPost)

	http.Handle("/", r)

	http.HandleFunc("/ws", handleWs)

	http.HandleFunc("/user", handleCreateToken)

	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 256,
	CheckOrigin: func(r *http.Request) bool {
		// 	// if "ws://"+r.Host == r.Header.Get("Origin") {
		return true
		// 	// }
		// 	// return false
	},
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	subprotocols := websocket.Subprotocols(r)
	if len(subprotocols) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := subprotocols[0]
	if _, ok := c.Get(token); !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	c.Delete(token)

	_, _ = upgrader.Upgrade(w, r, http.Header{"Sec-WebSocket-Protocol": []string{token}})
}

type user struct {
	Id   int    `json: "id"`
	Name string `json: "name"`
}

func handleCreateToken(w http.ResponseWriter, r *http.Request) {
	u := readBody(r.Body)
	if u == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token := createToken()

	if err := c.Add(token, strconv.Itoa(u.Id), cache.DefaultExpiration); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Authorization", token)
}

func readBody(body io.ReadCloser) *user {
	if body == nil {
		return nil
	}

	decoder := json.NewDecoder(body)

	var u user
	if err := decoder.Decode(&u); err != nil {
		return nil
	}

	return &u
}

func createToken() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
