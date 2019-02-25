package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/render"

	"github.com/adrianbrad/sandbox/mux+orm/config"
	"github.com/go-chi/chi"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	config := loadConfig()

	// * initDB is called first, then the return value is assigned to the defer
	defer initDB(config.Database)()

	r := chi.NewRouter()
	r.Use(logging)

	r.Get("/withlocator/{locator}", func(w http.ResponseWriter, req *http.Request) {
		vars := chi.URLParam(req, "locator")

		log.Println("locator", vars)
	})

	r.Post("/test/users", addUser)

	r.Get("/test/users", returnUsers)

	r.Put("/test/users/{id}", updateUser)

	r.Delete("/test/users/{id}", deleteUser)
	log.Fatal(http.ListenAndServe(":8080", r))
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var updatedUser User
	err := render.DecodeJSON(r.Body, &updatedUser)
	userId := chi.URLParam(r, "id")

	updateUserStatement := `
	UPDATE users
	SET name = $1
	WHERE id = $2
	RETURNING id;`

	var id int
	err = db.QueryRow(updateUserStatement, updatedUser.Name, userId).Scan(&id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println(id)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	deleteUserStatement := `
	DELETE FROM users
	WHERE id = $1;`

	userId := chi.URLParam(r, "id")

	res, err := db.Exec(deleteUserStatement, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := render.DecodeJSON(r.Body, &u); err != nil {
		return
	}
	insertUserStatement := `
	INSERT INTO users (name)
	VALUES ($1)
	RETURNING id`

	id := 0
	if err := db.QueryRow(insertUserStatement, u.Name).Scan(&id); err != nil {
		return
	}
	render.JSON(w, r, id)
}

func returnUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
	SELECT 
		id, name
	FROM users
	`)
	if err != nil {
		log.Println("Query error :", err)
		return
	}
	defer rows.Close()
	var users []User

	for rows.Next() {
		test := User{}
		err = rows.Scan(
			&test.ID,
			&test.Name,
		)
		if err != nil {
			log.Println("Mapping error:", err)
			return
		}
		users = append(users, test)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Reading rows error:", err)
	}
	render.JSON(w, r, users)
}

func loadConfig() config.Configuration {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	var config config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return config
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func initDB(config config.DatabaseConfiguration) func() error {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port,
		config.User, config.Pass, config.Name)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected to the database!")

	return db.Close
}
