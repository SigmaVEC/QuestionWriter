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
)

var (
	sessionKeyLength = 50 //In bytes
	sessionExpiry = 30 //In minutes
	emptyJson = "{}"
	db *sql.DB
	user string = "root"
	password string = "toor"
	database string = "QuestionWriter"
)

type RegisterRequest struct {
	RegisterNumber string
	Name string
	AcademicYear int
	Branch string
	Semester int
	CollegeName string
}

type RegisterResponse struct {
	SessionId string
}

type QuestionUpdateRequest struct {
	QuestionId int
	Answer string
}

type SubQuestionModel struct {
	QuestionId int
	Question string
	Choice []string
}

type QuestionModel struct {
	File string
	Description string
	SubQuestion []SubQuestionModel
}

type QuestionListResponse struct {
	Expiry int
	Question []QuestionModel
}

func getQuestion(question string, data QuestionListResponse) (int, error) {
	for i := 0; i < len(data.Question); i++ {
		if data.Question[i].Description == question {
			return i, nil
		}
	}

	return -1, error.Error("not found")
}

func generateSubQuestion(questionId int, subQuestion string) (SubQuestionModel, error) {
	var subQuestionObject SubQuestionModel
	dbChoices, err := db.Query("SELECT Choice FROM Choices WHERE QuestionId=?", questionId)
	defer dbChoices.Close()

	subQuestionObject.QuestionId = questionId
	subQuestionObject.Question = subQuestion

	if err == nil {
		if len(dbChoices) != 0 {
			for dbChoices.Next() {
				var choice string
				dbChoices.Scan(&choice)
				subQuestionObject.Choice = append(subQuestionObject.Choice, choice)
			}
		} else {
			subQuestionObject.Choice = nil
		}

		return subQuestionObject, nil
	} else {
		return nil, err
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
			_, err = db.Exec("INSERT INTO Session VALUES (?, ?, ?, ?, ?, ?, ?)", reply.SessionId, student.RegisterNumber, student.Name, student.AcademicYear, student.Branch, student.Semester, student.CollegeName)

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
	sessionId := r.FormValue("sessionId")
	dbSession := db.QueryRow("SELECT SessionId FROM Session WHERE SessionId=?", sessionId)
	err := dbSession.Scan(&sessionId)

	if err == nil && len(sessionId) != 0 {
		dbQuestions, err := db.Query("SELECT QuestionId, Question, File, SubQuestion FROM Questions");
		defer dbQuestions.Close()

		if err == nil && len(dbQuestions) != 0 {
			var reply QuestionListResponse
			reply.Expiry = sessionExpiry

			for dbQuestions.Next() {
				var questionId int
				var data [3]string
				dbQuestions.Scan(&questionId, &data[0], &data[1], &data[2])
				i, err := getQuestion(data[0], reply)

				if err == nil {
					subQuestion, err := generateSubQuestion(questionId, data[2])

					if err == nil {
						reply.Question[i].SubQuestion = append(reply.Question[i].SubQuestion, subQuestion)
					} else {
						io.WriteString(w, emptyJson)
					}
				} else {
					question := QuestionModel{ Description: data[0], File: data[1] }
					subQuestion, err := generateSubQuestion(questionId, data[2])

					if err == nil {
						question.SubQuestion = append(question.SubQuestion, subQuestion)
						reply.Question = append(reply.Question, question)
					} else {
						io.WriteString(w, emptyJson)
					}
				}
			}

			json.NewEncoder(w).Encode(reply)
		} else {
			io.WriteString(w, emptyJson)
		}
	} else {
		io.WriteString(w, emptyJson)
	}
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
		http.HandleFunc("/questions", getQuestionsHandler)
		http.HandleFunc("/logout", logoutHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}
