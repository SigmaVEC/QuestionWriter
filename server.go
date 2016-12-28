package main

import (
	"net/http"
	"html/template"
	"database/sql"
	//"io"
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


func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"sdhf")
	//io.WriteString(w, "asd")
}

func main() {
	fs := http.FileServer(http.Dir("./static")) //added this line
	http.Handle("/", fs)                        //added this line
	http.HandleFunc("/test", test)
	http.HandleFunc("/register", registerHandler)
	http.ListenAndServe(":9000", nil)
}
