package main

import (
	"crypto/rand"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
    var parseErr error
    t.once.Do(func() {
        templatePath := filepath.Join("templates", t.filename)
        t.templ, parseErr = template.ParseFiles(templatePath)
        if parseErr != nil {
            log.Printf("ERROR: Failed to parse template %s: %v", templatePath, parseErr)
        } else {
            log.Printf("Successfully loaded template: %s", templatePath)
        }
    })
    
    if parseErr != nil {
        http.Error(w, "Template parsing error: "+parseErr.Error(), http.StatusInternalServerError)
        return
    }
    
    if t.templ == nil {
        http.Error(w, "Template not initialized: "+t.filename, http.StatusInternalServerError)
        return
    }
    
    data := map[string]interface{}{
        "Host": r.Host,
    }
    
    if authCookie, err := r.Cookie("auth"); err == nil {
        data["UserData"] = objx.MustFromBase64(authCookie.Value)
    }
    
    if err := t.templ.Execute(w, data); err != nil {
        log.Printf("Error executing template %s: %v", t.filename, err)
        http.Error(w, "Template execution error", http.StatusInternalServerError)
    }
}

func generateSecurityKey() string {
    bytes := make([]byte, 32) // 32 random bytes
    rand.Read(bytes)
    return string(bytes)
}

func main() {
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		// Only attempt to load .env in local development
		// This silences the error in Railway
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
	
	// ✅ ADDED: Root route handler - redirects to login page
	http.Handle("/", &templateHandler{filename: "login.html"})
	
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.Handle("/room", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/auth/", loginHandler)
	// get the room going
	go r.run()
	
	// ✅ FIX: Use Railway's PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local development
	}
	
	// start the web server
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}