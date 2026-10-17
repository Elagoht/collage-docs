# Aşama 2: Anasayfa ve Giriş sayfası Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Anasayfa ve `introduction` sayfası Collage'ın felsefesini anlatsın: client framework'üne gerek yok; component yerine fragment; cache, tag, invalidate ve purge; sınırsız middleware ve plugin; hâlâ Go yazıyorsun; küçük çekirdekle hem site hem çok kullanıcılı, rol bazlı web uygulaması. EN ve TR birlikte.

**Architecture:** `templates/pages/home.html` spec'teki sıraya göre yeniden kurulur. Her bölüm bir "feature" satırıdır: solda metin, sağda kod ya da SVG şema. Metinler `locales/{en,tr}.json` içindeki `home.*` anahtarlarındadır. Kod blokları dile bağlı olmadığı için şablonda durur ve `elagoht/highlight` tarafından renklendirilir (`<pre><code class="language-go">`). `content/introduction.md` ve TR karşılığı aynı sırayla yeniden yazılır, bir mimari şema içerir. Şemalar satır içi SVG'dir ve `static/app.css` içindeki `d-*` sınıflarını kullanır.

**Tech Stack:** Go template, collage v0.65.0, elagoht/i18n (strict), elagoht/highlight, goldmark (raw HTML açık), düz CSS.

**Spec:** `docs/superpowers/specs/2026-10-17-docs-positioning-design.md` (Felsefe, Kararlar 2–4)

## Global Constraints

- Mesaj sırası (Furkan onayladı): 1) hero "Sitenizin client framework'üne ihtiyacı yok" 2) component değil fragment 3) cache, tag'ler, invalidation, purge 4) sınırsız middleware ve plugin 5) hâlâ Go yazıyorsun (`app.Handle`) 6) küçük çekirdek, hem web sitesi hem web uygulaması 7) geliştirme ve dağıtım şemaları 8) rehber haritası.
- Başka hiçbir framework, kütüphane ya da ürün adı geçmez; belirli bir projeye atıf yok. "Çok kullanıcılı, rol bazlı web uygulamaları yazılabilir" genel bir yetenek olarak söylenir.
- Uydurma rakam yok: sunucudaki JavaScript runtime'larıyla kıyas genel ifadelerle yapılır (daha yavaş, daha çok RAM ve disk, bir runtime ve bir bağımlılık ağacı).
- Çekirdek ile plugin ayrımı doğru yapılır. **Çekirdek:** fragment, fragment path, action, CSRF, cache/tag/invalidate/purge, guard, `app.Use`, `app.Handle`, static export. **Plugin:** session, honeypot, validate, flash, live (SSE), websocket, morph.
- "Batteries-included değil": ORM yok, migration yok, veritabanı katmanı yok. "Kuyruk yok" **denmez**, çünkü `elagoht/jobs` kuyruk sunuyor.
- Middleware karşılaştırması isim vermeden yapılır: tek bir middleware dosyası veren framework'lerde her şey o dosyada elle zincirlenir; collage'da `app.Use` istediğin kadar, her plugin kendi middleware'ini getirir, sıra listedeki sıradır.
- Yayımlanmış plugin sayısı: 44 (`content/plugins.md` açıklaması). Sayı metne yazılacaksa bu kaynaktan alınır.
- `elagoht/i18n` `strict: true` çalışır: her anahtar iki dilde de bulunur. Kullanılmayan eski `home.*` anahtarları silinir.
- TR sayfası orijinalin başlıklarını seviye seviye izler. TR'de terimler mevcut TR sayfalarındaki gibi İngilizce kalır (page, fragment, home page, canonical…).
- Go örneklerinde `any` yok. Metin bugünkü davranışı anlatır, "since vX" ya da "eskiden" anlatısı eklenmez.
- Şema kuralları: `<figure>` + `<svg role="img" aria-labelledby>` + `<figcaption>`. Markdown içindeki SVG bloklarında boş satır olmaz. Dar ekran için ya dar bir `viewBox` ya da `only-wide`/`only-narrow` ikilisi kullanılır.

## Review Focus

- 360px genişlikte yatay kaydırma olmamalı. Kod blokları kendi içinde kaymalı, sayfa kaymamalı (`TestHomeCodeScrolls`: `.feature pre` için `overflow-x: auto` kuralı CSS'te var).
- Koyu temada şemalar okunur kalmalı. Yeni SVG'lerde sabit renk yazılmaz, sadece `d-*` sınıfları kullanılır (`TestDiagramsUseClasses`: home.html ve introduction SVG'lerinde `fill="#` ya da `stroke="#` geçmiyor).
- Plugin ızgarasındaki her link `plugins` sayfasında var olan bir başlığa gider, iki dilde de (`TestHomePluginLinks`).
- Ekran okuyucu: bilgi taşıyan şemalarda `role="img"` ve bir etiket var, süs olanlarda `aria-hidden` (`TestIntroDiagramAccessible`).
- Arama indeksi: `home` zaten hariç. Introduction'daki SVG metni arama sonuçlarını kirletmemeli; figure `data-search-ignore` taşır.

---

### Task 1: Anasayfa

**Files:**
- Modify: `templates/pages/home.html` (yeniden yapılandırılır)
- Modify: `locales/en.json`, `locales/tr.json` (`home.*`)
- Modify: `static/app.css` (yeni `.feature`, `.plugin-grid`, `.scope` blokları; mevcut `.points` kaldırılırsa kuralları da gider)
- Test: `main_test.go`

**Interfaces:**
- Consumes: `pageURL "doc" "slug" "<slug>"`, `t`/`th` şablon fonksiyonları, mevcut `d-*` SVG sınıfları ve `#d-arrow` marker'ı.
- Produces: `.feature` bölümleri `id="why" | "fragments" | "cache" | "plugins" | "go" | "apps"`. Task 2'nin şemaları da `.diagram` ve `d-*` sınıflarını kullanır.

Sayfa yapısı ve her bölümün içeriği:

| # | `id` | Metin anahtarları | Sağ taraf |
|---|---|---|---|
| hero | — | `home.eyebrow`, `home.headline` ("Your site doesn't need a <mark>client framework</mark>."), `home.tagline` (HTML + biraz vanilla JS; sunucu sadece değişen parçayı gönderir; tek Go binary'si) | wordmark ve kurulum komutları (aynen) |
| 1 | `why` | `home.why.title`, `home.why.text` (landing, ürün sayfası, birkaç sayfalık site; JS runtime + bağımlılık ağacı + client bundle + hydration ile tek binary + HTML karşılaştırması; rakam yok) | yeni SVG: iki sütun, "typical stack" katmanları ve "collage" katmanları |
| 2 | `fragments` | `home.fragments.title` ("Fragments, not components"), `.text`, `.note` (fetch/XHR, SSE, WebSocket zaten bunun için var; morph ve otomatik yenileme `elagoht/live` ile) | Go: `WithFragmentPath("en", "/search/results", results)`; JS: `forms-and-actions`'taki 3 satırlık `fetch` + `outerHTML`; altında mevcut "layouts" şeması |
| 3 | `cache` | `home.cache.title`, `.text` (page'ler ve veri tag taşır; `InvalidateTags` stale yapıp arka planda yeniden render eder; `PurgeTags` hiçbir şey bırakmadan siler) | Go: `app.InvalidateTags(ctx, "author:ada")` ve `app.PurgeTags(ctx, "user:42")`; altında mevcut "invalidate" şeması |
| 4 | `plugins` | `home.plugins.title`, `.text` (tek middleware dosyası sunan framework'lerle isimsiz kıyas; Collage'da `app.Use` istediğin kadar; plugin'ler kendi middleware'leriyle gelir), `.groups.*` (grup adları) | Go: `Plugins: []collage.Plugin{session.New(…), honeypot.New(…), live.New(…)}` + `app.Use(requestID)`; altında plugin ızgarası |
| 5 | `go` | `home.go.title` ("Still Go"), `.text` (mevcut `http.Handler`/mux'ını `app.Handle("/api/", api)` ile tak; data katmanın senin Go kodun; tek proje, tek binary) | Go: `middleware-and-apis`'teki `app.Handle` örneği |
| 6 | `apps` | `home.apps.title`, `.text` (çekirdek küçük; batteries-included değil), `.in.*` (çekirdekte), `.plugin.*` (plugin ile), `.out.*` (yok: ORM, migration, veritabanı katmanı), `.roles` (çok kullanıcılı, rol bazlı web uygulaması: session + guard + CSRF + honeypot + live) | üç sütunlu `.scope` listesi |
| 7 | `how` | `home.howItWorks`, mevcut `home.diagrams.{fetch,live,dev,build}` | mevcut şemalar (layouts ve invalidate 2. ve 3. bölümlere taşındı) |
| 8 | `map` | `home.theGuide` | aynen |

Plugin ızgarası (`home.plugins.groups.*` başlıkları, linkler `{{pageURL "doc" "slug" "plugins"}}#elagoht<name>`):
- Forms and state: validate, honeypot, flash, session, uploads
- Security: secure, ratelimit, basicauth, oauth, fail2ban, consent
- Live updates: live, websocket
- SEO and content: meta, jsonld, sitemap, robots, feed, ogimage, markdown, highlight, toc, search, i18n
- Delivery and operations: minimizer, compress, bundle, opti-image, deploy, prometheus, otel, health, jobs, accesslog

`home.title` ve `home.description` yeni konumlandırmaya göre yazılır (ör. "collage — server-rendered Go sites and web apps without a client framework").

- [ ] **Step 1: Failing testleri yaz** (`main_test.go`)

```go
// The home page tells the story in the order the spec fixes, in every language.
func TestHomeSections(t *testing.T) {
	for _, path := range []string{"/en/", "/tr/"} {
		body := get(t, path).Body.String()
		last := -1
		for _, id := range []string{`id="why"`, `id="fragments"`, `id="cache"`, `id="plugins"`, `id="go"`, `id="apps"`, `id="how-title"`, `id="guide-title"`} {
			i := strings.Index(body, id)
			if i < 0 || i < last {
				t.Errorf("GET %s: %s missing or out of order", path, id)
			}
			last = i
		}
		for _, want := range []string{"WithFragmentPath", "InvalidateTags", "PurgeTags", "app.Handle", `class="chroma"`} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s has no %s", path, want)
			}
		}
	}
}
```

`TestHomePluginLinks`: `/en/` ve `/tr/` gövdesinden `docs/plugins/#([a-z0-9-]+)` eşleşmelerini topla (en az 30). Her biri için `/en/docs/plugins/` (TR'de `/tr/docs/plugins/`) gövdesinde `id="<anchor>"` olmalı.

`TestHomeCodeScrolls`: `static/app.css` içinde `.feature pre` seçicisine ait blokta `overflow-x: auto` var.

`TestDiagramsUseClasses`: `templates/pages/home.html` içinde `fill="#` ya da `stroke="#` yok.

- [ ] **Step 2:** `go test ./ -run 'TestHome|TestDiagramsUseClasses'` → FAIL (`id="why"` yok, plugin linkleri yok, `.feature pre` yok).

- [ ] **Step 3: Şablonu yeniden kur.** Mevcut `diagram-defs` sembolleri korunur. `layouts` ve `invalidate` figürleri bölüm 2 ve 3'e taşınır. Kod blokları `<pre><code class="language-go">` ve `<pre><code class="language-js">` olur, HTML kaçışlı yazılır. `.points` bölümü kaldırılır; içeriği `fragments`, `cache` ve `apps` bölümlerine dağılır, `points.served` metni `how` bölümündeki build şemasının yanında kalır. Yeni "why" SVG'si mevcut `d-frame`, `d-bar`, `d-a`, `d-cmd` sınıflarıyla çizilir; dar ekran için `only-narrow` sürümü olur.

- [ ] **Step 4: Metinleri iki dilde yaz** (`en.json`, `tr.json`); eski kullanılmayan anahtarları sil.

- [ ] **Step 5: CSS.** `.feature`: 2 sütunlu grid, 900px altında tek sütun; `.feature pre { overflow-x: auto; max-width: 100%; }`; `.plugin-grid` (gruplar, `a` chip'leri, mevcut `--accent` tonları); `.scope` (3 sütun → dar ekranda 1). Renkler mevcut token'lardan; koyu tema mevcut `@media (prefers-color-scheme: dark)` bloklarıyla uyumlu.

- [ ] **Step 6:** `go test ./...` → PASS. `go run . -collage-build -out dist -clean` → htmlcheck hatasız.

- [ ] **Step 7: Görsel kontrol.** `collage dev` ile `/en/` ve `/tr/` sayfalarını 1280px ve 375px genişlikte, açık ve koyu temada aç; yatay kaydırma yok, şemalar okunur, kod renkli.

- [ ] **Step 8:** `turkish-reviewer` agent'ını `locales/tr.json` (home) için çalıştır, düzeltmelerinden sonra `go test ./...`.

- [ ] **Step 9: Commit** — `site: home page tells why collage, in the order of the positioning spec`

---

### Task 2: Giriş sayfası (`introduction`)

**Files:**
- Modify: `content/introduction.md`, `content/tr/introduction.md`
- Modify: `static/app.css` (`.prose figure.diagram-figure` kuralları: kenarlık, padding, `figcaption` stili; mevcut `.diagram` ve `d-*` kuralları zaten global)
- Test: `main_test.go`

**Interfaces:**
- Consumes: Task 1'in `d-*` sınıfları ve `.diagram`.
- Produces: `figure.diagram-figure` sınıfı; Aşama 3 ve 4'teki sayfa şemaları bunu kullanır.

Başlık yapısı (EN; TR aynı seviyelerle):

```
# Introduction
## Why collage
## Fragments, not components
## A page knows what it is made of
## Middleware and plugins, as many as you need
## Still Go
## Sites and web apps
### What it leaves out
### When it is not the right tool
## How a request flows
## What you get
## Where to go next
```

İçerik:
- **Why collage:** Furkan'ın felsefesi. Az etkileşimli siteler; sunucudaki JS runtime'ının RAM ve disk maliyeti (rakamsız); HTML işin çoğunu görür, vanilla JS kalanını; fetch/XHR, SSE ve WebSocket var.
- **Fragments, not components:** mevcut `author`/`post`/`page` Go örneği aynen kalır, mevcut üç madde (concurrent fetch, page bilir, parça düşebilir) kalır. Üstüne `WithFragmentPath` + `fetch` örneği ve `/docs/forms-and-actions#a-fragment-at-its-own-url` linki eklenir.
- **A page knows what it is made of:** tag, `InvalidateTags` (stale + arka planda yeniden render) ve `PurgeTags` farkı; `/docs/caching#invalidating` linki.
- **Middleware and plugins:** tek middleware dosyasıyla isimsiz karşılaştırma, `app.Use` örneği, plugin kategorileri; `/docs/plugins` ve `/docs/middleware-and-apis` linkleri.
- **Still Go:** `app.Handle` örneği, data katmanı, tek binary; `/docs/middleware-and-apis#your-own-handler-apphandle`.
- **Sites and web apps:** çok kullanıcılı, rol bazlı web uygulaması: `elagoht/session` + `WithGuard` (rol kontrolü) + CSRF + `elagoht/honeypot` + `elagoht/live`. Kısa bir guard örneği (`WithGuard(requireRole("admin"))`, imza `pages-and-layouts#private-pages-guards`'taki `GuardFunc`'tan alınır).
  - **What it leaves out:** ORM, migration ve veritabanı katmanı yok; "session store yok" ifadesi kaldırılır, yerine "session çekirdekte değil, `elagoht/session` ile" gelir.
  - **When it is not the right tool:** state'i tamamen tarayıcıda yaşayan ekranlar (editör, canvas, çevrimdışı uygulama). Mevcut "client framework + `app.Handle`" önerisi kalır.
- **How a request flows:** mimari SVG. Tarayıcı → (plugin middleware'leri → `app.Use`) → router → sayfa / fragment path / action / `app.Handle` → render + cache ↔ senin data katmanın (DB, API, CMS). `role="img"`, `aria-labelledby`, `figcaption`, `data-search-ignore`. TR kopyada etiketler Türkçe.
- **What you get:** mevcut listedeki her madde korunur (strategies, data cache, forms + CSRF, links by name, head elements, non-HTML routes + `app.Handle`, static export, dev server ayrıntıları).
- `description` front matter'ı yeni konumlandırmaya göre güncellenir; `reference` alanına `PageBuilder.WithFragmentPath`, `App.InvalidateTags`, `App.PurgeTags`, `App.Use`, `App.Handle`, `FragmentBuilder.WithGuard` eklenir (test bunların pakette var olduğunu zaten kontrol ediyor).

- [ ] **Step 1: Failing test yaz** — `TestIntroDiagramAccessible`: `/en/docs/introduction/` ve `/tr/docs/introduction/` gövdesinde `<figure class="diagram-figure"`, `role="img"`, `aria-labelledby=` ve `<figcaption` var; `fill="#` ve `stroke="#` yok. `TestIntroPositioning`: EN gövdesinde `WithFragmentPath`, `PurgeTags`, `app.Handle` var; `session store` yok.
- [ ] **Step 2:** `go test ./ -run TestIntro` → FAIL.
- [ ] **Step 3:** `content/introduction.md` yaz.
- [ ] **Step 4:** `content/tr/introduction.md` yaz (aynı başlık seviyeleri, Türkçe SVG etiketleri).
- [ ] **Step 5:** CSS: `.prose figure.diagram-figure`.
- [ ] **Step 6:** `go run ./tools/checkcontent` ve `go test ./...` → PASS; export htmlcheck hatasız.
- [ ] **Step 7:** Görsel kontrol: introduction, 1280px ve 375px, açık ve koyu tema.
- [ ] **Step 8:** `turkish-reviewer` agent'ını `introduction` için çalıştır.
- [ ] **Step 9: Commit** — `docs: introduction tells why collage and what it is not`
