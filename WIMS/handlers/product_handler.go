package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"wims/models"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func IndexPage(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки", 500)
		return
	}

	_, role, display := GetSession(r)

	data := map[string]interface{}{
		"Products": products,
		"Username": display,
		"Role":     role,
	}

	tmpl.ExecuteTemplate(w, "index.html", data)
}

func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", 303)
		return
	}

	name := r.FormValue("name")
	barcode := r.FormValue("barcode")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	qty, _ := strconv.Atoi(r.FormValue("quantity"))

	models.CreateProduct(name, barcode, price, qty, username)

	http.Redirect(w, r, "/", 303)
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", 303)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	models.DeleteProduct(id, username)

	http.Redirect(w, r, "/", 303)
}

func SellProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", 303)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))

	models.SellProduct(id, qty, username)

	http.Redirect(w, r, "/", 303)
}
