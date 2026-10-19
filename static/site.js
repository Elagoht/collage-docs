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

// The results show only while the search box has the focus. A tap on a result
// would take the focus away first — mobile browsers do not focus a link they
// are tapping on — and the list would close before the tap reached the link,
// so pressing on the list leaves the focus where it is.
document.addEventListener("mousedown", (event) => {
  if (event.target instanceof Element && event.target.closest(".search-results")) event.preventDefault();
});
