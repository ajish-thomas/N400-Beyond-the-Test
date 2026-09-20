// Apply the saved choice before styles paint. Everything stays on this browser.
(() => {
  const key = "n400-theme";
  const choices = ["auto", "light", "dark"];
  let theme = "auto";
  try {
    const saved = localStorage.getItem(key);
    if (choices.includes(saved)) theme = saved;
  } catch (_) { /* Switching still works when storage is unavailable. */ }
  document.documentElement.dataset.theme = theme;
  document.addEventListener("DOMContentLoaded", () => {
    const control = document.querySelector(".theme-control");
    const select = document.getElementById("theme");
    if (!control || !select) return;
    select.value = theme;
    control.hidden = false;
    select.addEventListener("change", () => {
      if (!choices.includes(select.value)) return;
      document.documentElement.dataset.theme = select.value;
      try { localStorage.setItem(key, select.value); } catch (_) { /* Session only. */ }
    });
  });
})();
