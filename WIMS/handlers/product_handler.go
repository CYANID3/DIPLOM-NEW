package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"wims/models"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// IndexPage
func IndexPage(w http.ResponseWriter, r *http.Request) {
	products, _ := models.GetAllProducts()
	username, _ := GetSession(r)
	data := map[string]interface{}{
		"Products": products,
		"Username": username,
	}
	tmpl.ExecuteTemplate(w, "index.html", data)
}

// AddProductHandler
func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	qty, _ := strconv.Atoi(r.FormValue("quantity"))

	models.CreateProduct(name, price, qty, username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteProductHandler
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	models.DeleteProduct(id, username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// SellProductHandler
func SellProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))

	models.SellProduct(id, qty, username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
