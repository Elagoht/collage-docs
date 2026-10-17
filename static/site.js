// Pressing "/" anywhere but in a field puts the cursor in the search box.
document.addEventListener("keydown", (event) => {
  const target = event.target;
  const typing = target instanceof HTMLElement && (target.isContentEditable || /^(input|textarea|select)$/i.test(target.tagName));
  if (event.key !== "/" || typing || event.metaKey || event.ctrlKey || event.altKey) return;
  const input = document.querySelector(".search-input");
  if (!input) return;
  event.preventDefault();
  input.focus();
  input.select();
});
