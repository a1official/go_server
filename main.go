package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
)

type Student struct {
	Username string
	Password string
}

var accounts = make(map[string]string)

func readCSVFromFile(file io.Reader) ([]Student, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var students []Student
	for i, record := range records {
		if i == 0 || len(record) < 2 {
			continue
		}
		students = append(students, Student{
			Username: record[0],
			Password: record[1],
		})
	}
	return students, nil
}

func createAccountsFromCSV(students []Student) {
	for _, student := range students {
		accounts[student.Username] = student.Password
	}
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

func loginCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if storedPassword, exists := accounts[username]; exists && storedPassword == password {
		fmt.Fprintf(w, "✅ Login successful! Welcome, %s", username)
	} else {
		fmt.Fprintf(w, "❌ Invalid username or password.")
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	students, err := readCSVFromFile(file)
	if err != nil {
		http.Error(w, "Error reading CSV file", http.StatusInternalServerError)
		return
	}

	createAccountsFromCSV(students)

	fmt.Fprintf(w, "✅ CSV file uploaded and accounts created successfully!")
}

func main() {
	http.HandleFunc("/login", loginPageHandler)
	http.HandleFunc("/login-check", loginCheckHandler)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
