# 📦 WIMS - Web-based Inventory Management System

**WIMS** - веб-система для управления складом: учёт товаров, контроль остатков и продаж.

## 🚀 Возможности
- Дашборд (продажи, статистика, остатки)
- Управление товарами (CRUD, штрихкоды)
- Продажи с проверкой наличия
- Пользователи и роли
- Инвентаризация

## 🧱 Стек
- Go (backend)
- SQLite (БД)
- HTML/CSS (Bootstrap)
- html/template

## 📁 Структура

## ⚙️ Запуск
```bash
git clone https://github.com/your-username/wims.git
cd wims
go mod tidy
go run main.go
```
Открыть: http://localhost:8080

## 📌 Особенности
- Нельзя продать больше, чем есть
- Автоматический пересчёт остатков
- Простая админ-панель
