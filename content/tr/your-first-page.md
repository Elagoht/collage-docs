---
description: Uygulamalı bir rehber. Bir layout, bir data handler ve bir template ile bir tarif page'i kurarsınız, bir slot'a ikinci bir fragment eklersiniz ve bütün tarifleri static dosyalara export edersiniz.
reference: New, NewPage, NewFragment, FragmentBuilder.WithData, FragmentBuilder.WithTitle, PageBuilder.WithStaticParams, DataHandler, Load, ErrNotFound, ErrUnknownSlot
---

# İlk page'iniz

Bu rehberde sıfırdan açılmış bir projede küçük bir tarif sitesi kurarsınız. Önce
`/recipes/{slug}` adresinde bir tarifi yükleyen ve onu sitenin layout'u içinde
render eden bir page yazarsınız. Ardından ikinci bir fragment eklersiniz: diğer
tariflerin listesi. Bu fragment o page'in bir slot'una yerleşir. Son olarak siteyi,
her tarif için bir dosya olacak şekilde static olarak export edersiniz. Rehber
yaklaşık on beş dakika sürer. İçindeki her parçayı, yazacağınız her page'de yeniden
kullanacaksınız.

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
		WithTitle("cookbook").
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

`{{slot "content"}}`, `content` adında bir slot'tur ve render edildiği yerdir. Go
tarafında hiçbir şey tanımlanmaz: bir slot için template'in onu çağırması yeterlidir.
Bu slot'u hiçbir zaman kendiniz de doldurmazsınız. Bir page register edildiğinde
collage, page'in content fragment'ini bu slot'a yerleştirir. Tek bir layout'un
bütün page'ler tarafından paylaşılabilmesi bu sayede olur.

Template'te `<title>` yok. Başlığı layout `WithTitle` ile tanımlar.
`{{hoist "head"}}` de başlığın yerleştiği yerdir. Kendi başlığını tanımlayan bir
page ikinci bir başlık eklemez, sitenin başlığının yerini alır.
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
		WithData(homeView{Name: "cookbook"}).
		Build()

	// No Static() needed: nothing here fetches per render, so the page is static.
	return collage.NewPage("home").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/").
		Build()
}
```

`WithData`, template'e her render'da aynı değeri verir. `templates/pages/home.html`
içinde bu değer `.` olarak kullanılır:

```html
<h1>Hello from {{.Name}}</h1>
```

`homeView`'a bir alan ekleyin, değerini `WithData(...)` içinde verin ve template'te
kullanın; page bu alanı gösterir. Sabit bir veri için, örneğin bir link listesi ya
da bir başlık için, gereken tek şey budur.

Page hiçbir strateji tanımlamaz, yine de yorumda yazdığı gibi static'tir. Strateji
tanımlamayan bir page, render ettiği hiçbir şey render başına veri çekmiyorsa
static olur. Sabit veri ve sabit bir title hiçbir şey çekmez. Bu yüzden home page
bir kez render edilir, cache açıkken saklanır ve `collage export` onu bir dosyaya
yazar.

Tarif ise sabit değildir. URL'ye bağlıdır ve bir yerden gelir. Bunun için bir
fonksiyon gerekir.

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
		WithDataHandler(loadRecipe).
		Required().
		Build()

	return collage.NewPage("recipe").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/recipes/{slug}").
		Build()
}

// loadRecipe is the content fragment's data handler.
func loadRecipe(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	recipe, err := recipes.Get(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	rc.HoistTitle(recipe.Title + " — cookbook")
	return recipe, []string{"recipe:" + recipe.Slug}, nil
}
```

Kodu satır satır inceleyelim.

- **`NewFragment("recipe-content", "pages/recipe.html")`** fragment'e ve template'ine
  isim verir. Template yolu `templates/` dizinine göredir ve uzantıyı da içerir.
- **`WithDataHandler(loadRecipe)`** fragment'e data handler'ını verir. Bu,
  `WithDataHandler`'ın beklediği biçimde bir fonksiyondur, arada adapter yoktur.
- **Handler üç şey döndürür**: veri, verinin oluşturulduğu dependency tag'leri ve bir
  hata. Veri, `any` olarak dönen `recipes.Recipe` değeridir ve template onu `.`
  olarak alır. `recipe:pancakes` tag'i "bu page pancakes tarifini gösteriyor"
  anlamına gelir. O tarif değiştiğinde cache'teki kopyanın atılabilmesini sağlayan
  budur. Başka yerlerden de (bir test'ten, başka bir page'den) çağırdığınız bir
  loader ise `collage.DataHandler` ile kendi tipini dönebilir. Tag'i yoksa
  `collage.Load` kullanılır; bkz.
  [Data handler'lar](/docs/data-handlers#loaders-with-a-type-of-their-own).
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

Bu page de strateji tanımlamaz, ama home page'in aksine bir data handler'ı vardır. Bu
yüzden dynamic'tir: her request'te render edilir ve hiç cache'lenmez. collage,
`loadRecipe`'nin request'i ya da saati okuyup okumadığını anlamak için içine bakamaz.
Bu yüzden çıktısının herkes için aynı olduğunu varsaymaz. Page'i geliştirirken olması
gereken de budur. Page çalışır hâle geldiğinde bunu kendisi söyleyecek; bkz. aşağıdaki
[Export edin](#export-it) bölümü.

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

Handler'ın `recipes.Recipe` yerine `any` dönmesinin nedeni de budur. Veriyi okuyan
template'tir ve onu tipsiz okur: somut bir dönüş tipi de `{{.Name}}` hatasını
yakalamazdı.

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
		WithDataHandler(loadMore).
		Build()
}

func loadMore(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return nil, nil, err
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

`rc.Param("slug")` burada da çalışır: parametreler, path'i tanımlayan page'in
fragment'ine değil, request'e aittir.

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

Liste, page'deki tarifle ilgilidir. Bu yüzden her page'in paylaştığı layout'a değil,
tarif page'ine aittir. Listeyi tarif fragment'inin bir slot'una yerleştirin.
`pages/recipe.go` dosyasında:

```go
content := collage.NewFragment("recipe-content", "pages/recipe.html").
	WithDataHandler(loadRecipe).
	WithSlotFragment("more", MoreRecipes()).
	Required().
	Build()
```

Ardından slot'u ait olduğu yerde render edin. Bunun için
`templates/pages/recipe.html` dosyasında, `</main>`'den önce şunu ekleyin:

```html
  {{slot "more"}}
```

Slot'u hiçbir şey tanımlamaz. Layout'taki `{{slot "content"}}` gibi, template'teki
`{{slot "more"}}` tanımlar. Register işlemi, slot'a yapılan bağlamayı bu çağrıyla
karşılaştırır. Bu yüzden iki taraftan birindeki bir yazım hatası
(`WithSlotFragment("mroe", ...)`) programı `ErrUnknownSlot` ile durdurur. Hata,
slot'u ve template'in gerçekten çağırdığı slot'ları söyler. Dosyaları kaydedin.
Tarif page'inde artık diğer iki tarifin listesi var ve her biri kendi page'ine link
veriyor.

Bu page hakkında, hiçbir yerde açıkça yazılmamış üç şey geçerlidir.

- **Liste isteğe bağlıdır.** `WithSlot` zorunlu kılmadıkça bir slot isteğe
  bağlıdır. `MoreRecipes` de `Required()` değil. `loadMore` başarısız olursa page
  listesiz sunulur. Development'ta ayrıca hangi fragment'in neden başarısız olduğunu
  gösteren bir panel görünür. Bozuk bir sidebar, 500 hatası değil, yalnızca eksik
  bir sidebar olur.
- **Page'in tag'leri, iki fragment'in tag'lerinin birleşimidir.** Page artık hem
  `recipe:pancakes`'e hem de `recipes`'e bağlıdır. Page cache'lendikten sonra
  `recipes`'i invalidate etmek bütün tarif page'lerini cache'ten düşürür. Yeni bir
  tarif eklendiğinde olması gereken de budur. `recipe:pancakes`'i invalidate etmek
  ise yalnızca bu page'i düşürür.
- **İki handler birbirini gereğinden fazla beklemedi.** Bir child'ın handler'ı,
  parent'ınınki döndükten sonra başlar. Tek bir fragment'in slot'larındaki
  sibling'ler ise birlikte başlar. Tarif fragment'ine içinde yavaş bir fragment olan
  ikinci bir slot verirseniz page, sürelerin toplamını değil, en yavaş olanı bekler.
  Bkz. [Data handler'lar](/docs/data-handlers#when-handlers-run).

## Export edin

Tarifler request'ten request'e değişmez. Map'i düzenleyip deploy ettiğinizde
değişirler. Bu yüzden page'in her request'te render edilmesi gerekmez ve bütün site
static dosyalardan oluşabilir. `pages/recipe.go` dosyasındaki iki satır bunu söyler:

```go
	return collage.NewPage("recipe").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/recipes/{slug}").
		Static().
		WithStaticParams(recipeParams).
		Build()
}

// recipeParams lists the recipes a static build writes a page for.
func recipeParams(ctx context.Context, locale string) ([]map[string]string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return nil, err
	}
	params := make([]map[string]string, 0, len(list))
	for _, recipe := range list {
		params = append(params, map[string]string{"slug": recipe.Slug})
	}
	return params, nil
}
```

- **`Static()`**, collage'ın kendi başına karar veremediği şeydir. Page'in data
  handler'ları vardır ve döndürdüklerinin her okuyucu için aynı olduğunu yalnızca
  siz bilirsiniz. Sunulduğunda page artık bir kez render edilir ve tag'leri
  invalidate edilene kadar saklanır.
- **`WithStaticParams`**, bir build'in path'e bakarak cevaplayamayacağı soruyu
  cevaplar: hangi tarifler var? `/recipes/{slug}` birçok URL'i olan tek bir page'dir.
  Fonksiyon her URL için bir placeholder değerleri map'i döndürür. Page'in path'i
  olan her locale için bir kez çağrılır; bu sitenin tek locale'i var.

Export edin:

```sh
collage export
```

```
+ 6 files written
    dist/index.html
    dist/recipes/flatbread/index.html
    dist/recipes/omelette/index.html
    dist/recipes/pancakes/index.html
    dist/static/app.css
    dist/static/app.f106e88ebb47f51d.css

6 written · 0 skipped · 0 failed
```

Her tarif için bir dosya yazıldı. Her biri, bir request'in çalıştıracağı handler'lar
tarafından, o request'in taşıyacağı `slug` ile render edildi. Home page de siz bir
şey söylemeden oradadır: handler'ı yoktur, yani baştan beri static'tir.
`WithStaticParams` olmasaydı tarif page'i atlanır ve raporda adıyla gösterilirdi.
Prerender edemediği bir page içerdiği için build hatalı sayılmaz. `Static()`
olmasaydı da dynamic olduğu için atlanırdı. `collage serve`, `dist/`'i bir static
host'un göstereceği gibi gösterir; bkz. [Static export](/docs/static-export).

Listede olmayan bir tarif için dosya yazılmaz, ama çalışan bir server ona yine cevap
verir: `WithStaticParams`'ı yalnızca build okur. `/recipes/lasagne` her iki durumda
da 404 olarak kalır.

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
- [Caching](/docs/caching) — static bir page'in bir tarif değiştiğinde nasıl
  atıldığı ve bir page'in ne zaman her request'te render edilmesi gerektiği.
- [Static export](/docs/static-export) — `collage export`'un yazdığı, atladığı ve
  uyardığı her şey ve sonucun host edilmesi.
