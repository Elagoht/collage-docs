// Posts the form without leaving the page, and puts the list the server answers
// with in place of the old one. Without this script the form posts as any form
// does, and the server redirects back to the page.
const form = document.querySelector("#add");

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(form.action, {
    method: "POST",
    headers: { "Collage-Fetch": "1" },
    body: new FormData(form),
  });
  if (!response.ok) {
    form.submit();
    return;
  }
  document.querySelector("#items").outerHTML = await response.text();
  form.reset();
});

// Fetches the list again every five seconds, so an item added in another tab
// shows up here too.
setInterval(async () => {
  const items = document.querySelector("#items");
  const response = await fetch(items.dataset.src);
  if (response.ok) {
    items.outerHTML = await response.text();
  }
}, 5000);
