package main

import (
	"crypto/rand"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
    "github.com/joho/godotenv"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		templatePath := filepath.Join("templates", t.filename)
		t.templ = template.Must(template.ParseFiles(templatePath))
	})
	
	data := map[string]interface{}{
		"Host": r.Host,
	}
	
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	
	// ✅ FIX: Pass data instead of r to the template
	t.templ.Execute(w, data)
}

func generateSecurityKey() string {
    bytes := make([]byte, 32) // 32 random bytes
    rand.Read(bytes)
    return string(bytes)
}

func main() {
	
    if err := godotenv.Load(); err != nil {
        log.Printf("❌ Error loading .env file: %v", err)
	}
    
    // Check environment variables
    clientID := os.Getenv("GOOGLE_CLIENT_ID")
    clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
    callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")


    gomniauth.SetSecurityKey(generateSecurityKey())
    gomniauth.WithProviders(
        google.New(clientID, clientSecret, callbackURL),
    )
    



	r := newRoom()
	
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.Handle("/room", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/auth/", loginHandler)
	// get the room going
	go r.run()
	
	// start the web server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}