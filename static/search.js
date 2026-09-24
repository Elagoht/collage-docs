// Site search. The index is /search.json, a static document the export writes as a
// file: every page split at its headings. It is fetched the first time the box is
// used, never on page load, and searched in the browser — no server is involved.
(() => {
  const box = document.getElementById("search");
  if (!box) return;
  const input = box.querySelector("input");
  const list = box.querySelector("[role=listbox]");
  const MAX = 8;
  let index = null;
  let loading = null;
  let results = [];
  let active = -1;

  function load() {
    if (index) return Promise.resolve(index);
    if (!loading) {
      loading = fetch(box.dataset.index)
        .then((response) => {
          if (!response.ok) throw new Error(response.status + " " + response.statusText);
          return response.json();
        })
        .then((entries) => {
          index = entries.map((entry) => ({
            ...entry,
            lt: entry.t.toLowerCase(),
            lh: (entry.h || "").toLowerCase(),
            lx: entry.x.toLowerCase(),
          }));
          return index;
        })
        .catch((error) => {
          loading = null;
          throw error;
        });
    }
    return loading;
  }

  // A term found in the page title counts most, in a heading next, in the text
  // least; a term found nowhere rules the entry out, so every word typed narrows.
  function score(entry, terms) {
    let total = 0;
    for (const term of terms) {
      let found = 0;
      if (entry.lt.includes(term)) found += entry.lt.startsWith(term) ? 12 : 8;
      if (entry.lh.includes(term)) found += entry.lh.startsWith(term) ? 7 : 5;
      const at = entry.lx.indexOf(term);
      if (at !== -1) {
        found += 1;
        let count = 0;
        for (let i = at; i !== -1 && count < 5; i = entry.lx.indexOf(term, i + term.length)) count++;
        found += count * 0.5;
      }
      if (found === 0) return 0;
      total += found;
    }
    // A page's opening part stands for the page as a whole.
    if (!entry.h) total += 0.5;
    return total;
  }

  function search(query) {
    const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    if (terms.length === 0 || !index) return [];
    const scored = [];
    for (const entry of index) {
      const value = score(entry, terms);
      if (value > 0) scored.push({ entry, value });
    }
    scored.sort((a, b) => b.value - a.value);
    return scored.slice(0, MAX).map((item) => ({ ...item.entry, terms }));
  }

  // The text around the first match, with every term marked. Built from text
  // nodes, never parsed as HTML.
  function snippet(result) {
    const text = result.x;
    const lower = result.lx;
    let at = -1;
    for (const term of result.terms) {
      const i = lower.indexOf(term);
      if (i !== -1 && (at === -1 || i < at)) at = i;
    }
    let start = Math.max(0, at - 60);
    if (start > 0) {
      const space = text.indexOf(" ", start);
      if (space !== -1 && space < at) start = space + 1;
    }
    const piece = text.slice(start, start + 170);
    const fragment = document.createDocumentFragment();
    if (start > 0) fragment.append("…");
    const pattern = new RegExp(result.terms.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|"), "gi");
    let last = 0;
    for (const match of piece.matchAll(pattern)) {
      fragment.append(piece.slice(last, match.index));
      const mark = document.createElement("mark");
      mark.textContent = match[0];
      fragment.append(mark);
      last = match.index + match[0].length;
    }
    fragment.append(piece.slice(last));
    if (start + 170 < text.length) fragment.append("…");
    return fragment;
  }

  function render() {
    list.replaceChildren();
    active = results.length ? 0 : -1;
    results.forEach((result, i) => {
      const option = document.createElement("li");
      option.id = "search-result-" + i;
      option.setAttribute("role", "option");
      const link = document.createElement("a");
      link.href = result.u;
      link.tabIndex = -1;
      const title = document.createElement("span");
      title.className = "result-title";
      title.textContent = result.h ? result.t + " › " + result.h : result.t;
      const section = document.createElement("span");
      section.className = "result-section";
      section.textContent = result.s;
      const text = document.createElement("span");
      text.className = "result-text";
      text.append(snippet(result));
      link.append(title, section, text);
      option.append(link);
      option.addEventListener("mousemove", () => select(i));
      list.append(option);
    });
    if (input.value.trim() && results.length === 0) {
      const empty = document.createElement("li");
      empty.className = "result-empty";
      empty.textContent = box.dataset.empty.replace("{query}", input.value.trim());
      list.append(empty);
    }
    const open = list.childElementCount > 0;
    list.hidden = !open;
    input.setAttribute("aria-expanded", String(open));
    select(active);
  }

  function select(i) {
    active = i;
    [...list.querySelectorAll("[role=option]")].forEach((option, n) => {
      option.setAttribute("aria-selected", String(n === i));
    });
    if (i >= 0) {
      input.setAttribute("aria-activedescendant", "search-result-" + i);
      list.querySelector("#search-result-" + i)?.scrollIntoView({ block: "nearest" });
    } else {
      input.removeAttribute("aria-activedescendant");
    }
  }

  function close() {
    list.hidden = true;
    input.setAttribute("aria-expanded", "false");
    input.removeAttribute("aria-activedescendant");
  }

  input.addEventListener("focus", () => {
    load().then(() => input.value && update()).catch(() => {});
  });

  function update() {
    load()
      .then(() => {
        results = search(input.value);
        render();
      })
      .catch((error) => {
        list.replaceChildren();
        const failed = document.createElement("li");
        failed.className = "result-empty";
        failed.textContent = box.dataset.failed.replace("{error}", error.message);
        list.append(failed);
        list.hidden = false;
      });
  }

  input.addEventListener("input", update);

  input.addEventListener("keydown", (event) => {
    if (event.key === "ArrowDown" && results.length) {
      event.preventDefault();
      select((active + 1) % results.length);
    } else if (event.key === "ArrowUp" && results.length) {
      event.preventDefault();
      select((active - 1 + results.length) % results.length);
    } else if (event.key === "Enter" && active >= 0) {
      event.preventDefault();
      location.href = results[active].u;
    } else if (event.key === "Escape") {
      if (input.value) {
        input.value = "";
        results = [];
        close();
      } else {
        input.blur();
      }
    }
  });

  document.addEventListener("keydown", (event) => {
    const target = event.target;
    const typing = target instanceof HTMLElement && (target.isContentEditable || /^(input|textarea|select)$/i.test(target.tagName));
    if (event.key === "/" && !typing && !event.metaKey && !event.ctrlKey && !event.altKey) {
      event.preventDefault();
      input.focus();
      input.select();
    }
  });

  document.addEventListener("click", (event) => {
    if (!box.contains(event.target)) close();
  });
})();
