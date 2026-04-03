package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"wims/models"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// Главная страница
func IndexPage(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки товаров", http.StatusInternalServerError)
		return
	}

	username, _ := GetSession(r)

	data := map[string]interface{}{
		"Products": products,
		"Username": username,
	}

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// Добавление товара
func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil || price <= 0 {
		http.Error(w, "Некорректная цена", http.StatusBadRequest)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		http.Error(w, "Некорректное количество", http.StatusBadRequest)
		return
	}

	err = models.CreateProduct(name, price, qty, username)
	if err != nil {
		http.Error(w, "Ошибка добавления товара", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Удаление товара
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
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
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = models.DeleteProduct(id, username)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Продажа товара
func SellProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
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
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		http.Error(w, "Некорректное количество", http.StatusBadRequest)
		return
	}

	err = models.SellProduct(id, qty, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
