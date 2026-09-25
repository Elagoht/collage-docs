---
description: Uygulamalı bir rehber. Bir layout, tipli bir data handler ve bir template ile bir tarif page'i kurarsınız, ardından bir slot'a ikinci bir fragment eklersiniz.
reference: New, NewPage, NewFragment, Data, DataHandler, Load, ErrNotFound
---

# İlk page'iniz

Bu rehberde sıfırdan açılmış bir projede küçük bir tarif sitesi kurarsınız. Önce
`/recipes/{slug}` adresinde bir tarifi yükleyen ve onu sitenin layout'u içinde
render eden bir page yazarsınız. Ardından ikinci bir fragment eklersiniz: diğer
tariflerin listesi. Bu fragment, ilk fragment'in bir slot'una yerleşir. Rehber
yaklaşık on beş dakika sürer. İçindeki her parçayı, yazacağınız her page'de
yeniden kullanacaksınız.

Go ve `collage` CLI kurulu olmalı; bkz. [Kurulum](/docs/installation).

## Bir proje başlatın

```sh
collage new cookbook --template minimal
cd cookbook
go mod tidy
collage dev
```

[http://localhost:3000](http://localhost:3000) adresini açın ve `collage dev`'i
çalışır hâlde bırakın. Bundan sonra kaydettiğiniz her Go dosyası yeniden build
edilir ve tarayıcı kendini yeniler. Kaydettiğiniz her template ise rebuild
gerekmeden bir sonraki render'da görünür.

## Layout'a bakın

Minimal bir projede hazır bir layout vardır ve her page onu kullanır. Layout iki
dosyadan oluşur. Fragment'i `fragments/layouts/main.go` dosyasındadır:

```go
func Layout() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
			rc.HoistTitle("cookbook")
			return nil
		})).
		WithSlot("content", true, false).
		Build()
}
```

Template'i ise `templates/layouts/default.html` dosyasındadır:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  {{hoist "head"}}
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
</head>
<body>
  {{slot "content"}}
</body>
</html>
```

`WithSlot("content", true, false)`, `content` adında bir slot tanımlar. Bu slot
zorunludur ve tek bir fragment alır. `{{slot "content"}}` ise slot'un render
edildiği yerdir. Bu slot'u hiçbir zaman kendiniz doldurmazsınız. Bir page register
edildiğinde collage, page'in content fragment'ini bu slot'a yerleştirir. Tek bir
layout'un bütün page'ler tarafından paylaşılabilmesi bu sayede olur.

Template'te `<title>` yok. Başlığı layout'un handler'ı `rc.HoistTitle` ile
tanımlar; `{{hoist "head"}}` de başlığın yerleştiği yerdir. Kendi başlığını
tanımlayan bir page ikinci bir başlık eklemez, sitenin başlığının yerini alır.
Birazdan tarif page'i de bunu yapacak. Bkz.
[Head ve SEO](/docs/head-and-seo#keys-and-the-innermost-wins).

## Home page'e bakın

`pages/home.go` dosyasındaki home page, bir template'e veri vermenin en kısa yolunu
gösterir:

```go
// homeView is what templates/pages/home.html renders with, as ".".
type homeView struct {
	Name string
}

func HomePage() *collage.Page {
	content := collage.NewFragment("home-content", "pages/home.html").
		WithDataHandler(collage.Data(homeView{Name: "cookbook"})).
		Build()

	return collage.NewPage("home").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/").
		Static().
		Build()
}
```

`collage.Data`, template'e her render'da aynı değeri verir. `templates/pages/home.html`
içinde bu değer `.` olarak kullanılır:

```html
<h1>Hello from {{.Name}}</h1>
```

`homeView`'a bir alan ekleyin, değerini `collage.Data(...)` içinde verin ve
template'te kullanın; page bu alanı gösterir. Sabit bir veri için, örneğin bir link
listesi ya da bir başlık için, gereken tek şey budur. Tarif ise sabit değildir. URL'ye
bağlıdır ve bir yerden gelir. Bunun için bir fonksiyon gerekir.

## İçerik nereden geliyor

Gerçek bir site tarifleri bir veritabanından ya da bir CMS'ten yükler. Burada bir
map yeterli. `recipes/recipes.go` dosyasını oluşturun:

```go
// Package recipes is where this site's content comes from.
package recipes

import (
	"context"
	"fmt"
	"sort"

	"github.com/Elagoht/collage/pkg/collage"
)

type Recipe struct {
	Slug        string
	Title       string
	Minutes     int
	Ingredients []string
}

var all = map[string]Recipe{
	"pancakes":  {Slug: "pancakes", Title: "Pancakes", Minutes: 20, Ingredients: []string{"flour", "milk", "eggs", "butter"}},
	"omelette":  {Slug: "omelette", Title: "Omelette", Minutes: 10, Ingredients: []string{"eggs", "butter", "salt"}},
	"flatbread": {Slug: "flatbread", Title: "Flatbread", Minutes: 30, Ingredients: []string{"flour", "water", "salt", "olive oil"}},
}

// Get returns the recipe with slug, or an error wrapping collage.ErrNotFound.
func Get(ctx context.Context, slug string) (Recipe, error) {
	if err := ctx.Err(); err != nil {
		return Recipe{}, err
	}
	recipe, ok := all[slug]
	if !ok {
		return Recipe{}, fmt.Errorf("recipes: no recipe %q: %w", slug, collage.ErrNotFound)
	}
	return recipe, nil
}

// List returns every recipe, by title.
func List(ctx context.Context) ([]Recipe, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	list := make([]Recipe, 0, len(all))
	for _, recipe := range all {
		list = append(list, recipe)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Title < list[j].Title })
	return list, nil
}
```

Buradaki iki ayrıntı ileride önem kazanacak. Birincisi, eksik bir tarif
`collage.ErrNotFound`'un `%w` ile sarılmasıyla bildirilir. Framework "bu yok" (404)
ile "bu bozuldu" (500) durumlarını bu şekilde ayırt eder. İkincisi, iki fonksiyon da
bir `context.Context` alır, çünkü bir data handler'ın süresi context'i üzerinden
sınırlanır.

## Content fragment'i

Fragment, bir template'ten ve isteğe bağlı olarak o template'in render ettiği veriyi
çeken bir fonksiyondan oluşur. `pages/recipe.go` dosyasını oluşturun:

```go
package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/fragments/layouts"
	"cookbook/recipes"
)

// RecipePage is one recipe, at /recipes/{slug}.
func RecipePage() *collage.Page {
	content := collage.NewFragment("recipe-content", "pages/recipe.html").
		WithDataHandler(collage.DataHandler(loadRecipe)).
		Required().
		Build()

	return collage.NewPage("recipe").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/recipes/{slug}").
		Build()
}

// loadRecipe is the content fragment's data handler.
func loadRecipe(ctx context.Context, rc *collage.RenderContext) (recipes.Recipe, []string, error) {
	recipe, err := recipes.Get(ctx, rc.Param("slug"))
	if err != nil {
		return recipes.Recipe{}, nil, err
	}
	rc.HoistTitle(recipe.Title + " — cookbook")
	return recipe, []string{"recipe:" + recipe.Slug}, nil
}
```

Kodu satır satır inceleyelim.

- **`NewFragment("recipe-content", "pages/recipe.html")`** fragment'e ve template'ine
  isim verir. Template yolu `templates/` dizinine göredir ve uzantıyı da içerir.
- **`collage.DataHandler(loadRecipe)`**, kendi tipinize göre yazılmış bir handler'ı
  uyarlar. `loadRecipe` bir `recipes.Recipe` döndürdüğü için template de bir
  `recipes.Recipe` alır. Böylece kodunuzun hiçbir yerinde tipsiz değerlerle
  uğraşmanız gerekmez.
- **Handler üç şey döndürür**: veri, verinin oluşturulduğu dependency tag'leri ve bir
  hata. `recipe:pancakes` tag'i "bu page pancakes tarifini gösteriyor" anlamına
  gelir. O tarif değiştiğinde cache'teki kopyanın atılabilmesini sağlayan budur.
  Bildirecek tag'i olmayan bir handler, `collage.Load` ile tag döndürmekten
  kurtulabilir; bkz. [Data handler'lar](/docs/data-handlers#shorter-adapters-data-and-load).
- **`rc.Param("slug")`**, URL'de eşleşen `{slug}` değeridir.
- **`rc.HoistTitle`** page'e kendi `<title>`'ını verir. Content fragment'i layout'un
  içinde yer alır ve en içteki tanım kazanır. Bu yüzden bu başlık, layout'un
  `cookbook` başlığının yerini alır.
- **`Required()`**, page'in bu fragment olmadan var olamayacağını belirtir. Bu
  fragment başarısız olursa page de başarısız olur. Fragment `ErrNotFound`
  döndürürse page 404 olur.
- **`NewPage("recipe")`** page'e isim verir. Link'ler, testler ve plugin'ler page'e
  bu isimle başvurur. Bu yüzden isim, URL değiştiğinde değişmemelidir.
- **`WithPath("en", "/recipes/{slug}")`**, sitenin varsayılan locale'indeki URL'dir.

Strateji metodu çağrılmamış bir page `Dynamic()` olur: her request'te render edilir
ve hiç cache'lenmez. Page'i geliştirirken doğru varsayılan budur. Page çalışır hâle
geldiğinde `Incremental(10 * time.Minute)` ya da `Static()` onu cache'ler; bkz.
[Caching](/docs/caching).

## Template

`templates/pages/recipe.html` dosyasını oluşturun:

```html
<main class="recipe">
  <h1>{{.Title}}</h1>
  <p>Ready in {{.Minutes}} minutes.</p>

  <h2>Ingredients</h2>
  <ul>
    {{range .Ingredients}}<li>{{.}}</li>{{end}}
  </ul>
</main>
```

`.`, handler'ın döndürdüğü `recipes.Recipe` değeridir. Bu, Go'nun `html/template`
paketidir. Bu yüzden her değer, göründüğü yere uygun şekilde escape edilir. Başlığı
`<script>` olan bir tarif çalıştırılmaz, metin olarak yazdırılır. Bkz.
[Template'ler](/docs/templates).

## Page'i register edin

Uygulama bir page'den haberdar olmadıkça o page hiçbir işe yaramaz. `routes.go`
dosyasını açın ve listeye `pages.RecipePage()` ekleyin:

```go
for _, page := range []*collage.Page{pages.HomePage(), pages.RecipePage()} {
	if err := app.RegisterPage(page); err != nil {
		return fmt.Errorf("register page %q: %w", page.Name, err)
	}
}
```

Hatalar register sırasında yakalanır. Yazım hatası içeren bir template yolu, içi boş
kalmış zorunlu bir slot, aynı isme sahip iki page ya da hatalı biçimlendirilmiş bir
path bunlara örnektir. Bunların her biri, ilk ziyaretçinin hatayla karşılaşmasını
beklemeden programı başlangıçta durdurur. Hata mesajında ilgili page'in adı yer
alır.

## Sonucu görün

Kaydedin ve terminali izleyin: `collage dev` projeyi yeniden build eder ve yeniden
başlatır. Ardından
[localhost:3000/recipes/pancakes](http://localhost:3000/recipes/pancakes)
adresini açın.

Gördüğünüz page, `content` slot'una sizin fragment'iniz yerleştirilmiş layout'tur.
`templates/pages/recipe.html` dosyasını düzenleyin, örneğin bir cümle ekleyin ya da
bir başlığı değiştirin. Tarayıcı değişiklikle birlikte yenilenir. Bu sırada rebuild
yapılmadı, çünkü development'ta template'ler her request'te diskten okunur.

Şimdi [/recipes/lasagne](http://localhost:3000/recipes/lasagne) adresini deneyin.
`recipes.Get`, `collage.ErrNotFound`'u sardı ve fragment `Required()` olduğu için
response 404 olur. Proje kendi not-found page'ini register etmediği için response
body'si, collage'ın built-in ve sade not-found page'idir. `app.RegisterNotFoundPage`
ile siteye kendi page'inizi verebilirsiniz; bkz.
[Page'ler ve layout'lar](/docs/pages-and-layouts#not-found-and-error-pages).

Bir hatanın nasıl göründüğünü görmek için bilerek bir şeyi bozun. Template'te
`{{.Title}}` ifadesini `{{.Name}}` olarak değiştirin, kaydedin ve tarayıcıyı
yenileyin. Böyle bir alan olmadığı için zorunlu fragment başarısız olur ve page 500
döner. Development'ta error page, hatanın başladığı fragment olarak
`recipe-content`'i gösterir. Ayrıca hata zincirinin tamamını, `pages/recipe.html:2:8`
konumuna ve bulunamayan alana kadar yazdırır. Sonra değişikliği geri alın.

## Bir slot'a ikinci bir fragment ekleyin

Bir page nadiren tek parçadan oluşur. Diğer tariflerin listesini, kendi verisi olan
ayrı bir fragment olarak ekleyin. `pages/more.go` dosyasını oluşturun:

```go
package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/recipes"
)

// moreView is what the "more recipes" fragment renders.
type moreView struct {
	Recipes []recipes.Recipe
}

// MoreRecipes lists every recipe but the one on the page.
func MoreRecipes() *collage.Fragment {
	return collage.NewFragment("more-recipes", "fragments/more-recipes.html").
		WithDataHandler(collage.DataHandler(loadMore)).
		Build()
}

func loadMore(ctx context.Context, rc *collage.RenderContext) (moreView, []string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return moreView{}, nil, err
	}
	var view moreView
	for _, recipe := range list {
		if recipe.Slug != rc.Param("slug") {
			view.Recipes = append(view.Recipes, recipe)
		}
	}
	return view, []string{"recipes"}, nil
}
```

Bu fragment'in template'i `templates/fragments/more-recipes.html` dosyasıdır:

```html
<aside class="more-recipes">
  <h2>More recipes</h2>
  <ul>
    {{range .Recipes}}
      <li><a href="{{pageURL "recipe" "slug" .Slug}}">{{.Title}}</a></li>
    {{end}}
  </ul>
</aside>
```

`pageURL "recipe" "slug" .Slug`, link'i page'in path'inden değil, adından oluşturur.
Böylece `/recipes/{slug}` bir gün `/r/{slug}` olursa link de page'i takip eder. Bkz.
[Link'ler ve locale'ler](/docs/links-and-locales).

Şimdi tarif fragment'ine bir slot verin ve listeyi bu slot'a yerleştirin.
`pages/recipe.go` dosyasında:

```go
content := collage.NewFragment("recipe-content", "pages/recipe.html").
	WithDataHandler(collage.DataHandler(loadRecipe)).
	WithSlot("more", false, false).
	WithSlotFragment("more", MoreRecipes()).
	Required().
	Build()
```

Ardından slot'u ait olduğu yerde render edin. Bunun için
`templates/pages/recipe.html` dosyasında, `</main>` satırından önce şunu ekleyin:

```html
  {{slot "more"}}
```

İki dosyayı da kaydedin. Tarif page'inde artık diğer iki tarifin listesi var ve her
biri kendi page'ine link veriyor.

Bu page hakkında, hiçbir yerde açıkça yazılmamış üç şey geçerlidir.

- **Liste isteğe bağlıdır.** `WithSlot("more", false, false)` bu slot'u zorunlu
  olmayan bir slot olarak tanımladı. `MoreRecipes` de `Required()` değil. `loadMore`
  başarısız olursa page listesiz sunulur. Development'ta ayrıca hangi
  fragment'in neden başarısız olduğunu gösteren bir panel görünür. Bozuk bir sidebar,
  500 hatası değil, yalnızca eksik bir sidebar olur.
- **Page'in tag'leri, iki fragment'in tag'lerinin birleşimidir.** Page artık hem
  `recipe:pancakes`'e hem de `recipes`'e bağlıdır. Page cache'lendikten sonra
  `recipes`'i invalidate etmek bütün tarif page'lerini cache'ten düşürür. Yeni bir
  tarif eklendiğinde olması gereken de budur. `recipe:pancakes`'i invalidate etmek
  ise yalnızca bu page'i düşürür.
- **İki handler birbirini gerektiğinden fazla beklemedi.** Bir fragment'in
  handler'ı, üst fragment'inin handler'ı döndükten sonra başlar. Aynı fragment'in
  slot'larındaki kardeş fragment'lerin handler'ları ise birlikte başlar. Tarif
  fragment'ine, içinde yavaş bir fragment bulunan ikinci bir slot verirseniz page,
  sürelerin toplamını değil, en yavaş olanı bekler. Bkz.
  [Data handler'lar](/docs/data-handlers#when-handlers-run).

## Test edin

Bir collage uygulaması sunucu çalıştırmadan test edilir. `app.Handler()` sıradan bir
`http.Handler`'dır ve onu `net/http/httptest` ile çalıştırırsınız. `main_test.go`
dosyasını oluşturun. İçine, uygulamayı `main`'in kullandığı `newApp` fonksiyonuyla
kuran bir yardımcı fonksiyon ve bir test yazın:

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	cacheDir = t.TempDir() // a disk cache of its own, not the last run's
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestRecipePage(t *testing.T) {
	rec := get(t, "/recipes/pancakes")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /recipes/pancakes = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `href="/recipes/omelette"`) {
		t.Errorf("the page does not link to the other recipes:\n%s", rec.Body.String())
	}

	if rec := get(t, "/recipes/lasagne"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /recipes/lasagne = %d, want 404", rec.Code)
	}
}
```

```sh
go test ./...
```

## Sırada ne var

- [Page'ler ve layout'lar](/docs/pages-and-layouts) — path'ler, stratejiler, error
  page'leri ve register işleminin bir page'e ne yaptığı.
- [Fragment'ler ve slot'lar](/docs/fragments-and-slots) — slot'lar, hata politikası
  ve içerikten doldurulan slot'lar.
- [Data handler'lar](/docs/data-handlers) — handler sözleşmesi, fragment'ler arasında
  veri paylaşımı ve timeout'lar.
- [Caching](/docs/caching) — bu page'i `Dynamic()` olmaktan çıkarıp bir kez render
  edilen ve bir tarif değiştiğinde atılan bir page'e dönüştürmek.
