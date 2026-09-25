---
description: Uygulamalı bir rehber. Bir layout, bir data handler ve bir template ile bir tarif page'i kurarsınız, ardından layout'un bir slot'una ikinci bir fragment eklersiniz.
reference: New, NewPage, NewFragment, FragmentBuilder.WithData, FragmentBuilder.WithTitle, DataHandler, Load, ErrNotFound, ErrUnknownSlot
---

# İlk page'iniz

Bu rehberde sıfırdan açılmış bir projede küçük bir tarif sitesi kurarsınız. Önce
`/recipes/{slug}` adresinde bir tarifi yükleyen ve onu sitenin layout'u içinde
render eden bir page yazarsınız. Ardından ikinci bir fragment eklersiniz: diğer
tariflerin listesi. Bu fragment layout'un bir slot'una yerleşir, böylece her
page'de görünür. Rehber yaklaşık on beş dakika sürer. İçindeki her parçayı,
yazacağınız her page'de yeniden kullanacaksınız.

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
gereken de budur. Page çalışır hâle geldiğinde `Incremental(10 * time.Minute)` ya da
`Static()` onu cache'ler; bkz. [Caching](/docs/caching).

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

Bir page nadiren tek parçadan oluşur. Tariflerin listesini, kendi verisi olan ayrı
bir fragment olarak ekleyin. Bu liste her page'de görünmeli, bu yüzden tarif
page'ine değil layout'a aittir. Layout'un yanına `fragments/layouts/more.go`
dosyasını oluşturun:

```go
package layouts

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/recipes"
)

// moreView is what the "more recipes" fragment renders.
type moreView struct {
	Recipes []recipes.Recipe
}

// MoreRecipes lists every recipe but the one on the page, if the page shows one.
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

`rc.Param("slug")` burada da çalışır: parametreler path'i tanımlayan fragment'e
değil, request'e aittir. Home page'de `{slug}` yoktur, değer boş gelir ve liste
bütün tarifleri içerir.

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

Şimdi listeyi layout'un ikinci bir slot'una yerleştirin.
`fragments/layouts/main.go` dosyasında:

```go
func Layout() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("cookbook").
		WithSlotFragment("more", MoreRecipes()).
		Build()
}
```

Ardından slot'u ait olduğu yerde render edin. Bunun için
`templates/layouts/default.html` dosyasında, `content` slot'undan sonra şunu
ekleyin:

```html
<body>
  {{slot "content"}}
  {{slot "more"}}
</body>
```

`content`, register işleminin doldurduğu tek slot'tur. Layout'un diğer slot'larını
siz bir kez doldurursunuz ve layout'u kullanan her page onları alır. `content`'te
olduğu gibi `more`'u da hiçbir şey tanımlamaz, template'teki `{{slot "more"}}`
tanımlar. Register işlemi, slot'a yapılan bağlamayı bu çağrıyla karşılaştırır. Bu
yüzden iki taraftan birindeki bir yazım hatası (`WithSlotFragment("mroe", ...)`)
programı `ErrUnknownSlot` ile durdurur. Hata, slot'u ve template'in gerçekten
çağırdığı slot'ları söyler. Dosyaları kaydedin. Tarif page'inde artık diğer
iki tarifin listesi var ve her biri kendi page'ine link veriyor. Home page ise üç
tarifin hepsini listeliyor.

Bu page'ler hakkında, hiçbir yerde açıkça yazılmamış üç şey geçerlidir.

- **Liste isteğe bağlıdır.** `WithSlot` zorunlu olduğunu söylemedikçe bir slot
  isteğe bağlıdır. `MoreRecipes` de `Required()` değil. `loadMore` başarısız olursa
  page listesiz sunulur. Development'ta ayrıca hangi fragment'in neden başarısız
  olduğunu gösteren bir panel görünür. Bozuk bir sidebar, 500 hatası değil, yalnızca
  eksik bir sidebar olur.
- **Bir page'in tag'leri, bütün fragment'lerinin tag'lerinin birleşimidir.** Tarif
  page'i artık hem `recipe:pancakes`'e hem de `recipes`'e, home page ise `recipes`'e
  bağlıdır. Page'ler cache'lendikten sonra `recipes`'i invalidate etmek listeyi
  gösteren bütün page'leri cache'ten düşürür. Yeni bir tarif eklendiğinde olması
  gereken de budur. `recipe:pancakes`'i invalidate etmek ise yalnızca pancakes
  page'ini düşürür.
- **İki handler birbirini beklemedi.** Bir child'ın handler'ı, parent'ınınki
  döndükten sonra başlar. Tek bir fragment'in slot'larındaki sibling'ler ise
  birlikte başlar. `content` ve `more` layout'un
  iki slot'udur, bu yüzden `loadRecipe` ve `loadMore` aynı anda çalışır ve page,
  sürelerin toplamını değil, yavaş olanı bekler. Bkz.
  [Data handler'lar](/docs/data-handlers#when-handlers-run).

Siz dokunmadığınız hâlde bir şey de değişti: home page artık dynamic. Strateji
tanımlamıyor ve içinde hiçbir şey veri çekmediği için static'ti. Layout her page'in
parçasıdır ve içindeki listenin bir data handler'ı vardır. Bu yüzden layout'u
kullanan her page, `Static()` ya da `Incremental(ttl)` diyene kadar her request'te
render edilir. Bkz. [Caching](/docs/caching#a-page-that-declares-none).

Yalnızca tek bir page'in ihtiyaç duyduğu bir fragment ise o page'in content
fragment'ine aittir: fragment'i content fragment'ine bağlayın, `{{slot}}` çağrısını
da onun template'ine ekleyin. Bir fragment nerede tanımlanırsa orada görünür.

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
- [Caching](/docs/caching) — bu page'i dynamic olmaktan çıkarıp bir kez render
  edilen ve bir tarif değiştiğinde atılan bir page'e dönüştürmek.
