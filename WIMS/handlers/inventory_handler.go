package handlers

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"wims/models"
)

var (
	inventoryListTmpl = template.Must(template.ParseFiles("templates/inventory_list.html", "templates/navbar.html"))
	inventoryDocTmpl  = template.Must(template.ParseFiles("templates/inventory_doc.html", "templates/navbar.html"))
)

// InventoryListPage — список всех инвентаризаций
func InventoryListPage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	inventories, err := models.GetInventories()
	if err != nil {
		http.Error(w, "Ошибка загрузки инвентаризаций", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":    display,
		"Role":        role,
		"Inventories": inventories,
		"Settings":    settings,
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	}

	if err := inventoryListTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// CreateInventoryHandler — создать новый документ
func CreateInventoryHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	note := r.FormValue("note")
	id, err := models.CreateInventory(username, note)
	if err != nil {
		http.Redirect(w, r, "/inventory?error="+url.QueryEscape("Ошибка создания инвентаризации"), http.StatusSeeOther)
		return
	}

	noteStr := note
	if noteStr == "" {
		noteStr = "без примечания"
	}
	models.WriteAdminLog(username, "create_inventory", "Документ #"+strconv.Itoa(id), noteStr)
	http.Redirect(w, r, "/inventory/"+strconv.Itoa(id), http.StatusSeeOther)
}

// InventoryDocPage — страница документа инвентаризации
func InventoryDocPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	idStr := r.URL.Path[len("/inventory/"):]
	if idStr == "" {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// сохраняем фактические количества
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/inventory/"+idStr+"?error="+url.QueryEscape("Ошибка формы"), http.StatusSeeOther)
			return
		}
		for key, vals := range r.Form {
			if len(key) > 5 && key[:5] == "item_" {
				itemID, err := strconv.Atoi(key[5:])
				if err != nil {
					continue
				}
				actualQty, err := strconv.Atoi(vals[0])
				if err != nil || actualQty < 0 {
					continue
				}
				models.UpdateInventoryItem(itemID, actualQty)
			}
		}
		http.Redirect(w, r, "/inventory/"+idStr+"?success="+url.QueryEscape("Сохранено"), http.StatusSeeOther)
		return
	}

	inv, items, err := models.GetInventoryByID(id)
	if err != nil || inv == nil {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":    display,
		"Role":        role,
		"CurrentUser": username,
		"Inventory":   inv,
		"Items":       items,
		"Settings":    settings,
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	}

	if err := inventoryDocTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// CompleteInventoryHandler — завершить инвентаризацию
func CompleteInventoryHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/inventory?error="+url.QueryEscape("Неверный ID"), http.StatusSeeOther)
		return
	}

	if err := models.CompleteInventory(id, username); err != nil {
		http.Redirect(w, r, "/inventory/"+idStr+"?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	// собираем статистику для лога
	_, items, _ := models.GetInventoryByID(id)
	surplus, shortage, total := 0, 0, 0
	for _, item := range items {
		if item.Diff > 0 { surplus++ }
		if item.Diff < 0 { shortage++ }
		if item.Diff != 0 { total++ }
	}
	detail := fmt.Sprintf("позиций: %d, излишков: %d, недостач: %d", len(items), surplus, shortage)
	models.WriteAdminLog(username, "complete_inventory", "Документ #"+idStr, detail)
	http.Redirect(w, r, "/inventory/"+idStr+"?success="+url.QueryEscape("Инвентаризация завершена, остатки скорректированы"), http.StatusSeeOther)
}

// ExportInventoryCSVHandler — экспорт строк инвентаризации
func ExportInventoryCSVHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	data, err := models.ExportInventoryCSV(id)
	if err != nil {
		http.Error(w, "Ошибка экспорта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="inventory_`+idStr+`.csv"`)
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = ';'
	writer.WriteAll(data)
	writer.Flush()
}
