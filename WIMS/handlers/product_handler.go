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
	_, role, display, ok := RequireAuth(w, r)
	if !ok {
		return
	}

	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Products": products,
		"Username": display,
		"Role":     role,
		"Settings": settings,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
	}

	tmpl.ExecuteTemplate(w, "index.html", data)
}

// SellProductHandler — продажа, доступна всем авторизованным
func SellProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireAuth(w, r)
	if !ok {
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

	if err := models.SellProduct(id, qty, username); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Продажа выполнена"), http.StatusSeeOther)
}

// AddProductHandler — добавление товара, только manager + admin
func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	barcode := r.FormValue("barcode")
	category := r.FormValue("category")

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

	minStock, _ := strconv.Atoi(r.FormValue("min_stock"))

	if err := models.CreateProduct(name, barcode, category, price, qty, minStock, username); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Ошибка добавления товара"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Товар добавлен"), http.StatusSeeOther)
}

// DeleteProductHandler — удаление товара, только manager + admin
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
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

	if err := models.DeleteProduct(id, username); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Ошибка удаления товара"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+url.QueryEscape("Товар удалён"), http.StatusSeeOther)
}
