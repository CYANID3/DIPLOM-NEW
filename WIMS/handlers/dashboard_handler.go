package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var dashboardTmpl = template.Must(template.ParseFiles("templates/dashboard.html", "templates/navbar.html"))

func DashboardPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireRole(w, r, "admin", "manager")
	if !ok {
		return
	}

	summary, err := models.GetSalesSummary()
	if err != nil {
		http.Error(w, "Ошибка статистики", http.StatusInternalServerError)
		return
	}

	topProducts, err := models.GetTopProducts()
	if err != nil {
		http.Error(w, "Ошибка топ товаров", http.StatusInternalServerError)
		return
	}

	staffStats, err := models.GetStaffStats()
	if err != nil {
		http.Error(w, "Ошибка статистики сотрудников", http.StatusInternalServerError)
		return
	}

	dailyStats, err := models.GetDailyStats()
	if err != nil {
		http.Error(w, "Ошибка дневной статистики", http.StatusInternalServerError)
		return
	}

	lowStock, err := models.GetLowStockProducts()
	if err != nil {
		http.Error(w, "Ошибка остатков", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":    display,
		"Role":        role,
		"CurrentUser": username,
		"Summary":     summary,
		"TopProducts": topProducts,
		"StaffStats":  staffStats,
		"DailyStats":  dailyStats,
		"LowStock":    lowStock,
		"Settings":    settings,
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	}

	if err := dashboardTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
