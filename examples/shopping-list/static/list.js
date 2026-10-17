// Posts the form without leaving the page, and puts the list the server answers
// with in place of the old one. Without this script the form posts as any form
// does, and the server redirects back to the page.
const form = document.querySelector("#add");

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    const response = await fetch(form.action, {
      method: "POST",
      headers: { "Collage-Fetch": "1" },
      body: new FormData(form),
    });
    const items = document.querySelector("#items");
    if (!response.ok || !items) {
      throw new Error(`answered ${response.status}`);
    }
    items.outerHTML = await response.text();
    form.reset();
  } catch {
    // No answer, or not the list: post the form the ordinary way, and let the
    // browser show whatever the server says.
    form.submit();
  }
});

// Fetches the list again every five seconds, so an item added in another tab
// shows up here too. A tick the server does not answer is skipped.
setInterval(async () => {
  const items = document.querySelector("#items");
  if (!items) {
    return;
  }
  try {
    const response = await fetch(items.dataset.src);
    if (response.ok) {
      items.outerHTML = await response.text();
    }
  } catch {
    // Try again on the next tick.
  }
}, 5000);
