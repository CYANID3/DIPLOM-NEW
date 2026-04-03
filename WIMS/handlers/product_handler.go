package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"wims/models"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func IndexPage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
		return
	}

	errorMsg := r.URL.Query().Get("error")

	data := map[string]interface{}{
		"Products": products,
		"Username": display,
		"Role":     role,
		"Error":    errorMsg,
	}

	tmpl.ExecuteTemplate(w, "index.html", data)
}

func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	barcode := r.FormValue("barcode")

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "Неверная цена", http.StatusBadRequest)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Неверное количество", http.StatusBadRequest)
		return
	}

	err = models.CreateProduct(name, barcode, price, qty, username)
	if err != nil {
		http.Error(w, "Ошибка создания товара", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	err = models.DeleteProduct(id, username)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func SellProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/?error=Неверный+ID", http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Redirect(w, r, "/?error=Неверное+количество", http.StatusSeeOther)
		return
	}

	err = models.SellProduct(id, qty, username)
	if err != nil {
		http.Redirect(w, r, "/?error="+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
