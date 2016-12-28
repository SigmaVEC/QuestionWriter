package main

import (
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"crypto/rand"
	"encoding/base64"
	"io"
	"fmt"
)

var (
	maxQuestions = 50
	sessionKeyLength=50
	emptyJson = "{}"
	db *sql.DB
	user string = "test"
	password string = "test"
	database string = "QuestionWriter"
)

type RegisterRequest struct {
	StudentId string
	Name string
	AcademicYear int
	Department string
	Section string
}

type RegisterResponse struct {
	SessionId string
}

type QuestionUpdateResponse struct {
	Question string
	Answer string
}

type QuestionListResponse struct {
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	studentJson := r.FormValue("studentJson")
	data := []byte(studentJson)
	var student RegisterRequest
	var reply RegisterResponse
	err := json.Unmarshal(data, &student)

	if err == nil {
		b := make([]byte, sessionKeyLength)
		_, err = rand.Read(b)

		if err == nil {
			reply.SessionId = base64.URLEncoding.EncodeToString(b)
			_, err = db.Exec("INSERT INTO Session VALUES (?, ?, ?, ?, ?, ?)", reply.SessionId, student.StudentId, student.Name, student.AcademicYear, student.Department, student.Section)

			if err == nil {
				json.NewEncoder(w).Encode(reply)
			} else {
				io.WriteString(w, emptyJson)
			}
		} else {
			io.WriteString(w, emptyJson)
		}
	} else {
		io.WriteString(w, emptyJson)
	}
}

func getQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Get Questions Handler
}

func updateQuestionHandler(w http.ResponseWriter, r * http.Request) {
	//TODO: Update Question Handler
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Logout Handler
}

func main() {
	var dbErr error
	db, dbErr = sql.Open("mysql", user + ":" + password + "@/" + database)
	defer db.Close()

	if dbErr == nil {
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/", fs)
		http.HandleFunc("/register", registerHandler)
		http.HandleFunc("/login", loginHandler)
		http.HandleFunc("/logout", logoutHandler)
		http.ListenAndServe(":80", nil)
	} else {
		panic(dbErr)
	}
}
