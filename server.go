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
	"errors"
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

	return -1, errors.New("not found")
}

func generateSubQuestion(questionId int, subQuestion string) (SubQuestionModel, error) {
	var subQuestionObject SubQuestionModel
	dbChoices, err := db.Query("SELECT Choice FROM Choices WHERE QuestionId=?", questionId)
	defer dbChoices.Close()

	subQuestionObject.QuestionId = questionId
	subQuestionObject.Question = subQuestion

	if err == nil {
		hasChoices := false

		for dbChoices.Next() {
			hasChoices = true
			var choice string
			dbChoices.Scan(&choice)
			subQuestionObject.Choice = append(subQuestionObject.Choice, choice)
		}

		if hasChoices == false {
			subQuestionObject.Choice = nil
		}

		return subQuestionObject, nil
	} else {
		return subQuestionObject, err
	}
}

func displayWebPage(w http.ResponseWriter, file string) {
	t, _ := template.ParseFiles(file)
	t.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "Register.html")
}

func dashBoardHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "dashboard.html")
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

		if err == nil {
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
	sessionId := r.FormValue("sessionId")
	updateJson := r.FormValue("updateJson")
	dbSession := db.QueryRow("SELECT SessionId FROM Session WHERE SessionId=?", sessionId)
	err := dbSession.Scan(&sessionId)

	if err == nil && len(sessionId) != 0 {
		data := []byte(updateJson)
		var update QuestionUpdateRequest
		err = json.Unmarshal(data, &update)

		if err == nil {
			var temp int
			dbQuestion := db.QueryRow("SELECT QuestionId FROM Questions WHERE QuestionId=?", update.QuestionId)
			err = dbQuestion.Scan(&temp)

			if err == nil {
				dbStudentAnswer := db.QueryRow("SELECT QuestionId FROM StudentAnswers WHERE SessionId=? AND QuestionId=?", sessionId, update.QuestionId)
				err = dbStudentAnswer.Scan(&temp)

				if err == nil {
					_, err = db.Exec("UPDATE StudentAnswers SET Answer=? WHERE SessionId=? AND QuestionId=?", update.Answer, sessionId, update.QuestionId)

					if err == nil {
						json.NewEncoder(w).Encode(struct {
							Message string
						}{Message: "Success"})
					} else {
						io.WriteString(w, emptyJson)
					}
				} else {
					_, err = db.Exec("INSERT INTO StudentAnswers VALUES(?,?,?)", sessionId, update.QuestionId, update.Answer)

					if err == nil {
						json.NewEncoder(w).Encode(struct {
							Message string
						}{Message: "Success"})
					} else {
						io.WriteString(w, emptyJson)
					}
				}
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
		http.HandleFunc("/dashboard", dashBoardHandler)
		http.HandleFunc("/login", loginHandler)
		http.HandleFunc("/questions", getQuestionsHandler)
		http.HandleFunc("/update", updateQuestionHandler)
		http.HandleFunc("/logout", logoutHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}
