package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/adrianbrad/sandbox/mux+orm/config"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

var db *sql.DB

type test struct {
	name string
}

func main() {
	config := loadConfig()

	// * initDB is called first, then the return value is assigned to the defer
	defer initDB(config.Database)()

	r := mux.NewRouter()
	r.Use(logging)

	r.HandleFunc("/{locator}/users", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		log.Println("users", vars)
	}).Methods(http.MethodGet, http.MethodPost, http.MethodPut)

	r.HandleFunc("/{locator}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		log.Println("locator", vars)
	}).Methods(http.MethodGet)

	r.HandleFunc("/users/test", returnTest)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func returnTest(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
	SELECT 
		name
	FROM test
	`)
	if err != nil {
		log.Println("Query error :", err)
		return
	}
	defer rows.Close()
	var tests []test

	for rows.Next() {
		log.Println("one time")
		test := test{}
		err = rows.Scan(
			&test.name,
		)
		if err != nil {
			log.Println("Mapping error:", err)
			return
		}
		tests = append(tests, test)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Reading rows error:", err)
	}
	fmt.Fprint(w, tests)
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
		log.Println(r.RequestURI)
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
