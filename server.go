package main

import (
	"net/http"
	"html/template"
	"database/sql"
	"io"
	"fmt"
)

var (
	maxQuestions = 50
	db *sql.DB
)

type RegisterResponse struct {
	SessionId string
}

type QuestionResponse struct {
	Question string
	Answer string
}

type QuestionResult struct {
	Expiry int
	Question []struct {
		File string
		Type string
		Description string
		SubQuestion []struct {
			Question string
			Type string
			Choice map[int] string
		}
	}
}

func displayWebPage(w http.ResponseWriter, file string) {
	t, _ := template.ParseFiles(file)
	t.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "Register.html")
}

func jqueryHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "js/jquerymin.js")
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Print("sdhf")
	io.WriteString(w, "asd")
}

func main() {
	http.HandleFunc("/test", test)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/js/jquerymin.js", jqueryHandler)
	http.ListenAndServe(":9000", nil)
}