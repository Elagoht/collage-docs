# Aşama 3: Öğretici ve örnek proje Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Pluginsiz, sadece `WithFragmentPath`, bir action ve birkaç satır vanilla JS ile sayfadaki bir alanı sunucudan güncelleyen çalışan bir örnek proje ve onu baştan sona kurduran bir öğretici (EN + TR).

**Architecture:** `examples/shopping-list/`, `collage new shopping-list` çıktısının kendisidir: ayrı bir Go modülü (`module shopping-list`). Okuyucu aynı komutla aynı iskeleti alır. Öğreticinin gösterdiği her dosya bu dizindeki dosyanın birebir aynısıdır, bir test bunu kontrol eder. Kök modüldeki bir test, örneğin kendi testlerini `go test ./...` ile çalıştırır. Böylece CI hem örneği hem öğreticiyi doğru tutar.

**Tech Stack:** collage v0.65.0, `collagetest`, vanilla JS, goldmark (ham HTML açık).

**Spec:** `docs/superpowers/specs/2026-10-17-docs-positioning-design.md` (Kararlar → 5. Öğretici ve örnek proje)

## Global Constraints

- Plugin yok. Sadece çekirdek: page, fragment, `WithFragmentPath`, `WithAction`, `{{csrfToken}}`, `collage.FetchHeader`, `RenderFragment`, `SeeOther`, `InvalidateTags`, `LoadTagged`.
- JavaScript kapalıyken de çalışır: form normal post eder, action 303 ile page'e döner.
- Başka hiçbir framework, kütüphane ya da ürün adı geçmez (site metni, kod, test, commit mesajı).
- Go kodunda `any` yok. Metin bugünkü davranışı anlatır.
- Öğretici "bunu yerelde çalıştır" der; sitede canlı demo vaat edilmez.
- Yeni sayfa iki `nav.json`'a, "Getting started" bölümüne `your-first-page`'den sonra girer. TR başlıkları EN ile seviye seviye aynıdır.
- TR'de "parça" kullanılmaz, "kısım" kullanılır; terimler mevcut TR sayfalarındaki gibidir.

## Review Focus

- Boş ya da yalnızca boşluk olan bir öğe adı: eklenmez, liste değişmez, cevap yine 303 (form) ya da fragment (fetch) olur (`TestAddIgnoresBlank`).
- Aynı page başka bir sekmede açıkken eklenen öğe, yenileme ile görünür; fragment path cevabı cache'lenmez (`TestItemsFragmentIsFresh`: ekledikten sonra `/list/items` yeni öğeyi içerir).
- CSRF: token'sız bir POST reddedilir (`TestAddWithoutTokenIsRefused`: 403).
- Öğe adındaki HTML kaçırılır (`TestItemNameIsEscaped`: `<b>` gövdede `&lt;b&gt;` olarak çıkar).
- Eşzamanlı ekleme: store bir mutex ile korunur (`go test -race`).

---

### Task 1: Örnek proje

**Files:**
- Create: `examples/shopping-list/` (`collage new shopping-list --dir examples/shopping-list` çıktısı, `go mod tidy` ile `go.sum`)
- Create: `examples/shopping-list/list/list.go`
- Create: `examples/shopping-list/fragments/pages/list.go`
- Create: `examples/shopping-list/pages/list.go`
- Create: `examples/shopping-list/templates/pages/list.html`, `templates/fragments/items.html`
- Create: `examples/shopping-list/static/list.js`
- Modify: `examples/shopping-list/routes.go` (`pages.List()` eklenir)
- Create: `examples/shopping-list/list_test.go`
- Test: kök `main_test.go` → `TestExampleShoppingList`

**Interfaces:**
- `package list`: `type Item struct { Name string }`; `type List struct` (mutex + `[]Item`); `func New(names ...string) *List`; `func (l *List) Items() []Item` (kopya döner); `func (l *List) Add(name string) bool` (kırpılmış ad boşsa `false`); `var Shopping = New("Milk", "Bread")`; `const Tag = "list"`.
- `package fragments`: `func Items() *collage.Fragment` (`"items"`, `fragments/items.html`, `LoadTagged` → `[]list.Item`, tag `list.Tag`); `func List(items *collage.Fragment) *collage.Fragment` (`"list"`, `pages/list.html`, slot `"items"`).
- `package pages`: `func List() *collage.Page` — `/list`, `WithFragmentPath("en", "/list/items", items)`, `WithAction("POST", add)`; `add` ekler, `InvalidateTags: []string{list.Tag}`; istek `collage.FetchHeader` taşıyorsa `RenderFragment(items)`, taşımıyorsa `SeeOther("/list")`.
- `static/list.js`: `#add` formunun submit'ini `fetch` ile `Collage-Fetch: 1` header'ı ve `FormData` ile gönderir, cevabı `#items`'ın `outerHTML`'ine koyar, formu temizler; cevap başarısızsa formu normal gönderir. Her 5 saniyede `#items`'ın `data-src`'sinden fragment'i çekip yerine koyar.

- [ ] **Step 1:** İskeleti üret, `go mod tidy`, `go test ./...` (iskelet testi yok, derlenir).
- [ ] **Step 2: Failing testleri yaz** (`list_test.go`): `TestListPage` (GET `/list` 200, "Milk" ve `data-src="/list/items"` var), `TestItemsFragment` (GET `/list/items` 200, `<html` yok, `id="items"` var), `TestAddWithForm` (`Submit(page, "/list", name=Eggs)` → 303, `Location` `/list`, ardından GET `/list` "Eggs" içerir), `TestAddWithFetch` (form değerleri + token ile `Collage-Fetch: 1` başlıklı POST → 200, gövde `id="items"` ve "Butter" içerir, `<html` içermez), `TestAddIgnoresBlank`, `TestItemsFragmentIsFresh`, `TestAddWithoutTokenIsRefused`, `TestItemNameIsEscaped`. Her test kendi `list.Shopping`'ini sıfırlar (`list.Shopping = list.New("Milk", "Bread")`) ve `cacheDir = t.TempDir()` kullanır.
- [ ] **Step 3:** `go test ./...` → FAIL (`/list` 404).
- [ ] **Step 4:** Kodu yaz (Interfaces'e göre).
- [ ] **Step 5:** `go test -race ./...` → PASS; `go vet ./...` temiz.
- [ ] **Step 6:** Kök `TestExampleShoppingList`: `exec.Command("go", "test", "./...")`, `Dir: "examples/shopping-list"`; hata çıktısıyla düşer. `go test ./...` kökte PASS.
- [ ] **Step 7:** Elle dene: `cd examples/shopping-list && go run .`, iki sekmede `/list`; birinde ekle, diğerinde 5 saniye içinde görün; JS kapalıyken form çalışır.
- [ ] **Step 8: Commit** — `examples: a shopping list that updates one part of its page`

---

### Task 2: Öğretici sayfası (EN + TR)

**Files:**
- Create: `content/update-part-of-a-page.md`, `content/tr/update-part-of-a-page.md`
- Modify: `content/nav.json`, `content/tr/nav.json`
- Modify: `content/your-first-page.md` ve TR (`Where to go next`'e link)
- Test: `main_test.go` → `TestTutorialShowsTheExample`

**Interfaces:**
- Consumes: Task 1'in dosyaları.
- Produces: her kod bloğundan önce `<!-- file: <yol> -->` işareti; test, işaretten sonraki ilk kod bloğunun `examples/shopping-list/<yol>` ile birebir aynı olduğunu iki dilde kontrol eder.

Başlıklar:

```
# Update part of a page
## What you will build
## Start a project
## The list
## The items fragment
## The page
## Open the fragment at its own URL
## Add an item
## Swap it in with a few lines of JavaScript
## Keep it current
## Run it
## Test it
## Where to go next
```

- Bir SVG şema ("Add an item" altında): tarayıcı → `POST /list` (`Collage-Fetch`) → action: ekle + `InvalidateTags("list")` → `RenderFragment(items)` → tarayıcı `#items`'ı değiştirir; JS yokken aynı POST → 303 → `GET /list`. `figure.diagram-figure`, `role="img"`, `data-search-ignore`, renk sınıflarla.
- Her adımda neden: fragment path neden açıkça tanımlanır, fragment path cevabı neden cache'lenmez, page neden tag'le invalidate edilir, CSRF token'ı `FormData` ile nasıl taşınır.
- "Keep it current": polling; daha fazlası için `elagoht/live` linki (plugin olduğu açıkça söylenir).

- [ ] **Step 1: Failing test** `TestTutorialShowsTheExample`: iki dildeki sayfada en az 7 `<!-- file:` işareti; her biri dosyayla birebir aynı.
- [ ] **Step 2:** → FAIL (sayfa yok).
- [ ] **Step 3:** EN sayfası + nav.
- [ ] **Step 4:** TR sayfası + nav.
- [ ] **Step 5:** `go run ./tools/checkcontent`, `go test ./...`, export → temiz.
- [ ] **Step 6:** Görsel kontrol 1280 ve 375.
- [ ] **Step 7:** `turkish-reviewer` (`update-part-of-a-page`).
- [ ] **Step 8: Commit** — `docs: a tutorial that updates one part of a page`
