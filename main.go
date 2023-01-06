package main

import (
	"context"
	"crud-database/connection"
	"fmt"

	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnection()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public/"))))

	route.HandleFunc("/", homePage).Methods("GET")
	route.HandleFunc("/contact", contactPage).Methods("GET")
	route.HandleFunc("/project", projectPage).Methods("GET")
	route.HandleFunc("/project", addProject).Methods("POST")
	route.HandleFunc("/project/{id}", detailProject).Methods("GET")
	route.HandleFunc("/editProject/{id}", editProject).Methods("GET")
	route.HandleFunc("/updateProject/{id}", updateProject).Methods("POST")
	route.HandleFunc("/deleteProject/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	fmt.Println("Server running on port:5000")
	http.ListenAndServe("localhost:5000", route)
}

var Data = map[string]interface{}{
	"IsLogin": true,
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/index.html") //parse menguraikan tmpl
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	rows, errQuery := connection.Conn.Query(context.Background(), "SELECT * FROM tb_project")
	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	var result []dataReceive

	for rows.Next() {
		var each = dataReceive{}

		err := rows.Scan(&each.ID, &each.Projectname, &each.Startdate, &each.Enddate, &each.Description, &each.Technologies)
		if err != nil {
			fmt.Println("Message : " + err.Error())
			return
		}

		each.Duration = countduration(each.Startdate, each.Enddate)

		result = append(result, each)
	}

	dataMain := map[string]interface{}{ //membuat data untuk projects
		"Projects": result,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataMain)
}

func projectPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/myProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

type dataReceive struct {
	ID           int
	Projectname  string
	Description  string
	Technologies []string
	Startdate    time.Time
	Enddate      time.Time
	Duration     string
}

func countduration(start time.Time, end time.Time) string {
	distance := end.Sub(start)

	//Ubah milisecond menjadi bulan, minggu dan hari
	monthDistance := int(distance.Hours() / 24 / 30)
	weekDistance := int(distance.Hours() / 24 / 7)
	daysDistance := int(distance.Hours() / 24)

	// variable buat menampung durasi yang sudah diolah
	var duration string
	// pengkondisian yang akan mengirimkan durasi yang sudah diolah
	if monthDistance >= 1 {
		duration = strconv.Itoa(monthDistance) + " months"
	} else if monthDistance < 1 && weekDistance >= 1 {
		duration = strconv.Itoa(weekDistance) + " weeks"
	} else if monthDistance < 1 && daysDistance >= 0 {
		duration = strconv.Itoa(daysDistance) + " days"
	} else {
		duration = "0 days"
	}
	// Duration End

	return duration
}

func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5242880) //Pemanggilan method tersebut membuat file yang terupload disimpan sementara pada memory dengan alokasi adalah sesuai dengan maxMemory. Jika ternyata kapasitas yang sudah dialokasikan tersebut tidak cukup, maka file akan disimpan dalam temporary file.
	//5242880 byte to biner = 5mb

	if err != nil {
		log.Fatal(err)
	}

	projectname := r.PostForm.Get("project-name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]

	const timeFormat = "2006-01-02"
	startDate, _ := time.Parse(timeFormat, r.PostForm.Get("start-date"))
	endDate, _ := time.Parse(timeFormat, r.PostForm.Get("end-date"))

	_, insertRow := connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(project_name, start_date, end_date, description, technologies) VALUES ($1, $2, $3, $4, $5)", projectname, startDate, endDate, description, technologies)
	if insertRow != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/project-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	var resultData = dataReceive{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", ID).Scan(&resultData.ID, &resultData.Projectname, &resultData.Startdate, &resultData.Enddate, &resultData.Description, &resultData.Technologies)

	resultData.Duration = countduration(resultData.Startdate, resultData.Enddate)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	detailProject := map[string]interface{}{
		"Projects": resultData,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, detailProject)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, deleteRows := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if deleteRows != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + deleteRows.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/editProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	var editData = dataReceive{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", ID).Scan(&editData.ID, &editData.Projectname, &editData.Startdate, &editData.Enddate, &editData.Description, &editData.Technologies)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	dataEdit := map[string]interface{}{
		"Projects": editData,
	}

	tmpl.Execute(w, dataEdit)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5242880)

	if err != nil {
		log.Fatal(err)
	}

	projectname := r.PostForm.Get("project-name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]

	const timeFormat = "2006-01-02"
	startDate, _ := time.Parse(timeFormat, r.PostForm.Get("start-date"))
	endDate, _ := time.Parse(timeFormat, r.PostForm.Get("end-date"))
	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, updateRow := connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5 WHERE id = $6", projectname, startDate, endDate, description, technologies, ID)
	if updateRow != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; chartset=utf-8")

	tmpl, err := template.ParseFiles("view/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; chartset=utf-8")

	tmpl, err := template.ParseFiles("view/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	sessions, _ := store.Get(r, "SESSION_ID")

	if sessions.Values["IsLogin"] != true {
		Data["Islogin"] = false
	} else {
		Data["IsLogin"] = sessions.Values["IsLogin"].(bool)
		Data["username"] = sessions.Values["name"].(string)

	}

	dataMain := map[string]interface{}{
		"Data": Data,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataMain)
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Options.MaxAge = 10800

	session.AddFlash("Login success", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
