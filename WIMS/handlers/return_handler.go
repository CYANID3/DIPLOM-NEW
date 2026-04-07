package handlers

import (
	"encoding/csv"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"wims/models"
)

var returnTmpl = template.Must(template.ParseFiles("templates/return.html", "templates/navbar.html"))

func ReturnPage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	sells, err := models.GetSellHistory()
	if err != nil {
		http.Error(w, "Ошибка загрузки истории продаж", http.StatusInternalServerError)
		return
	}

	returns, err := models.GetReturns()
	if err != nil {
		http.Error(w, "Ошибка загрузки возвратов", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Sells":    sells,
		"Returns":  returns,
		"Settings": settings,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
	}

	if err := returnTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func CreateReturnHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/returns", http.StatusSeeOther)
		return
	}

	historyID, err := strconv.Atoi(r.FormValue("history_id"))
	if err != nil {
		http.Redirect(w, r, "/returns?error="+url.QueryEscape("Выберите продажу"), http.StatusSeeOther)
		return
	}

	productID, err := strconv.Atoi(r.FormValue("product_id"))
	if err != nil {
		http.Redirect(w, r, "/returns?error="+url.QueryEscape("Неверный товар"), http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		http.Redirect(w, r, "/returns?error="+url.QueryEscape("Неверное количество"), http.StatusSeeOther)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	productName := r.FormValue("product_name")
	barcode     := r.FormValue("barcode")
	note        := r.FormValue("note")

	if err := models.CreateReturn(username, historyID, productID, qty, productName, barcode, note, price); err != nil {
		http.Redirect(w, r, "/returns?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(username, "return", productName, strconv.Itoa(qty)+" шт.")
	http.Redirect(w, r, "/returns?success="+url.QueryEscape("Возврат оформлен"), http.StatusSeeOther)
}

func ExportReturnsCSVHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	returns, err := models.GetReturns()
	if err != nil {
		http.Error(w, "Ошибка экспорта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="returns.csv"`)
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = ';'
	writer.Write([]string{"ID", "Пользователь", "Товар", "Штрихкод", "Кол-во", "Цена", "Сумма", "Примечание", "Время"})
	for _, ret := range returns {
		writer.Write([]string{
			strconv.Itoa(ret.ID),
			ret.UserDisplay,
			ret.ProductName,
			ret.Barcode,
			strconv.Itoa(ret.Quantity),
			strconv.FormatFloat(ret.Price, 'f', 2, 64),
			strconv.FormatFloat(ret.ReturnTotal, 'f', 2, 64),
			ret.Note,
			ret.Timestamp,
		})
	}
	writer.Flush()
}
