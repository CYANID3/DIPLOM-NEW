package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"wims/models"
)

var transactionTmpl = template.Must(template.ParseFiles("templates/transaction.html", "templates/navbar.html"))

func TransactionPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireAuth(w, r)
	if !ok {
		return
	}

	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки товаров", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":    display,
		"Role":        role,
		"CurrentUser": username,
		"Products":    products,
		"Settings":    settings,
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	}

	if err := transactionTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// GetProductDataHandler — JSON данные товара для автозаполнения
func GetProductDataHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := RequireAuth(w, r)
	if !ok {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	p, err := models.GetProductByID(id)
	if err != nil || p == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w,
		`{"id":%d,"name":%q,"barcode":%q,"price":%.2f,"quantity":%d,"category":%q}`,
		p.ID, p.Name, p.Barcode, p.Price, p.Quantity, p.Category,
	)
}

func SellTransactionHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/transaction", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/transaction?error="+url.QueryEscape("Выберите товар"), http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		http.Redirect(w, r, "/transaction?error="+url.QueryEscape("Неверное количество"), http.StatusSeeOther)
		return
	}

	if err := models.SellProduct(id, qty, username); err != nil {
		http.Redirect(w, r, "/transaction?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/transaction?success="+url.QueryEscape("Продажа выполнена"), http.StatusSeeOther)
}
