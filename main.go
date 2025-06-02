package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Student struct {
	Username string
	Password string
}

var accounts = make(map[string]string) // username -> password (plain for now)

func readCSV(filename string) ([]Student, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var students []Student
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(record) < 2 {
			continue // skip invalid lines
		}
		students = append(students, Student{
			Username: strings.TrimSpace(record[0]),
			Password: strings.TrimSpace(record[1]),
		})
	}
	return students, nil
}

func readExcel(filename string) ([]Student, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var students []Student
	// Assuming header: RollNo, Name, Password
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 3 {
			continue
		}
		students = append(students, Student{
			Username: strings.TrimSpace(row[0]), // RollNo as username
			Password: strings.TrimSpace(row[2]), // Password column
		})
	}
	return students, nil
}

func createAccountsFromCSV(filename string) error {
	students, err := readCSV(filename)
	if err != nil {
		return err
	}

	for _, student := range students {
		accounts[student.Username] = student.Password
		createUserFolder(student.Username)
	}

	return nil
}

func createAccountsFromExcel(filename string) error {
	students, err := readExcel(filename)
	if err != nil {
		return err
	}

	for _, student := range students {
		accounts[student.Username] = student.Password
		createUserFolder(student.Username)
	}

	return nil
}

func createUserFolder(username string) {
	userFolder := filepath.Join("uploads", username)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		os.MkdirAll(userFolder, 0755)
	}
}

func uploadCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write([]byte(`<html><body>
		<h2>Upload CSV file</h2>
		<form action="/upload-csv" method="post" enctype="multipart/form-data">
		<input type="file" name="csvfile" accept=".csv" required>
		<input type="submit" value="Upload CSV">
		</form>
		<br><a href="/user-login">Back to Home</a>
		</body></html>`))
		return
	}

	file, _, err := r.FormFile("csvfile")
	if err != nil {
		http.Error(w, "Failed to read CSV file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "upload-*.csv")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	io.Copy(tempFile, file)

	err = createAccountsFromCSV(tempFile.Name())
	if err != nil {
		http.Error(w, "Error creating accounts from CSV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Accounts created from CSV successfully!<br><a href=\"/user-login\">Back to Home</a>"))
}

func uploadExcelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write([]byte(`<html><body>
		<h2>Upload Excel file</h2>
		<form action="/upload-excel" method="post" enctype="multipart/form-data">
		<input type="file" name="excelfile" accept=".xlsx,.xls" required>
		<input type="submit" value="Upload Excel">
		</form>
		<br><a href="/user-login">Back to Home</a>
		</body></html>`))
		return
	}

	file, _, err := r.FormFile("excelfile")
	if err != nil {
		http.Error(w, "Failed to read Excel file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	io.Copy(tempFile, file)

	err = createAccountsFromExcel(tempFile.Name())
	if err != nil {
		http.Error(w, "Error creating accounts from Excel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Accounts created from Excel successfully!<br><a href=\"/user-login\">Back to Home</a>"))
}

func accountsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h2>All Accounts</h2><ul>")
	for user := range accounts {
		fmt.Fprintf(w, "<li>%s</li>", user)
	}
	fmt.Fprintf(w, "</ul>")
}

// /user-login page with 5 buttons including Dashboard & Admin Login
func userLoginPage(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Student Portal</title>
		<style>
			body { font-family: Arial, sans-serif; padding: 40px; }
			h2 { margin-bottom: 30px; }
			.button-container {
				display: flex;
				gap: 20px;
				flex-wrap: wrap;
			}
			.button {
				background-color: #4CAF50;
				border: none;
				color: white;
				padding: 15px 30px;
				text-align: center;
				text-decoration: none;
				display: inline-block;
				font-size: 16px;
				cursor: pointer;
				border-radius: 5px;
				transition: background-color 0.3s;
			}
			.button:hover {
				background-color: #45a049;
			}
		</style>
	</head>
	<body>
		<h2>Welcome to Student Portal</h2>
		<div class="button-container">
			<a href="/login-form" class="button">Login</a>
			<a href="/upload-csv" class="button">Upload CSV</a>
			<a href="/upload-excel" class="button">Upload Excel</a>
			<a href="/dashboard" class="button">Dashboard</a>
			<a href="/admin-login" class="button">Admin Login</a>
		</div>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// Login form page
func loginFormPage(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head><title>Login</title></head>
	<body>
		<h2>Student Login</h2>
		<form action="/login-auth" method="post">
			<label>Username (Roll No):</label><br>
			<input type="text" name="username" required><br><br>
			<label>Password:</label><br>
			<input type="password" name="password" required><br><br>
			<input type="submit" value="Login">
		</form>
		<br>
		<a href="/user-login">Back to Home</a>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/user-login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	expectedPassword, exists := accounts[username]
	if !exists || expectedPassword != password {
		http.Error(w, "❌ Invalid username or password", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
}

// Basic dashboard page
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "Missing user parameter", http.StatusBadRequest)
		return
	}

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Dashboard</title></head>
		<body>
			<h2>Welcome, %s!</h2>
			<p>You are logged in.</p>
			<a href="/user-login">Logout</a>
		</body>
		</html>
	`, user)

	w.Write([]byte(html))
}

// Basic admin login page (expand as needed)
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head><title>Admin Login</title></head>
	<body>
		<h2>Admin Login</h2>
		<form action="/admin-auth" method="post">
			<label>Admin Username:</label><br>
			<input type="text" name="adminuser" required><br><br>
			<label>Admin Password:</label><br>
			<input type="password" name="adminpass" required><br><br>
			<input type="submit" value="Login">
		</form>
		<br>
		<a href="/user-login">Back to Home</a>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// Basic admin auth handler (dummy)
func adminAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin-login", http.StatusSeeOther)
		return
	}

	adminUser := r.FormValue("adminuser")
	adminPass := r.FormValue("adminpass")

	// Replace this with real admin check
	if adminUser == "admin" && adminPass == "admin123" {
		w.Write([]byte("<h2>Admin logged in successfully!</h2><br><a href=\"/user-login\">Back to Home</a>"))
	} else {
		http.Error(w, "Invalid admin credentials", http.StatusUnauthorized)
	}
}

func main() {
	// Create uploads directory if not exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	http.HandleFunc("/user-login", userLoginPage)
	http.HandleFunc("/login-form", loginFormPage)
	http.HandleFunc("/login-auth", loginAuthHandler)
	http.HandleFunc("/upload-csv", uploadCSVHandler)
	http.HandleFunc("/upload-excel", uploadExcelHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/admin-login", adminLoginHandler)
	http.HandleFunc("/admin-auth", adminAuthHandler)
	http.HandleFunc("/accounts", accountsHandler)

	fmt.Println("Server running at http://localhost:8080/user-login")
	http.ListenAndServe(":8080", nil)
}
