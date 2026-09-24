---
description: Uygulamalı bir eğitim — bir layout, tipli bir data handler ve bir şablonla bir tarif sayfası kurun, ardından bir slot'a ikinci bir fragment ekleyin.
---

# İlk sayfanız

Bu eğitim, sıfırdan bir projede küçük bir tarif sitesi kurar: `/recipes/{slug}`
adresinde bir tarifi yükleyip sitenin layout'u içinde render eden bir sayfa, sonra
da ilk fragment'in bir slot'una yerleştirilen ikinci bir fragment — diğer
tariflerin listesi. Yaklaşık on beş dakika sürer ve her parçası, yazacağınız her
sayfada kullanacağınız bir parçadır.

Go'ya ve `collage` CLI'ına ihtiyacınız var; bkz. [Kurulum](/docs/installation).

## Bir proje başlatın

```sh
collage new cookbook --template minimal
cd cookbook
go mod tidy
collage dev
```

[http://localhost:3000](http://localhost:3000) adresini açın ve `collage dev`'i
çalışır hâlde bırakın. Bundan sonra kaydettiğiniz her Go dosyası yeniden derlenir
ve tarayıcı kendini yeniler; kaydettiğiniz her şablon ise yeniden derleme olmadan
bir sonraki render'da görünür.

## Layout'a bakın

Minimal bir projede zaten bir layout vardır ve her sayfa onu kullanır. İki
dosyadan oluşur. Fragment, `fragments/layouts/main.go` içinde:

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

Ve şablonu, `templates/layouts/default.html`:

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

`WithSlot("content", true, false)`, `content` adında, zorunlu ve tek bir fragment
tutan bir slot bildirir. `{{slot "content"}}` ise onun render edildiği yerdir. Bu
slot'u hiçbir zaman kendiniz doldurmazsınız: bir sayfa kaydedildiğinde collage
sayfanın içerik fragment'ini onun içine koyar. Tek bir layout'u her sayfanın
paylaşabilmesini sağlayan budur.

Şablonda `<title>` yok. Layout'un handler'ı `rc.HoistTitle` ile bir başlık bildirir
ve `{{hoist "head"}}` onun yerleştiği yerdir. Kendi başlığını bildiren bir sayfa,
ikinci bir başlık eklemek yerine sitenin başlığının yerini alır — birazdan tarif
sayfası da bunu yapacak. Bkz.
[Head ve SEO](/docs/head-and-seo#keys-and-the-innermost-wins).

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

İki ayrıntı ileride önem kazanacak. Eksik bir tarif, `collage.ErrNotFound`
`%w` ile sarılarak bildirilir; framework "bu yok"u (404) "bu bozuldu"dan (500)
böyle ayırt eder. Ve iki fonksiyon da bir `context.Context` alır, çünkü bir data
handler'ın süresi context'i üzerinden sınırlandırılır.

## İçerik fragment'i

Bir fragment bir şablondan ve isteğe bağlı olarak, şablonun render ettiğini getiren
bir fonksiyondan oluşur. `pages/recipe.go` dosyasını oluşturun:

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

Satır satır ele alalım.

- **`NewFragment("recipe-content", "pages/recipe.html")`** fragment'e ve şablonuna
  ad verir. Şablon yolu, uzantı dahil, `templates/`'e göredir.
- **`collage.DataHandler(loadRecipe)`**, kendi tipinize göre yazılmış bir handler'ı
  uyarlar. `loadRecipe` bir `recipes.Recipe` döndürür; dolayısıyla şablon da bir
  `recipes.Recipe` alır ve kodunuzun hiçbir yerinde tipsiz değerlerle uğraşmanız
  gerekmez.
- **Handler üç şey döndürür**: veri, verinin kurulduğu bağımlılık etiketleri ve bir
  hata. `recipe:pancakes` etiketi "bu sayfa pancakes tarifini gösteriyor" der; o
  tarif değiştiğinde önbellekteki kopyanın atılabilmesini sağlayan budur. Bkz.
  [Data handler'lar](/docs/data-handlers#the-contract).
- **`rc.Param("slug")`**, URL'nin eşleştiği `{slug}`'dır.
- **`rc.HoistTitle`** sayfaya kendi `<title>`'ını verir. İçerik fragment'i layout'un
  içinde durur ve en içteki bildirim kazanır; dolayısıyla layout'un `cookbook`
  başlığının yerini alır.
- **`Required()`**, sayfanın bu fragment olmadan var olamayacağını söyler. Bu fragment
  hata verirse sayfa da hata verir; `ErrNotFound`'u ise sayfayı 404 yapar.
- **`NewPage("recipe")`** sayfaya ad verir. Bağlantılar, testler ve plugin'ler
  sayfaya bu adla başvurur; bu yüzden URL değiştiğinde değişmemelidir.
- **`WithPath("en", "/recipes/{slug}")`**, sitenin varsayılan locale'indeki URL'dir.

Strateji metodu olmayan bir sayfa `Dynamic()`'tir: her istekte render edilir, hiç
önbelleğe alınmaz. Sayfayı kurarken doğru varsayılan budur. Çalıştığında
`Incremental(10 * time.Minute)` ya da `Static()` onu önbelleğe alır — bkz.
[Önbellek](/docs/caching).

## Şablon

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

`.`, handler'ın döndürdüğü `recipes.Recipe`'dir. Bu Go'nun `html/template`'idir;
dolayısıyla her değer, göründüğü yere göre kaçışlanır. `<script>` başlıklı bir
tarif çalıştırılmaz, yazdırılır. Bkz. [Şablonlar](/docs/templates).

## Sayfayı kaydedin

Uygulama bir sayfadan haberdar olana kadar o sayfa hiçbir şey yapmaz. `routes.go`'yu
açın ve listeye `pages.RecipePage()`'i ekleyin:

```go
for _, page := range []*collage.Page{pages.HomePage(), pages.RecipePage()} {
	if err := app.RegisterPage(page); err != nil {
		return fmt.Errorf("register page %q: %w", page.Name, err)
	}
}
```

Hatalar kayıt sırasında yakalanır. Yazım hatalı bir şablon yolu, içi boş zorunlu bir
slot, aynı adı taşıyan iki sayfa, hatalı bir yol — bunların her biri, ilk
ziyaretçinin bulmasını beklemek yerine, programı başlangıçta sayfayı adlandıran bir
hatayla durdurur.

## Görün

Kaydedin ve terminali izleyin: `collage dev` yeniden derler ve yeniden başlatır.
Şimdi [localhost:3000/recipes/pancakes](http://localhost:3000/recipes/pancakes)
adresini açın.

Sayfa, `content` slot'unda sizin fragment'iniz bulunan layout'tur.
`templates/pages/recipe.html`'i düzenleyin — bir cümle ekleyin, bir başlığı
değiştirin — tarayıcı değişiklikle yenilenir; yeniden derleme olmadı, çünkü
geliştirmede şablonlar her istekte diskten okunur.

Şimdi [/recipes/lasagne](http://localhost:3000/recipes/lasagne) adresini deneyin.
`recipes.Get` `collage.ErrNotFound`'u sardı, fragment `Required()`; dolayısıyla
yanıt bir 404'tür. Gövde collage'ın kendi sade bulunamadı sayfasıdır, çünkü proje
henüz bir tane kaydetmedi; `app.RegisterNotFoundPage` siteye kendi sayfasını verir
— bkz. [Sayfalar ve layout'lar](/docs/pages-and-layouts#not-found-and-error-pages).

Hatanın neye benzediğini görmek için bilerek bir şeyi bozun. Şablonda `{{.Title}}`'ı
`{{.Name}}` yapın, kaydedin ve yenileyin: alan yok, zorunlu fragment hata veriyor ve
sayfa bir 500. Geliştirmede hata sayfası, hatanın başladığı fragment olarak
`recipe-content`'i adlandırır ve `pages/recipe.html:2:8`'e ve bulamadığı alana kadar
hata zincirinin tamamını yazdırır. Geri alın.

## Bir slot'a ikinci bir fragment ekleyin

Bir sayfa nadiren tek parçadır. Diğer tariflerin bir listesini, kendi verisi olan
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

Şablonu, `templates/fragments/more-recipes.html`:

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

`pageURL "recipe" "slug" .Slug`, bağlantıyı sayfanın yolundan değil adından kurar;
böylece `/recipes/{slug}` bir gün `/r/{slug}` olursa bağlantı sayfayı izler. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales).

Şimdi tarif fragment'ine bir slot verin ve listeyi içine koyun. `pages/recipe.go`
içinde:

```go
content := collage.NewFragment("recipe-content", "pages/recipe.html").
	WithDataHandler(collage.DataHandler(loadRecipe)).
	WithSlot("more", false, false).
	WithSlotFragment("more", MoreRecipes()).
	Required().
	Build()
```

Ve slot'u ait olduğu yerde, `templates/pages/recipe.html` içinde `</main>`'den önce
render edin:

```html
  {{slot "more"}}
```

İkisini de kaydedin. Tarif sayfasında artık diğer iki tarifin, her biri kendi
sayfasına bağlanan bir listesi var.

Bu sayfa hakkında, hiçbir yere yazılmamış üç şey doğrudur.

- **Liste isteğe bağlıdır.** `WithSlot("more", false, false)` slot'u zorunlu değil
  olarak bildirdi ve `MoreRecipes` `Required()` değil. `loadMore` hata verirse sayfa
  liste olmadan sunulur — geliştirmede de hangi fragment'in neden hata verdiğini
  söyleyen bir panelle. Bozuk bir kenar çubuğu, eksik bir kenar çubuğudur; 500 değil.
- **Sayfanın etiketleri, iki fragment'in etiketleridir.** Sayfa artık
  `recipe:pancakes`'e ve `recipes`'e bağlıdır. Önbelleğe alındıktan sonra
  `recipes`'i geçersiz kılmak her tarif sayfasını düşürür — bir tarif eklemenin
  yapması gereken de budur — `recipe:pancakes`'i geçersiz kılmak ise yalnızca bu
  sayfayı düşürür.
- **İki handler birbirini gerektiğinden fazla beklemedi.** Bir çocuğun handler'ı,
  ebeveyninin handler'ı döndükten sonra başlar; tek bir fragment'in slot'larındaki
  kardeşler birlikte başlar. Tarif fragment'ine içinde yavaş bir fragment olan
  ikinci bir slot verin; sayfa onların toplamını değil, en yavaşını bekler. Bkz.
  [Data handler'lar](/docs/data-handlers#when-handlers-run).

## Test edin

Bir collage uygulaması sunucu olmadan test edilir: `app.Handler()` sıradan bir
`http.Handler`'dır ve onu `net/http/httptest` sürer. Uygulamayı `main`'in
kullandığı aynı `newApp` üzerinden kuran bir yardımcı ve bir test içeren bir
`main_test.go` oluşturun:

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

## Buradan nereye

- [Sayfalar ve layout'lar](/docs/pages-and-layouts) — yollar, stratejiler, hata
  sayfaları ve kaydın bir sayfaya ne yaptığı.
- [Fragment'ler ve slot'lar](/docs/fragments-and-slots) — slot'lar, hata politikası
  ve içerikten doldurulan slot'lar.
- [Data handler'lar](/docs/data-handlers) — handler sözleşmesi, fragment'ler arasında
  veri paylaşımı ve zaman aşımları.
- [Önbellek](/docs/caching) — bu sayfayı `Dynamic()` olmaktan çıkarıp bir kez render
  edilen ve bir tarif değiştiğinde atılan bir sayfaya dönüştürmek.
