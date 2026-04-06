package handlers

import (
	"encoding/csv"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"wims/models"
)

var productsAdminTmpl = template.Must(template.ParseFiles("templates/products_admin.html", "templates/navbar.html"))
var productEditTmpl   = template.Must(template.ParseFiles("templates/product_edit.html", "templates/navbar.html"))

// ProductsAdminPage — страница управления товарами (manager + admin)
func ProductsAdminPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	products, err := models.GetAllProducts()
	if err != nil {
		http.Error(w, "Ошибка загрузки товаров", http.StatusInternalServerError)
		return
	}

	categories, err := models.GetCategories()
	if err != nil {
		http.Error(w, "Ошибка загрузки категорий", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":    display,
		"Role":        role,
		"CurrentUser": username,
		"Products":    products,
		"Categories":  categories,
		"Settings":    settings,
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	}

	if err := productsAdminTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// AddProductAdminHandler — добавление товара (manager + admin)
func AddProductAdminHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
		return
	}

	name     := r.FormValue("name")
	barcode  := r.FormValue("barcode")
	category := r.FormValue("category")
	if category == "" {
		category = r.FormValue("category_new")
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Redirect(w, r, "/admin/products?error="+url.QueryEscape("Неверная цена"), http.StatusSeeOther)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Redirect(w, r, "/admin/products?error="+url.QueryEscape("Неверное количество"), http.StatusSeeOther)
		return
	}

	minStock, _ := strconv.Atoi(r.FormValue("min_stock"))

	if err := models.CreateProduct(name, barcode, category, price, qty, minStock, username); err != nil {
		http.Redirect(w, r, "/admin/products?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(username, "add_product", name, "")
	http.Redirect(w, r, "/admin/products?success="+url.QueryEscape("Товар добавлен"), http.StatusSeeOther)
}

// EditProductPage — страница редактирования товара (manager + admin)
func EditProductPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
		return
	}

	product, err := models.GetProductByID(id)
	if err != nil || product == nil {
		http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		name     := r.FormValue("name")
		barcode  := r.FormValue("barcode")
		category := r.FormValue("category")
		if category == "" {
			category = r.FormValue("category_new")
		}

		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if err != nil {
			http.Redirect(w, r, "/admin/products/edit?id="+idStr+"&error="+url.QueryEscape("Неверная цена"), http.StatusSeeOther)
			return
		}

		minStock, _ := strconv.Atoi(r.FormValue("min_stock"))

		if err := models.UpdateProduct(id, name, barcode, category, price, minStock); err != nil {
			http.Redirect(w, r, "/admin/products/edit?id="+idStr+"&error="+url.QueryEscape("Ошибка сохранения"), http.StatusSeeOther)
			return
		}

		models.WriteAdminLog(username, "edit_product", name, "")
		http.Redirect(w, r, "/admin/products?success="+url.QueryEscape("Товар обновлён"), http.StatusSeeOther)
		return
	}

	categories, _ := models.GetCategories()

	data := map[string]interface{}{
		"Username":   display,
		"Role":       role,
		"Product":    product,
		"Categories": categories,
		"Error":      r.URL.Query().Get("error"),
	}

	if err := productEditTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// DeleteProductAdminHandler — удаление товара (manager + admin)
func DeleteProductAdminHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/admin/products?error="+url.QueryEscape("Неверный ID"), http.StatusSeeOther)
		return
	}

	// получаем имя до удаления для лога
	product, _ := models.GetProductByID(id)
	name := ""
	if product != nil {
		name = product.Name
	}

	if err := models.DeleteProduct(id, username); err != nil {
		http.Redirect(w, r, "/admin/products?error="+url.QueryEscape("Ошибка удаления"), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(username, "delete_product", name, "")
	http.Redirect(w, r, "/admin/products?success="+url.QueryEscape("Товар удалён"), http.StatusSeeOther)
}

// ExportProductsCSVHandler — экспорт товаров в CSV
func ExportProductsCSVHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	data, err := models.ExportProductsCSV()
	if err != nil {
		http.Error(w, "Ошибка экспорта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="products.csv"`)
	// BOM для корректного открытия в Excel
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = ';' // точка с запятой — стандарт для Excel в русской локали
	writer.WriteAll(data)
	writer.Flush()
}

// ExportHistoryCSVHandler — экспорт истории в CSV
func ExportHistoryCSVHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	data, err := models.ExportHistoryCSV()
	if err != nil {
		http.Error(w, "Ошибка экспорта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="history.csv"`)
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = ';' // точка с запятой — стандарт для Excel в русской локали
	writer.WriteAll(data)
	writer.Flush()
}
