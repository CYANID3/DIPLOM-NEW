package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"wims/models"
)

var regradeTmpl = template.Must(template.ParseFiles("templates/regrade.html", "templates/navbar.html"))

func RegradePage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки товаров", http.StatusInternalServerError)
		return
	}

	regradings, err := models.GetRegradings()
	if err != nil {
		http.Error(w, "Ошибка загрузки пересортов", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":   display,
		"Role":       role,
		"Products":   products,
		"Regradings": regradings,
		"Settings":   settings,
		"Error":      r.URL.Query().Get("error"),
		"Success":    r.URL.Query().Get("success"),
	}

	if err := regradeTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func CreateRegradeHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/regrade", http.StatusSeeOther)
		return
	}

	fromID, err := strconv.Atoi(r.FormValue("from_id"))
	if err != nil {
		http.Redirect(w, r, "/regrade?error="+url.QueryEscape("Выберите товар-источник"), http.StatusSeeOther)
		return
	}

	fromQty, err := strconv.Atoi(r.FormValue("from_qty"))
	if err != nil || fromQty <= 0 {
		http.Redirect(w, r, "/regrade?error="+url.QueryEscape("Неверное количество для списания"), http.StatusSeeOther)
		return
	}

	toID, err := strconv.Atoi(r.FormValue("to_id"))
	if err != nil {
		http.Redirect(w, r, "/regrade?error="+url.QueryEscape("Выберите товар-назначение"), http.StatusSeeOther)
		return
	}

	toQty, err := strconv.Atoi(r.FormValue("to_qty"))
	if err != nil || toQty <= 0 {
		http.Redirect(w, r, "/regrade?error="+url.QueryEscape("Неверное количество для оприходования"), http.StatusSeeOther)
		return
	}

	note := r.FormValue("note")

	if err := models.CreateRegrade(username, fromID, fromQty, toID, toQty, note); err != nil {
		http.Redirect(w, r, "/regrade?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(username, "regrade", r.FormValue("from_id")+"→"+r.FormValue("to_id"), note)
	http.Redirect(w, r, "/regrade?success="+url.QueryEscape("Пересорт оформлен"), http.StatusSeeOther)
}
