# 📦 WIMS - Web-based Inventory Management System

**WIMS** - веб-система для управления складом: учёт товаров, контроль остатков и продаж.

## 🚀 Возможности
- Учёт товаров (добавление, редактирование, удаление)
- Продажи с проверкой остатка
- Пользователи и роли (admin / user)
- Админ-панель (управление пользователями)
- Логи действий (admin log)
- Пересрт и инвентаризация 
- Статистика и контроль остатков

## 🧱 Стек
- Go (net/http)
- SQLite (modernc.org/sqlite)
- HTML templates (`html/template`)
- Bootstrap (frontend)

## 📁 Структура

WIMS/
├── main.go
├── database/ # инициализация БД, миграции
├── models/ # бизнес-логика и SQL
├── handlers/ # HTTP-обработчики
├── templates/ # HTML-шаблоны
└── static/ # статика

## ⚙️ Запуск
```bash
git clone https://github.com/CYANID3/DIPLOM-NEW.git
cd wims
go mod tidy
go run main.go
```
Открыть: http://localhost:8080

##🔐 Доступ

При первом запуске создаётся пользователь:

login: admin
password: admin

## 📌 Особенности
- Нельзя продать больше, чем есть
- Автоматический пересчёт остатков
- Простая админ-панель

## 🧠 Архитектура

Проект построен по MVC-подобному принципу:

- **main.go** - инициализация приложения, роутинг
- **database/** - подключение к БД, миграции
- **models/** - работа с данными и SQL-запросы
- **handlers/** - обработка HTTP-запросов (бизнес-логика)
- **templates/** - HTML-шаблоны (view)
- **static/** - CSS/JS

### 🔄 Поток запроса
1. HTTP-запрос приходит в `handler`
2. Handler вызывает `models`
3. Models работают с БД
4. Handler передаёт данные в `template`
5. Возвращается HTML-ответ

---

## 🌐 API (роуты)

### 🔐 Аутентификация
- `GET /login` - страница входа
- `POST /login` - авторизация
- `GET /logout` - выход

### 📦 Товары
- `GET /products` - список товаров
- `GET /product/add` - форма добавления
- `POST /product/add` - создать товар
- `GET /product/edit?id={id}` - форма редактирования
- `POST /product/edit` - обновить товар
- `POST /product/delete` - удалить товар

### 💰 Продажи
- `GET /sales` - список продаж
- `POST /sales/create` - оформить продажу

### 👤 Пользователи (admin)
- `GET /admin/users` - список пользователей
- `GET /admin/users/add` - форма создания
- `POST /admin/users/add` - создать пользователя

### 📊 Дополнительно
- `GET /` - дашборд
- `GET /profile` - профиль пользователя

---

## 🔐 Авторизация

Используется middleware:

- `RequireAuth` - доступ только для авторизованных
- `RequireRole("admin")` - доступ только для администратора

Применяется на уровне роутов.

---

## 🗄️ Модель данных (упрощённо)

### Product
- id
- name
- barcode
- price
- quantity

### User
- id
- username
- password
- role

### Sale
- id
- product_id
- quantity
- total
- created_at

---

## ⚠️ Логика и ограничения

- Продажа невозможна, если `quantity < requested`
- При продаже:
  - уменьшается остаток
  - создаётся запись в `sales`
- Удаление/изменение проверяется на корректность

---

## 🔍 Логирование

- HTTP-запросы логируются
- Действия администратора фиксируются (admin log)

---

## 🔧 Расширяемость

Текущая архитектура позволяет:
- легко добавить REST API
- заменить SQLite на PostgreSQL
- вынести frontend (например, React)
- добавить кэш или очередь
