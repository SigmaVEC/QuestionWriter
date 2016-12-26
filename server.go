package main

import (
	"net/http"
	"html/template"
)

type QuestionResponse struct {
	Result []struct {
		Question int
		Answer string
	}
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

func main() {
	http.HandleFunc("/register", registerHandler)
	http.ListenAndServe(":9000", nil)
}