package main

import (
    "fmt"
    "html/template"
    "net/http"
    "os"
    "io"
    "strconv"
    
    "github.com/gorilla/mux"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/joho/godotenv"
    "github.com/google/uuid"
    "github.com/gorilla/sessions"
    "github.com/go-gomail/gomail"
)

var (
    templates *template.Template
    db  *gorm.DB
    err error
)

type User struct {
    gorm.Model
    Name     string `gorm:"not null"`
    Username string `gorm:"unique;not null"`
    Password string `gorm:"not null"`
    Email    string `gorm:"unique;not null"`
}

type Post struct {
    gorm.Model
    Title     string `gorm:"not null"`
    Body      string `gorm:"not null"`
    ImagePath string
}

func init() {
    load_dot_env()
    templates = template.Must(template.ParseGlob("templates/*.html"))
    
    // connect to db
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        os.Getenv("DB_USERNAME"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )
    
    db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        fmt.Println(err)
        panic("Failed to connect to the database")
    }

    // creating tables
    db.AutoMigrate(&User{}, &Post{})
}

func main() {
    r := mux.NewRouter()
    
    // Routes
    r.HandleFunc("/", homeHandler).Methods("GET")
    r.HandleFunc("/dashboard", dashboardHandler).Methods("GET")
    r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
    r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
    r.HandleFunc("/add-post", addPostPageHandler).Methods("GET")
    r.HandleFunc("/add-post", addPostHandler).Methods("POST")
    r.HandleFunc("/about", aboutHandler).Methods("GET")
    r.HandleFunc("/logout", logoutHandler).Methods("GET")
    r.HandleFunc("/privacy", privacyHandler).Methods("GET")
    r.HandleFunc("/profile", profileHandler).Methods("GET")


    // Serve static files from the "static" directory
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/", homeHandler)

	http.Handle("/", r)

	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", nil)

    http.Handle("/", r)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    err := templates.ExecuteTemplate(w, tmpl+".html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    var posts []Post
    db.Find(&posts)

    tmpl, err := template.ParseFiles("templates/home.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, struct{ Posts []Post }{Posts: posts})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    var posts []Post
    db.Find(&posts)

    tmpl, err := template.ParseFiles("templates/dashboard.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, struct{ Posts []Post }{Posts: posts})
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "about", nil)
}

func privacyHandler(w http.ResponseWriter, r *http.Request){
    renderTemplate(w, "privacy", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")

		newUser := User{Name: name, Username: username, Password: password, Email: email}

		db.Create(&newUser)
        sendWelcomeEmail(email, name)

        session, err := store.Get(r, "user-session")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        session.Values["username"] = newUser.Username
        session.Values["name"]     = newUser.Name
        session.Values["email"]    = newUser.Email
        err = session.Save(r, w)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "register", nil)
}

var store = sessions.NewCookieStore([]byte("your-secret-key"))
func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        var user User
        result := db.Where("username = ? AND password = ?", username, password).First(&user)
        if result.Error == nil {
            session, err := store.Get(r, "user-session")
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            session.Values["username"] = user.Username
            session.Values["name"] = user.Name
            session.Values["email"] = user.Email
            err = session.Save(r, w)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
            return
        }
    }
    renderTemplate(w, "login", map[string]interface{}{"Error": "Invalid username or password"})
}

func addPostHandler(w http.ResponseWriter, r *http.Request) {
    title := r.FormValue("title")
    body := r.FormValue("body")

    file, handler, err := r.FormFile("image")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()

    imagePath := "static/images/" + uuid.New().String() + handler.Filename
    f, err := os.Create(imagePath)
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer f.Close()
    io.Copy(f, file)

    // افزودن پست به دیتابیس با اطلاعات آپلود شده
    post := Post{Title: title, Body: body, ImagePath: imagePath}
    db.Create(&post)

    // بازگشت به صفحه home
    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func addPostPageHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("templates/add-post.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "user-session")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    username, ok := session.Values["username"].(string)
    if !ok {
        http.Error(w, "User not logged in", http.StatusUnauthorized)
        return
    }

    var user User
    result := db.Where("username = ?", username).First(&user)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    renderTemplate(w, "profile", user)
}

func load_dot_env() {
    err := godotenv.Load(".env")
    if err != nil {
        fmt.Println(err)
    }
}

func sendWelcomeEmail(email, name string) {
    mailPort, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
    if err != nil {
        fmt.Println("Error converting MAIL_PORT to int:", err)
        return
    }
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("MAIL_FROM")) 
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Welcome to Liara Blog")
	body := fmt.Sprintf("Dear %s,\n\nWelcome to Liara Blog! We're excited to have you on board.", name)
	m.SetBody("text/plain", body)

    d := gomail.NewDialer(os.Getenv("MAIL_HOST"), mailPort, os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Error sending welcome email:", err)
	}
}

