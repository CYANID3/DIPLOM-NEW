package handlers

import (
	"html/template"
	"net/http"
	"net/url"
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

	data := map[string]interface{}{
		"Products": products,
		"Username": display,
		"Role":     role,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
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
		http.Redirect(w, r, "/?error="+url.QueryEscape("Неверная цена"), http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Неверное количество"), http.StatusSeeOther)
		return
	}

	err = models.CreateProduct(name, barcode, price, qty, username)
	if err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Ошибка добавления товара"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Товар добавлен"), http.StatusSeeOther)
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
		http.Redirect(w, r, "/?error="+url.QueryEscape("Неверный ID"), http.StatusSeeOther)
		return
	}

	err = models.DeleteProduct(id, username)
	if err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Ошибка удаления товара"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Товар удалён"), http.StatusSeeOther)
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
		http.Redirect(w, r, "/?error="+url.QueryEscape("Неверный ID"), http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Неверное количество"), http.StatusSeeOther)
		return
	}

	err = models.SellProduct(id, qty, username)
	if err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Продажа выполнена"), http.StatusSeeOther)
}
