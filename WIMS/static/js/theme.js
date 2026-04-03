(function () {
    const saved = localStorage.getItem("theme");

    if (saved) {
        document.documentElement.setAttribute("data-theme", saved);
    } else {
        document.documentElement.setAttribute("data-theme", "light");
    }
})();

function toggleTheme() {
    const current = document.documentElement.getAttribute("data-theme");
    const next = current === "dark" ? "light" : "dark";

    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem("theme", next);
}