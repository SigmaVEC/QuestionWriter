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
	"strconv"
	"time"
)

var (
	sessionKeyLength = 50 //In bytes
	sessionExpiry = 60 //In minutes
	emptyJson = "{}"
	db *sql.DB
	user string = "test"
	password string = "test"
	database string = "QuestionWriter"
)

type RegisterRequest struct {
	RegisterNumber string
	Name           string
	AcademicYear   string
	Department     string
	Year           string
	Semester       string
}

type RegisterResponse struct {
	SessionId string
}

type QuestionUpdateRequest struct {
	QuestionId int
	Answer     string
}

type SubQuestionModel struct {
	QuestionId int
	Question   string
	Choice     []string
}

type QuestionModel struct {
	File        string
	Description string
	SubQuestion []SubQuestionModel
}

type QuestionListResponse struct {
	Expiry   int
	Question []QuestionModel
}

type SubQuestionResultModel struct {
	QuestionId    int
	Question      string
	Answer        string
	CorrectAnswer string
	Reason        string
}

type ResultAnalysisResponse struct {
	SubQuestion []SubQuestionResultModel
}

func isValidSession(sessionId string) bool {
	var timeout time.Time
	dbSession := db.QueryRow("SELECT SessionId, Timeout FROM Session WHERE SessionId=?", sessionId)
	err := dbSession.Scan(&sessionId, &timeout)

	if err == nil && len(sessionId) != 0 && time.Now().Before(timeout) {
		return true
	} else {
		return false
	}
}

func displayWebPage(w http.ResponseWriter, file string) {
	t, _ := template.ParseFiles(file)
	t.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "Register.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "dashboard.html")
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	displayWebPage(w, "Result.html")
}

func getQuestion(question string, file string, data QuestionListResponse) (int, error) {
	for i := 0; i < len(data.Question); i++ {
		if data.Question[i].Description == question && data.Question[i].File == file {
			return i, nil
		}
	}

	return -1, errors.New("not found")
}

func generateSubQuestion(questionId int, subQuestion string) (SubQuestionModel, error) {
	var subQuestionObject SubQuestionModel
	dbChoices, err := db.Query("SELECT Choice FROM Choices WHERE QuestionId=?", questionId)
	defer dbChoices.Close()

	if err == nil {
		subQuestionObject.QuestionId = questionId
		subQuestionObject.Question = subQuestion
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
			var duration time.Duration = sessionExpiry * time.Minute
			timeout := time.Now().Add(duration)
			_, err = db.Exec("INSERT INTO Session VALUES (?, ?, ?, ?, ?, ?, ?, ?)", reply.SessionId, student.RegisterNumber, student.Name, student.AcademicYear, student.Department, student.Year, student.Semester, timeout)

			if err == nil {
				w.Header().Set("Content-Type", "application/json")
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

	if isValidSession(sessionId) {
		dbQuestions, err := db.Query("SELECT QuestionId, Question, File, SubQuestion FROM Questions");
		defer dbQuestions.Close()

		if err == nil {
			var reply QuestionListResponse
			reply.Expiry = sessionExpiry

			for dbQuestions.Next() {
				var questionId int
				var data [3]string
				dbQuestions.Scan(&questionId, &data[0], &data[1], &data[2])
				i, err := getQuestion(data[0], data[1], reply)

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

			w.Header().Set("Content-Type", "application/json")
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

	if isValidSession(sessionId) {
		data := []byte(updateJson)
		var update QuestionUpdateRequest
		err := json.Unmarshal(data, &update)

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
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(struct {
							Message string
						}{ Message: "Success" })
					} else {
						io.WriteString(w, emptyJson)
					}
				} else {
					_, err = db.Exec("INSERT INTO StudentAnswers VALUES(?,?,?)", sessionId, update.QuestionId, update.Answer)

					if err == nil {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(struct {
							Message string
						}{ Message: "Success" })
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

func getAnswerHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := r.FormValue("sessionId")
	questionId := r.FormValue("questionId")

	if isValidSession(sessionId) {
		question, err := strconv.Atoi(questionId)

		if err == nil {
			var answer string
			row := db.QueryRow("SELECT Answer FROM StudentAnswers WHERE SessionId=? AND QuestionId=?", sessionId, question)
			err = row.Scan(&answer)

			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(QuestionUpdateRequest {
					QuestionId: question,
					Answer: answer })
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

func reportHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := r.FormValue("sessionId")

	if isValidSession(sessionId) {
		var questionLength int
		row := db.QueryRow("SELECT MAX(QuestionId) FROM Questions")
		err := row.Scan(&questionLength)

		if err == nil {
			var subQuestionArray []SubQuestionResultModel
			for i := 1; i <= questionLength; i++ {
				var subQuestion, answer, reason string
				row = db.QueryRow("SELECT SubQuestion, Answer, AnswerReason FROM Questions WHERE QuestionId=?", i)
				err = row.Scan(&subQuestion, &answer, &reason)

				if err == nil {
					var studentAnswer string
					row = db.QueryRow("SELECT Answer FROM StudentAnswers WHERE SessionId=? AND QuestionId=?", sessionId, i)
					err = row.Scan(&studentAnswer)

					if err == nil {
						subQuestionArray = append(subQuestionArray, SubQuestionResultModel {
							QuestionId: i,
							Question: subQuestion,
							Answer: studentAnswer,
							CorrectAnswer: answer,
							Reason: reason })
					} else {
						io.WriteString(w, emptyJson)
					}
				} else {
					io.WriteString(w, emptyJson)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ResultAnalysisResponse {
				SubQuestion: subQuestionArray })
		} else {
			io.WriteString(w, emptyJson)
		}
	} else {
		io.WriteString(w, emptyJson)
	}
}

func studentDetailsHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := r.FormValue("sessionId")

	if isValidSession(sessionId) {
		var data [6]string
		row := db.QueryRow("SELECT StudentId, Name, AcademicYear, Department, Year, Semester FROM Session WHERE SessionId=?", sessionId)
		err := row.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5])

		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(RegisterRequest {
				RegisterNumber: data[0],
				Name: data[1],
				AcademicYear: data[2],
				Department: data[3],
				Year: data[4],
				Semester: data[5] })
		} else {
			io.WriteString(w, emptyJson)
		}
	} else {
		io.WriteString(w, emptyJson)
	}
}

func main() {
	var dbErr error
	db, dbErr = sql.Open("mysql", user + ":" + password + "@/" + database)
	defer db.Close()

	if dbErr == nil {
		http.Handle("/", http.FileServer(http.Dir("./static")))
		http.HandleFunc("/register", registerHandler)
		http.HandleFunc("/dashboard", dashboardHandler)
		http.HandleFunc("/results", resultsHandler)
		http.HandleFunc("/login", loginHandler)
		http.HandleFunc("/questions", getQuestionsHandler)
		http.HandleFunc("/update", updateQuestionHandler)
		http.HandleFunc("/getanswer", getAnswerHandler)
		http.HandleFunc("/studentDetails", studentDetailsHandler)
		http.HandleFunc("/report", reportHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}