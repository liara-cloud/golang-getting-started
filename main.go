package main

import (
    "fmt"
    "html/template"
    "net/http"
    "os"
    "io"
    
    "github.com/gorilla/mux"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/joho/godotenv"
    "github.com/google/uuid"
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
    r.HandleFunc("/login", loginHandler).Methods("GET")
    r.HandleFunc("/register", registerHandler).Methods("GET")
    r.HandleFunc("/register", registerHandler).Methods("POST")
    r.HandleFunc("/add-post", addPostPageHandler).Methods("GET")
    r.HandleFunc("/add-post", addPostHandler).Methods("POST")
    r.HandleFunc("/about", aboutHandler).Methods("GET")
    r.HandleFunc("/about", logoutHandler).Methods("POST")

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


func loginHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "login", nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "about", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        name := r.FormValue("name")
        username := r.FormValue("username")
        password := r.FormValue("password")
        email := r.FormValue("email")

        newUser := User{Name: name, Username: username, Password: password, Email: email}
        db.Create(&newUser)

        http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
        return
    }

    // درخواست از نوع GET
    renderTemplate(w, "register", nil)
}

func addPostHandler(w http.ResponseWriter, r *http.Request) {
    title := r.FormValue("title")
    body := r.FormValue("body")

    // دریافت فایل آپلود شده
    file, handler, err := r.FormFile("image")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()

    // ذخیره فایل آپلود شده در محل مورد نظر

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
    http.Redirect(w, r, "/", http.StatusSeeOther)
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

func load_dot_env() {
    err := godotenv.Load(".env")
    if err != nil {
        fmt.Println(err)
    }
}