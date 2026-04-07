// Вставляет модалку выхода и перехватывает ссылки /logout.
// Подключать ПОСЛЕ bootstrap.bundle.min.js.
(function () {
    // вставляем модалку один раз
    const existing = document.getElementById("logoutModal");
    if (!existing) {
        document.body.insertAdjacentHTML("beforeend", [
            '<div class="modal fade" id="logoutModal" tabindex="-1" data-bs-backdrop="static" data-bs-keyboard="false">',
            '<div class="modal-dialog modal-dialog-centered modal-sm">',
            '<div class="modal-content">',
            '<div class="modal-header">',
            '<h5 class="modal-title">Выход из аккаунта</h5>',
            '</div>',
            '<div class="modal-body">',
            '<p class="mb-0">Вы уверены, что хотите выйти?</p>',
            '</div>',
            '<div class="modal-footer">',
            '<button type="button" class="btn btn-secondary" id="logoutCancelBtn">Отмена</button>',
            '<button type="button" class="btn btn-danger" id="logoutConfirmBtn">Выйти</button>',
            '</div>',
            '</div></div></div>'
        ].join(""));
    }

    const modalEl  = document.getElementById("logoutModal");
    const modal    = new bootstrap.Modal(modalEl);

    document.getElementById("logoutCancelBtn").addEventListener("click", function () {
        modal.hide();
    });

    document.getElementById("logoutConfirmBtn").addEventListener("click", function () {
        // убираем модалку из DOM полностью перед переходом
        modal.hide();
        modalEl.addEventListener("hidden.bs.modal", function () {
            window.location.href = "/logout";
        }, { once: true });
    });

    // перехватываем все ссылки /logout на странице
    // querySelectorAll запускаем здесь — скрипт уже в конце body, DOM готов
    document.querySelectorAll('a[href="/logout"]').forEach(function (link) {
        link.addEventListener("click", function (e) {
            e.preventDefault();
            modal.show();
        });
    });
}());

// ===== АВАТАР =====
// Берём первый символ логина через JS (поддержка кириллицы)
(function () {
    var el = document.getElementById('userAvatar');
    if (el && el.dataset.username) {
        // [...str][0] корректно берёт первый Unicode-символ
        el.textContent = [...el.dataset.username][0].toUpperCase();
    }
}());

// ===== АВАТАР =====
// берём первый символ как руну (корректно для кириллицы)
(function () {
    var el = document.getElementById('userAvatar');
    if (!el) return;
    var username = el.getAttribute('data-username') || '?';
    // Array.from корректно разбивает строку по Unicode code points
    var first = Array.from(username)[0] || '?';
    el.textContent = first.toUpperCase();
}());
