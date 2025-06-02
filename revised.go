package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type Server struct {
	Name string
	IP   string
}

var (
	servers = []Server{
		{Name: "Server1", IP: "192.168.1.1"},
		{Name: "Server2", IP: "192.168.1.2"},
	}
	mu sync.Mutex
)

// Template for homepage with Add Server button, server dropdown, and Login button
const homepageTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Homepage</title>
</head>
<body>
	<h1>Welcome to the Homepage</h1>

	<!-- Add Server Button -->
	<form action="/add-server" method="post" style="margin-bottom:20px;">
		<input type="text" name="servername" placeholder="Server Name" required>
		<input type="text" name="serverip" placeholder="Server IP" required>
		<input type="submit" value="Add Server">
	</form>

	<!-- Server selection and Login -->
	<form action="/login-server" method="post">
		<label for="servers">Select Server:</label>
		<select name="serverip" id="servers" required>
			{{range .Servers}}
			<option value="{{.IP}}">{{.Name}} ({{.IP}})</option>
			{{end}}
		</select>
		<input type="submit" value="Login">
	</form>
</body>
</html>
`

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("homepage").Parse(homepageTemplate))
	mu.Lock()
	defer mu.Unlock()
	tmpl.Execute(w, struct {
		Servers []Server
	}{
		Servers: servers,
	})
}

// Handle adding new server from form submission
func addServerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("servername")
	ip := r.FormValue("serverip")

	if name == "" || ip == "" {
		http.Error(w, "Server name and IP required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	servers = append(servers, Server{Name: name, IP: ip})
	mu.Unlock()

	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

// Handle login to selected server (just a stub, no real login here)
func loginServerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ip := r.FormValue("serverip")
	if ip == "" {
		http.Error(w, "No server selected", http.StatusBadRequest)
		return
	}

	// Just respond with a simple message for now
	fmt.Fprintf(w, "<h2>Logged into server with IP: %s</h2><a href=\"/homepage\">Back to Homepage</a>", ip)
}

func main() {
	http.HandleFunc("/homepage", homepageHandler)
	http.HandleFunc("/add-server", addServerHandler)
	http.HandleFunc("/login-server", loginServerHandler)

	fmt.Println("Server started at http://localhost:8080/homepage")
	http.ListenAndServe(":8080", nil)
}
