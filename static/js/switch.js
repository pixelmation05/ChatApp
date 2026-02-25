const toggle = document.getElementById("toggle");
const html = document.documentElement;

// load saved theme
const saved = localStorage.getItem("theme");
if (saved) {
  html.setAttribute("data-theme", saved);
  toggle.checked = saved === "dark";
}

// listen to toggle
toggle.addEventListener("change", () => {
  const theme = toggle.checked ? "dark" : "light";
  html.setAttribute("data-theme", theme);
  localStorage.setItem("theme", theme);
});