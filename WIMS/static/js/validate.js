/**
 * validate.js — глобальная валидация числовых полей
 *
 * Подход: используем стандартный механизм браузера (Constraint Validation API).
 * Браузер сам показывает нативный попап «заполните это поле» / «значение должно
 * быть не менее 0» — точно так же как у input[type=email].
 *
 * Всё что нужно — правильно выставить атрибуты min/required на инпутах
 * и добавить novalidate на форму, чтобы управлять моментом показа.
 *
 * Дополнительно: при вводе отрицательного числа сразу подсвечиваем поле
 * красной рамкой через класс is-invalid (Bootstrap), не дожидаясь submit.
 */

document.addEventListener('DOMContentLoaded', function () {

    // ── 1. Живая подсветка: is-invalid при отрицательном значении ────────────
    // Находим все number-инпуты у которых есть атрибут min
    document.querySelectorAll('input[type="number"][min]').forEach(function (inp) {
        var minVal = parseFloat(inp.getAttribute('min'));

        function check() {
            if (inp.value === '') {
                // пустое поле — не красим заранее, браузер сам скажет при submit
                inp.classList.remove('is-invalid', 'is-valid');
                return;
            }
            var v = parseFloat(inp.value);
            if (isNaN(v) || v < minVal) {
                inp.classList.add('is-invalid');
                inp.classList.remove('is-valid');
            } else {
                inp.classList.remove('is-invalid');
                inp.classList.add('is-valid');
            }
        }

        inp.addEventListener('input', check);
        inp.addEventListener('blur', check);
    });

    // ── 2. Submit: даём браузеру показать нативные попапы ─────────────────────
    // Для форм с novalidate вызываем reportValidity() — это и есть нативный
    // попап прямо у поля, точно как у type=email
    document.querySelectorAll('form[novalidate]').forEach(function (form) {
        form.addEventListener('submit', function (e) {
            if (!form.checkValidity()) {
                e.preventDefault();
                form.reportValidity(); // браузер сам выберет первый невалидный инпут и покажет попап
            }
        });
    });

    // ── 3. Сброс подсветки при открытии Bootstrap-модалок ────────────────────
    document.querySelectorAll('.modal').forEach(function (modal) {
        modal.addEventListener('show.bs.modal', function () {
            modal.querySelectorAll('input').forEach(function (inp) {
                inp.classList.remove('is-valid', 'is-invalid');
            });
        });
    });

});
