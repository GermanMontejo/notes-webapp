package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

type Book struct {
	Title       string
	Description string
	Author      string
}

type EditBook struct {
	Book
	Id string
}

var templates map[string]*template.Template
var library = make(map[string]Book)
var id int

func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Error:", err)
	}
	templates["add"] = template.Must(template.ParseFiles(filepath.Join(cwd, "/notes-webapp/templates/add.html")))
	templates["index"] = template.Must(template.ParseFiles(filepath.Join(cwd, "/notes-webapp/templates/index.html")))
	templates["edit"] = template.Must(template.ParseFiles(filepath.Join(cwd, "/notes-webapp/templates/edit.html")))
}

func renderTemplate(resp http.ResponseWriter, name string, dataObject interface{}) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(resp, "Error: Template not found!", http.StatusBadRequest)
	}

	err := tmpl.ExecuteTemplate(resp, name, dataObject)
	if err != nil {
		log.Fatal("Error in executing template:", err)
		http.Error(resp, "Error in executing template", http.StatusInternalServerError)
	}
}

func saveBooks(resp http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	title := req.PostFormValue("title")
	description := req.PostFormValue("description")
	author := req.PostFormValue("author")
	i := strconv.Itoa(id)
	book := Book{title, description, author}
	library[i] = book
	id++
	log.Println("library:", library)
	http.Redirect(resp, req, "/", 302)
}

func addBook(resp http.ResponseWriter, req *http.Request) {
	renderTemplate(resp, "add", nil)
}

func getBooks(resp http.ResponseWriter, req *http.Request) {
	renderTemplate(resp, "index", library)
}

func editBook(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bookId := vars["id"]
	book, ok := library[bookId]
	if !ok {
		http.NotFound(resp, req)
	}

	editBook := EditBook{book, bookId}
	renderTemplate(resp, "edit", editBook)
}

func updateBook(resp http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	title := req.PostFormValue("title")
	description := req.PostFormValue("description")
	author := req.PostFormValue("author")
	vars := mux.Vars(req)
	bookId := vars["id"]
	bookToUpd, ok := library[bookId]
	if !ok {
		http.NotFound(resp, req)
		return
	}
	// overwrite existing book fields with update fields.
	bookToUpd.Author = author
	bookToUpd.Description = description
	bookToUpd.Title = title
	delete(library, bookId)
	library[bookId] = bookToUpd
	http.Redirect(resp, req, "/", 302)
}

func deleteBook(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bookId := vars["id"]
	// check if book exists in the library
	_, ok := library[bookId]
	if !ok {
		http.NotFound(resp, req)
	}
	// if found, delete from the library
	delete(library, bookId)
	http.Redirect(resp, req, "/", 302)
}

func main() {
	m := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("public"))
	m.Handle("/public/", fs)
	m.HandleFunc("/", getBooks)
	m.HandleFunc("/books/add", addBook)
	m.HandleFunc("/books/save", saveBooks)
	m.HandleFunc("/books/edit/{id}", editBook)
	m.HandleFunc("/books/update/{id}", updateBook)
	m.HandleFunc("/books/delete/{id}", deleteBook)
	server := &http.Server{
		Addr:    ":8080",
		Handler: m,
	}
	log.Println("Listening on port 8080.")
	server.ListenAndServe()
}
