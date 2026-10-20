---
description: Plugin'siz, uygulamalı bir rehber: öğeleri kendi URL'sinde bir fragment olan, form ve birkaç satır JavaScript ile yerinde güncellenen bir alışveriş listesi.
reference: PageBuilder.WithFragmentPath, PageBuilder.WithAction, PageBuilder.Incremental, RenderFragment, SeeOther, FetchHeader, ActionResult, LoadTagged, CacheStatusHeader
---

# Page'in bir kısmını güncellemek

Bu rehberde bir alışveriş listesi kurarsınız. Liste, bulunduğu page'in bir
kısmıdır. Bir öğe eklendiğinde yalnızca o kısım değiştirilir, page'in geri kalanına
dokunulmaz. Başka bir sekmede açık olan liste de yeni öğeyi birkaç saniye sonra
gösterir. Rehber ne bir plugin ne de client tarafında bir framework kullanır.
Gerekenler kendi URL'sinde açılmış bir fragment, bir action ve yaklaşık yirmi satır
sade JavaScript'tir. JavaScript kapalıyken de aynı page sıradan bir form olarak
çalışır.

Bitmiş proje dokümantasyonun repository'sinde,
[`examples/shopping-list`](https://github.com/Elagoht/collage-docs/tree/main/examples/shopping-list)
dizinindedir ve aşağıdaki her dosya o projenin dosyasının kendisidir.
[İlk page'iniz](/docs/your-first-page) rehberini okuduğunuz varsayılır.

## Ne kuracaksınız

- `/list`: listeyi ve ona ekleme yapan bir formu gösteren, liste değişene kadar
  cache'lenen bir page.
- `/list/items`: tek başına liste. Bu, page'in render ettiği fragment'in kendi
  URL'sinde açılmış halidir.
- Bir öğe ekleyen `POST /list`. Formu gönderen bir tarayıcı `/list`'e redirect
  edilir. Fragment'i isteyen bir script'e ise liste fragment'iyle cevap verilir.
- Formu `fetch` ile gönderen, listeyi yerine koyan ve listeyi her beş saniyede bir
  yeniden çeken bir script.

## Bir proje başlatın

```sh
collage new shopping-list
cd shopping-list
go mod tidy
collage dev
```

`collage new` layout'u, bir home page'i ve `main.go`'yu yazar. Bu rehber liste
için bir package, iki fragment, bir page, iki template ve bir script ekler. Page'i
de `routes.go`'da register eder.

## Liste

Liste, verinizin gerçekte durduğu yerin (bir veritabanı, bir API) yerini tutar.
Page'ler onu okur, action ona yazar. Projede başka hiçbir şey listenin nasıl
saklandığını bilmez. `list/list.go`'yu oluşturun:

<!-- file: list/list.go -->

```go
// Package list keeps the shopping list. It stands in for a database: the pages
// read it and the action writes to it, and none of them know how it is stored.
package list

import (
	"strings"
	"sync"
)

// Tag is the dependency tag of everything rendered from the list. Invalidating
// it renders those pages again.
const Tag = "list"

// Item is one thing to buy.
type Item struct {
	Name string
}

// List is a shopping list several readers can add to at once.
type List struct {
	mu    sync.Mutex
	items []Item
}

// Shopping is the list the site shows.
var Shopping = New("Milk", "Bread")

// Returns a list holding names, in order.
func New(names ...string) *List {
	l := &List{}
	for _, name := range names {
		l.Add(name)
	}
	return l
}

// Returns the items, in the order they were added. The slice is a copy, so a
// template can range over it while someone else adds to the list.
func (l *List) Items() []Item {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Item(nil), l.items...)
}

// Adds name to the end of the list, and reports whether it did: a name that is
// blank once trimmed is not added.
func (l *List) Add(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, Item{Name: name})
	return true
}
```

Öğeleri tutmanın ötesinde iki şey önemlidir. `Tag`, listeden render edilen her
şeyin dependency tag'idir. Listeyi okuyan fragment bu tag'i belirtir, listeyi
değiştiren action da onu invalidate eder. Ayrıca listeye aynı anda birkaç
request'ten öğe eklemek güvenlidir. Bu gereklidir, çünkü birkaç okuyucu aynı anda
ekleme yapacaktır.

## Items fragment'i

Bir fragment, bir template ve onun ihtiyaç duyduğunu çeken fonksiyondur.
`fragments/pages/list.go`'yu scaffold'daki `home.go`'nun yanına oluşturun:

<!-- file: fragments/pages/list.go -->

```go
package fragments

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/list"
)

// Returns the list's items: templates/fragments/items.html. It reads the list on
// every render and tags what it rendered, so invalidating list.Tag renders it
// again.
func Items() *collage.Fragment {
	return collage.NewFragment("items", "fragments/items.html").
		WithData(collage.LoadTagged(func(context.Context, *collage.RenderContext) ([]list.Item, []string, error) {
			return list.Shopping.Items(), []string{list.Tag}, nil
		})).
		Build()
}

// Returns the list page's content: templates/pages/list.html, with items in its
// "items" slot.
func List(items *collage.Fragment) *collage.Fragment {
	return collage.NewFragment("list", "pages/list.html").
		WithTitle("Shopping list").
		WithSlotFragment("items", items).
		Build()
}
```

`Items`, her render'da öğeleri ve geldikleri tag'i döndürür. `List`, page'in
içeriğidir ve `Items` onun bir slot'undadır. Fragment `List`'in içinde
oluşturulmaz, dışarıdan verilir. Page aynı fragment'i iki yerde kullanacaktır:
slot'ta ve kendi URL'sinde. İkisinin tek bir fragment olması gerekir.

Listenin template'i, `<ul>` dahil fragment'in bütün markup'ıdır:

<!-- file: templates/fragments/items.html -->

```html
<ul id="items" data-src="{{fragmentURL "list" "items"}}">
  {{range .}}<li>{{.Name}}</li>
  {{else}}<li>Nothing to buy.</li>
  {{end}}
</ul>
```

`id`, script'in yerine koyduğu öğeyi gösterir. `data-src` ise script'in listeyi
nereden çektiğini söyler. `fragmentURL` bu adresi page'in ve fragment'in adından
oluşturur. Böylece page'in path'i değişirse adres de onunla birlikte değişir.
`{{.Name}}`, bir template'in yazdığı her şey gibi escape edilir: `<b>Salt</b>` adlı
bir öğe kalın yazı olarak değil, bu karakterlerin kendisi olarak görünür.

## Page

Page'in template'i listeyi slot'una, altına da bir form koyar:

<!-- file: templates/pages/list.html -->

```html
<main>
  <h1>Shopping list</h1>

  {{slot "items"}}

  <form id="add" method="post" action="{{pageURL "list"}}">
    {{csrfToken}}
    <label>Item <input name="name" required autocomplete="off"></label>
    <button>Add</button>
  </form>

  <script src="{{asset "/static/list.js"}}" defer></script>
</main>
```

`{{csrfToken}}`, forgery token'ını taşıyan hidden input'u render eder. Bu input
olmadan gönderilen bir form 403 ile reddedilir. Formu gönderen bir script de onu
diğer alanlarla birlikte gönderir. Form, action adresini `pageURL` ile belirtir.
Böylece page'in kendi URL'sine post edilir. Script ise `static/`'ten gelir. Orada
`{{asset}}` ona içeriğiyle birlikte değişen bir adres verir.

## Fragment'i kendi URL'sinde açın

Şimdi page'in kendisi. `pages/list.go`'yu oluşturun:

<!-- file: pages/list.go -->

```go
package pages

import (
	"context"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/fragments/layouts"
	fragments "shopping-list/fragments/pages"
	"shopping-list/list"
)

// Returns the shopping list page at /list, cached for an hour or until the list
// changes.
//
// Its items fragment is also opened at /list/items, so a script can fetch that
// part alone. Posting the form adds an item: a script that asks for it is
// answered with the items fragment, a browser on its own with a redirect back.
func List() *collage.Page {
	items := fragments.Items()
	return collage.NewPage("list").
		WithLayouts(layouts.Master()).
		WithContent(fragments.List(items)).
		WithPath("en", "/list").
		WithFragmentPath("en", "/list/items", items).
		Incremental(time.Hour).
		WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			list.Shopping.Add(rc.Request.PostFormValue("name"))

			result := collage.SeeOther("/list")
			if rc.Request.Header.Get(collage.FetchHeader) != "" {
				result = collage.RenderFragment(items)
			}
			result.InvalidateTags = []string{list.Tag}
			return result, nil
		}).
		Build()
}
```

Bu rehberin asıl konusu `WithFragmentPath("en", "/list/items", items)` satırıdır.
Bu satır `items` fragment'ini `/list/items`'ta açar. Oraya yapılan bir GET, o
fragment'i, yani data handler'ını ve template'ini, etrafında layout olmadan render
eder. Açıkça tanımlanmayan hiçbir şeye bu yolla erişilemez. Bir fragment path, ait
olduğu page'in bir kısmıdır. Page'in guard'ları varsa fragment path de onlardan
geçer.

Page `Incremental(time.Hour)`'dır: bir kez render edilir, sonra bir saat boyunca
ya da oluşturulduğu bir tag invalidate edilene kadar cache'ten sunulur. Bu olmasa
data handler page'in her request'te render edilmesine yol açardı. Bu doğru olurdu,
ama tag'e yapacak bir iş bırakmazdı. Bir fragment path'in cevabı hiçbir zaman
cache'lenmez, bu yüzden `/list/items` her zaman listenin o anki halidir.

Page'i `routes.go`'da register edin:

<!-- file: routes.go -->

```go
package main

import (
	"shopping-list/pages"

	"github.com/Elagoht/collage/pkg/collage"
)

// Registers every page, document and action to app. A new route goes here.
func register(app *collage.App) error {
	return app.Register(
		pages.Home(),
		pages.List(),
	)
}
```

[http://localhost:6060/list](http://localhost:6060/list)'i, ardından
[/list/items](http://localhost:6060/list/items)'ı açın: aynı liste, bir kez
page'in içinde, bir kez de tek başına.

## Bir öğe ekleyin

Action, `WithAction`'a verilen fonksiyondur. Gönderilen adı okur, ekler ve nasıl
cevap vereceğine karar verir:

- Formu gönderen bir tarayıcı `SeeOther("/list")`, yani bir 303 alır. Böylece
  tarayıcı page'i bir GET ile yükler ve sayfayı yenilemek formu yeniden göndermez.
- `Collage-Fetch` header'ını (`collage.FetchHeader`) gönderen bir script
  `RenderFragment(items)` alır. Bu, az önce eklenen öğeyle render edilmiş listedir.
  Cevapta başka hiçbir şey yoktur.

Her iki durumda da sonuç `InvalidateTags: []string{list.Tag}` taşır. Cevap
yazılmadan önce listeden oluşturulan her cache'lenmiş page stale olarak işaretlenir.
Page arka planda ya da en geç `/list`'e yapılan bir sonraki GET'te yeniden render
edilir. Böylece o GET, hem bu sekmede hem de diğer bütün sekmelerde yeni öğeyi
gösterir.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 380" role="img" aria-labelledby="add-title">
<title id="add-title">Script varken tarayıcı formu Collage-Fetch header'ıyla gönderir. collage öğeyi ekler, list tag'ini invalidate eder ve items fragment'iyle cevap verir. Script de bu fragment'i listenin yerine koyar. Script yokken tarayıcı formu gönderir, /list'e redirect edilir ve page'i yeniden render edilmiş haliyle yükler. Script her beş saniyede bir /list/items'ı çeker.</title>
<defs><marker id="add-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="10" y="8" width="100" height="24" rx="12" /><text x="60" y="24" text-anchor="middle">tarayıcı</text>
<rect class="d-cmd" x="210" y="8" width="100" height="24" rx="12" /><text class="d-cmd-text" x="260" y="24" text-anchor="middle">collage</text>
<line class="d-rule" x1="60" y1="32" x2="60" y2="372" />
<line class="d-rule" x1="260" y1="32" x2="260" y2="372" />
<text class="d-small" x="160" y="50" text-anchor="middle">script varken</text>
<line class="d-flow" x1="60" y1="66" x2="254" y2="66" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="61" text-anchor="middle">POST /list · Collage-Fetch</text>
<rect class="d-soft-a" x="200" y="74" width="114" height="40" rx="5" /><text class="d-small" x="257" y="90" text-anchor="middle">öğeyi ekle</text><text class="d-small" x="257" y="104" text-anchor="middle">invalidate "list"</text>
<line class="d-flow" x1="260" y1="130" x2="66" y2="130" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="125" text-anchor="middle">200 · items fragment'i</text>
<rect class="d-b" x="8" y="138" width="118" height="24" rx="5" /><text class="d-on-b d-small" x="67" y="154" text-anchor="middle">#items'ı değiştir</text>
<line class="d-rule" x1="10" y1="180" x2="310" y2="180" />
<text class="d-small" x="160" y="196" text-anchor="middle">script yokken</text>
<line class="d-flow" x1="60" y1="212" x2="254" y2="212" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="207" text-anchor="middle">POST /list</text>
<line class="d-flow" x1="260" y1="236" x2="66" y2="236" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="231" text-anchor="middle">303 · Location: /list</text>
<line class="d-flow" x1="60" y1="260" x2="254" y2="260" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="255" text-anchor="middle">GET /list</text>
<line class="d-flow" x1="260" y1="284" x2="66" y2="284" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="279" text-anchor="middle">200 · yeniden render edilmiş page</text>
<line class="d-rule" x1="10" y1="302" x2="310" y2="302" />
<text class="d-small" x="160" y="318" text-anchor="middle">her beş saniyede bir</text>
<line class="d-flow" x1="60" y1="334" x2="254" y2="334" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="329" text-anchor="middle">GET /list/items</text>
<line class="d-flow" x1="260" y1="358" x2="66" y2="358" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="353" text-anchor="middle">200 · items fragment'i</text>
</svg>
<figcaption>Tek action, iki cevap. Script varken form yalnızca listeyle cevaplanır ve page olduğu yerde kalır. Script yokken form bir redirect ile cevaplanır. Redirect'in götürdüğü page de, action onun tag'ini invalidate ettiği için yeniden render edilir. Script ayrıca listeyi her beş saniyede bir tek başına ister.</figcaption>
</figure>

Adın başındaki ve sonundaki boşluklar kırpılır, boş kalan bir ad eklenmez. Cevap
her iki durumda da aynıdır, böylece bir script ikisini birbirinden ayırmak zorunda
kalmaz. Reddedilen bir gönderime hata
mesajıyla cevap vermek
[Başarısızlık render eder, başarı redirect eder](/docs/forms-and-actions#failure-renders-success-redirects)
bölümünün konusudur.

## Birkaç satır JavaScript ile yerine koyun

`static/list.js`'i oluşturun:

<!-- file: static/list.js -->

```js
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
```

Submit handler'ı tarayıcının kendi post'unu durdurur ve aynı post'u `fetch` ile
yapar. `new FormData(form)`, forgery token'ı da dahil formun bütün alanlarını
gönderir. Bu yüzden token için ayrıca bir şey eklemeniz gerekmez. Action'dan
fragment'i isteyen şey `Collage-Fetch` header'ıdır. Cevap, `<ul>` dahil listenin
tamamıdır. `outerHTML` onu `id` ve `data-src` ile birlikte eskisinin yerine koyar.
Böylece script bir sonraki değiştirmede listeyi aynı şekilde bulur. Cevap hiç
gelmezse (ağ bağlantısı yoksa) ya da gelen cevap liste değilse form sıradan yoldan
gönderilir ve tarayıcı sunucu ne döndürdüyse onu gösterir. Script'in ikinci yarısı
sunucunun cevap vermediği turu atlar, listenin olmadığı bir page'de de hiçbir şey
yapmaz.

### Aynısı elagoht/client ile

Yukarıdaki script mekanizmanın tamamıdır ve onu bir kez elle yazmaya değer. Bundan
sonra [elagoht/client](/docs/plugins#elagohtclient) plugin'i sizi `fetch` ve
`outerHTML` satırlarından kurtarabilir; bitmiş proje onu kullanmaz. Plugin'i
ekleyin ve script'ini master layout'un `<head>`'ine koyun:

```sh
collage plugin add elagoht/client
```

```html
<head>
  {{clientjs}}
</head>
```

Artık polling yarısı tek bir satırdır. `collage.replace("#items")`
element'in `data-src`'sini çeker ve cevabı, `outerHTML` satırının yaptığı gibi,
onun yerine koyar. `collage.morph("#items")` aynı request'i yapar, ama listeyi
yerinde patch'ler. Böylece okuyucunun page'de yaptığı şeye dokunulmaz:

```js
// Every five seconds, the list as it is now, patched into the page in place.
// collage.replace("#items") would put it in whole, as outerHTML did.
setInterval(() => collage.morph("#items").catch(() => {}), 5000);
```

Başarısız olan bir çağrı (cevap gelmemesi, tek bir element olmayan bir cevap,
liste üzerinde daha yeni bir çağrı) listeyi olduğu gibi bırakır ve reject olur.
`.catch(() => {})` da onu, elle yazılan sürümde olduğu gibi, atlanan bir tura
çevirir.

Submit yarısı kendi `fetch`'ini korur: form'u post eder, `replace` ve `morph` ise
`GET` ile çeker. Action'ın cevabı, elde olan markup'ı swap eden ve tek bir element
olmayan cevabı reddeden `collage.put` ile yerleşir:

```js
const form = collage.$("#add");

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    const response = await fetch(form.action, {
      method: "POST",
      headers: { "Collage-Fetch": "1" },
      body: new FormData(form),
    });
    if (!response.ok) throw new Error(`answered ${response.status}`);
    collage.put("#items", await response.text());
    form.reset();
  } catch {
    form.submit();
  }
});
```

## Listeyi güncel tutun

Script'in ikinci yarısı her beş saniyede bir `/list/items`'ı ister. Böyle bir
liste için "canlı" olmanın bütün yolu budur. Bir fragment path her request'te
yeniden render edilir, bu yüzden başka bir sekmede eklenen bir öğe bir sonraki
cevapta yer alır.
Page'i iki sekmede açın, birinde bir öğe ekleyin ve diğerini izleyin.

Polling, ara sıra değişen bir page için uygundur. Sık değişen ya da sunucuda bir
şey değiştiği anda değişmesi gereken bir page için
[elagoht/live](/docs/plugins#elagohtlive) plugin'i vardır. Bu plugin, template'te
işaretlenen fragment'leri bir server-sent event stream'i üzerinden yeniler ve
bunun için az önce tanımladığınız fragment path'leri kullanır.

## Çalıştırın

`collage dev` projeyi en başından beri sunuyor.
[http://localhost:6060/list](http://localhost:6060/list)'i açın. Bir öğe ekleyin,
sonra tarayıcıda JavaScript'i kapatıp bir tane daha ekleyin: page yeniden yüklenir
ve liste iki seferde de doğrudur.

Development modu cache'i okumaz. Page'in cache'ten sunulduğunu görmek için
`collage dev`'i durdurun ve projeyi deploy edildiği şekilde çalıştırın:

```sh
go run .
```

[http://localhost:6060/list](http://localhost:6060/list)'i tarayıcının network
paneli açıkken açın ve sayfayı yenileyin: page'in cevabı
`X-Collage-Cache: HIT` taşır, yani render edilmeden sunulmuştur. Bir öğe ekleyin,
geri döndüğünüz page onu gösterir.

## Test edin

Testler uygulamayı `collagetest` üzerinden, bir tarayıcı gibi kullanır.
Cookie'ler saklanır, form hidden token'ıyla gönderilir, script'in gönderdiği
request ise elle oluşturulur. `list_test.go`'yu oluşturun:

<!-- file: list_test.go -->

```go
package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"shopping-list/list"
)

// client builds the application main builds, on a fresh list and a disk cache of
// its own, and returns a browser for it.
func client(t *testing.T) *collagetest.Client {
	t.Helper()
	list.Shopping = list.New("Milk", "Bread")
	app, err := newApp(collage.Env{CacheDir: t.TempDir()})
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	// Shut down before the directory is removed: an invalidation renders the
	// pages it reached again in the background, into that directory.
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })
	return collagetest.New(t, app.Handler())
}

// fetchAdd posts name the way static/list.js does: the form's fields, its token
// among them, and the header that asks for the fragment instead of a redirect.
func fetchAdd(c *collagetest.Client, page *collagetest.Response, name string) *collagetest.Response {
	body := url.Values{"_csrf": {page.CSRFToken()}, "name": {name}}
	req := c.Request(http.MethodPost, "/list", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(collage.FetchHeader, "1")
	return c.Do(req)
}

func TestListPage(t *testing.T) {
	page := client(t).Get("/list").WantStatus(http.StatusOK)
	for _, want := range []string{"Milk", "Bread", `data-src="/list/items"`} {
		if !strings.Contains(page.Body, want) {
			t.Errorf("the page has no %s:\n%s", want, page.Body)
		}
	}
}

func TestItemsFragment(t *testing.T) {
	items := client(t).Get("/list/items").WantStatus(http.StatusOK)
	if strings.Contains(items.Body, "<html") || !strings.Contains(items.Body, `id="items"`) {
		t.Errorf("/list/items is not the items fragment alone:\n%s", items.Body)
	}
}

func TestAddWithForm(t *testing.T) {
	c := client(t)
	res := c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Eggs"}}).WantStatus(http.StatusSeeOther)
	if res.Location() != "/list" {
		t.Errorf("Location = %q, want /list", res.Location())
	}
	if page := c.Get("/list"); !strings.Contains(page.Body, "Eggs") {
		t.Errorf("the page does not show the item just added:\n%s", page.Body)
	}
}

func TestAddWithFetch(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "Butter").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<html") || !strings.Contains(res.Body, `id="items"`) || !strings.Contains(res.Body, "Butter") {
		t.Errorf("a fetch was not answered with the items fragment:\n%s", res.Body)
	}
}

func TestAddIgnoresBlank(t *testing.T) {
	c := client(t)
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"   "}}).WantStatus(http.StatusSeeOther)
	res := fetchAdd(c, c.Get("/list"), "").WantStatus(http.StatusOK)
	if n := strings.Count(res.Body, "<li"); n != 2 {
		t.Errorf("the list has %d items after blank names, want 2:\n%s", n, res.Body)
	}
}

func TestItemsFragmentIsFresh(t *testing.T) {
	c := client(t)
	c.Get("/list/items")
	fetchAdd(c, c.Get("/list"), "Tea").WantStatus(http.StatusOK)
	if items := c.Get("/list/items"); !strings.Contains(items.Body, "Tea") {
		t.Errorf("/list/items does not show the item just added:\n%s", items.Body)
	}
}

func TestAddWithoutTokenIsRefused(t *testing.T) {
	c := client(t)
	req := c.Request(http.MethodPost, "/list", strings.NewReader("name=Jam"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Do(req).WantStatus(http.StatusForbidden)
	if items := c.Get("/list/items"); strings.Contains(items.Body, "Jam") {
		t.Errorf("a post without a token added an item:\n%s", items.Body)
	}
}

func TestItemNameIsEscaped(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "<b>Salt</b>").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<b>") || !strings.Contains(res.Body, "&lt;b&gt;Salt") {
		t.Errorf("the item's name was not escaped:\n%s", res.Body)
	}
}

func TestPageIsCachedUntilTheListChanges(t *testing.T) {
	c := client(t)
	c.Get("/list").WantStatus(http.StatusOK)
	if page := c.Get("/list"); page.Header.Get(collage.CacheStatusHeader) != "HIT" {
		t.Errorf("the second request was not served from the cache: %s = %q", collage.CacheStatusHeader, page.Header.Get(collage.CacheStatusHeader))
	}
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Honey"}}).WantStatus(http.StatusSeeOther)
	if page := c.Get("/list"); !strings.Contains(page.Body, "Honey") {
		t.Errorf("the cached page still shows the old list:\n%s", page.Body)
	}
}
```

`client`, test bittiğinde uygulamayı kapatır. Bunun nedeni, invalidation'ın
ulaştığı page'leri arka planda yeniden render edip testin cache dizinine
yazmasıdır. Testleri çalıştırın:

```sh
go test ./...
```

## Sırada ne var

- [Form'lar ve action'lar](/docs/forms-and-actions): bir action'ın neyle cevap
  verebileceği, forgery koruması ve
  [kendi URL'sinde bir fragment](/docs/forms-and-actions#a-fragment-at-its-own-url)
  konusunun tamamı.
- [Caching](/docs/caching): stratejiler, tag'ler ve invalidate etmenin ne yaptığı.
- [Test yazmak](/docs/testing): `collagetest`'in geri kalanı.
- [Plugin kullanmak](/docs/plugins): polling yetmediğinde.
