// ===== ТЕМА =====
(function () {
    const saved = localStorage.getItem("theme") || "light";
    document.documentElement.setAttribute("data-theme", saved);
})();

function toggleTheme() {
    const current = document.documentElement.getAttribute("data-theme");
    const next = current === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem("theme", next);
}

// ===== МОДАЛКА ВЫХОДА =====
// Вставляем модалку в DOM сразу как только документ готов.
// Перехватываем все ссылки href="/logout" на странице.
document.addEventListener("DOMContentLoaded", function () {
    // вставить модалку один раз
    if (!document.getElementById("logoutModal")) {
        const modalHTML = `
<div class="modal fade" id="logoutModal" tabindex="-1">
<div class="modal-dialog modal-dialog-centered modal-sm">
<div class="modal-content">
<div class="modal-header">
<h5 class="modal-title">Выход из аккаунта</h5>
<button type="button" class="btn-close" data-bs-dismiss="modal"></button>
</div>
<div class="modal-body">
<p class="mb-0">Вы уверены, что хотите выйти?</p>
</div>
<div class="modal-footer">
<button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
<a href="/logout" class="btn btn-danger" id="logoutConfirmBtn">Выйти</a>
</div>
</div>
</div>
</div>`;
        document.body.insertAdjacentHTML("beforeend", modalHTML);
    }

    // перехватить все ссылки на /logout
    document.querySelectorAll('a[href="/logout"]').forEach(function (link) {
        link.addEventListener("click", function (e) {
            // на странице профиля уже есть своя логика с unsavedModal —
            // там logoutLink перехватывается отдельно и может показать другую модалку.
            // Если unsavedModal существует и форма грязная — не вмешиваемся.
            if (document.getElementById("unsavedModal") && window._profileFormDirty) {
                return; // профиль сам разберётся
            }
            e.preventDefault();
            new bootstrap.Modal(document.getElementById("logoutModal")).show();
        });
    });
});
