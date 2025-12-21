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
    var err error
    
    t.once.Do(func() {
        templatePath := filepath.Join("templates", t.filename)
        
        t.templ, err = template.ParseFiles(templatePath)
        if err != nil {
            log.Printf("Failed to parse template %s: %v", templatePath, err)
        }
    })
    
    if err != nil || t.templ == nil {
        http.Error(w, "Template failed to load", http.StatusInternalServerError)
        return
    }
    
    data := map[string]interface{}{
        "Host": r.Host,
    }
    
    if authCookie, err := r.Cookie("auth"); err == nil {
        data["UserData"] = objx.MustFromBase64(authCookie.Value)
    }
    
    if err := t.templ.Execute(w, data); err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Template error", http.StatusInternalServerError)
    }
}

func generateSecurityKey() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return string(bytes)
}

func main() {
    // Initialize database
    initDB()
    
    // Check environment variables
    clientID := os.Getenv("GOOGLE_CLIENT_ID")
    clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
    callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")

    if clientID != "" && clientSecret != "" {
        gomniauth.SetSecurityKey(generateSecurityKey())
        gomniauth.WithProviders(
            google.New(clientID, clientSecret, callbackURL),
        )
    }
    
    r := newRoom()
    
    http.Handle("/", &templateHandler{filename: "login.html"})
    http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
    http.Handle("/login", &templateHandler{filename: "login.html"})
    http.Handle("/room", r)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/auth/", loginHandler)
    
    go r.run()
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Starting server on :%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}