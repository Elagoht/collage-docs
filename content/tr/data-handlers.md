---
description: Fragment'in verisini nasıl çektiği: handler sözleşmesi, dependency tag'ler, 404'ler, eşzamanlılık, render context'i, timeout'lar, template'in veriyle kontrolü.
reference: Data, FragmentBuilder.WithData, FragmentBuilder.WithTitle, FragmentBuilder.WithoutTypeCheck, LoadTagged, Load, Value, RenderContext, Key, NewKey, Once, Effect, ErrNotFound, ErrConflictingData, TemplateTypeError, ErrTemplateType, PanicError
---

# Data handler'lar

Data handler, bir fragment'in template'inde render edeceği veriyi almak için
çağırdığı fonksiyondur. Request'in context'ini ve render context'ini alır. Geriye
kendi Go tipindeki veriyi, bu verinin geldiği dependency tag'leri ve bir hata
döner.

Bir page'in dış dünyayla konuştuğu her şey (bir veritabanı, bir CMS, bir API) data
handler'larda olur, başka hiçbir yerde olmaz.

```go
content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
		post, err := store.Post(ctx, rc.Param("slug"))
		if err != nil {
			return nil, nil, err
		}
		return post, []string{"post:" + post.Slug}, nil
	})).
	Required().
	Build()
```

## Sözleşme

Bir fragment'in verisi `WithData` ile ayarlanır. `WithData` bir `collage.Data`
alır. Bunu yalnızca dört constructor üretir ve her biri verinin nereden geldiğini
söyler:

| Constructor | Ne çalıştırır | Template'in `.`'sı |
| --- | --- | --- |
| `collage.LoadTagged(fn)`, `fn` `(T, []string, error)` döner | Her render'da `fn`'i çalıştırır ve dependency tag'lerini bildirir | `T` |
| `collage.Load(fn)`, `fn` `(T, error)` döner | Her render'da `fn`'i çalıştırır, tag bildirmez | `T` |
| `collage.Value(v)` | Hiçbir şey: `v` her render'da verilir | `v`'nin tipi |
| `collage.Effect(fn)`, `fn` `error` döner | Her render'da `fn`'i, tanımladığı şeyler için çalıştırır | Hiçbir şey |

`WithData` kullanmayan bir fragment, `collage.Effect` kullanan bir fragment gibi
verisiz render edilir.

Handler, kendi tipini dönen sıradan bir fonksiyondur. Onu yukarıdaki gibi yerinde
yazabilir ya da bir isim verip geçirebilirsiniz:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	// ...
}

content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	Build()
```

`T` fonksiyonun imzasından gelir. Page register edilirken template de bu tipe
göre kontrol edilir (bkz.
[Template'ler nasıl kontrol edilir](#how-templates-are-checked)). Bu yüzden
handler ile template'in kendi verisine dair varsayımı birbirinden ayrışamaz.

Bu şekilde yazılan bir loader'ı bir test'ten ya da bir sitemap handler'ından
çağırmak da kolaydır: bunlar assert edilmesi gereken bir değer yerine bir `*Post`
alır. Constructor'ların `WithData`'nın bir biçimi değil de ayrı fonksiyonlar
olmasının nedeni, Go method'larının type parameter alamamasıdır.

v0.49.0'dan önce bir handler `WithDataHandler` ile verilir ve
`(any, []string, error)` dönerdi. Sabit veri ise `WithData(v)` ile verilirdi. İkisi
de kaldırıldı: handler'ı gerçek tipini dönecek şekilde `collage.LoadTagged` ile,
değeri de `collage.Value` ile sarın. (`LoadTagged`'ın adı v0.60.0'a kadar
`DataHandler` idi.)

Bir `LoadTagged`'ın üç dönüş değerinin her birinin ayrı bir görevi vardır.

- **Veri:** Template'in `.` olarak render ettiği şeydir. Template için yazılmış bir
  struct, yani bir "view", template'e bir veritabanı satırı vermekten genellikle
  daha anlaşılırdır. Her fragment kendi verisini alır. Bir child, parent'ının
  verisini görmez. Hata durumunda veri atılır. Böylece bir hatanın yanında dönen
  nil bir `*Post`, template'e mevcutmuş gibi görünen bir değer olarak hiç ulaşmaz.
- **Tag'ler:** Bu verinin oluşturulduğu içerik öğeleridir, örneğin
  `"post:hello-world"` ya da `"author:ada"`. Page'deki her fragment'ten, page'in
  kendi `WithDependencyTags` tag'leriyle birlikte toplanır ve cache'lenen page ile
  birlikte saklanır. Böylece bir tag'i invalidate ettiğinizde, o tag'i gösteren her
  page de invalidate edilir. Veri değişen hiçbir şeye bağlı değilse `nil` dönün.
  Ayrıntılar için [Caching](/docs/caching) sayfasına bakın.
- **Hata:** nil olmayan bir hata fragment'i başarısız kılar. Bunun page için ne
  anlama geldiğine fragment'in
  [failure policy'si](/docs/fragments-and-slots#when-a-fragment-fails) karar verir.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 244" role="img" aria-labelledby="dh-contract-title">
<title id="dh-contract-title">Bir data handler ctx ve rc alır ve üç şey döner. Veri template'in noktası olur ve hata varsa atılır. Tag'ler page'in tag'lerine eklenir, hata varsa korunur ve cache'lenen page'le saklanır. Hata, fragment'in failure policy'sine gider: required bir fragment page'i 500 ile, hata ErrNotFound'u wrap ediyorsa 404 ile başarısız kılar; fallback'i olan bir fragment fallback'ini render eder; diğerleri hiçbir şey render etmez.</title>
<defs><marker id="dh-contract-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">ctx, rc</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#dh-contract-arrow)" />
<rect class="d-a" x="90" y="50" width="140" height="34" rx="5" /><text class="d-on-a" x="160" y="65" text-anchor="middle">data handler</text><text class="d-on-a d-small" x="160" y="78" text-anchor="middle">LoadTagged(fn)</text>
<path class="d-flow" d="M130 84C130 96 56 92 56 104" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="160" y1="84" x2="160" y2="104" marker-end="url(#dh-contract-arrow)" />
<path class="d-flow" d="M190 84C190 96 264 92 264 104" marker-end="url(#dh-contract-arrow)" />
<rect class="d-frame" x="8" y="108" width="96" height="26" rx="5" /><text x="56" y="125" text-anchor="middle">veri: T</text>
<rect class="d-frame" x="112" y="108" width="96" height="26" rx="5" /><text x="160" y="125" text-anchor="middle">tag'ler</text>
<rect class="d-frame" x="216" y="108" width="96" height="26" rx="5" /><text x="264" y="125" text-anchor="middle">hata</text>
<line class="d-flow" x1="56" y1="134" x2="56" y2="150" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="160" y1="134" x2="160" y2="150" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="264" y1="134" x2="264" y2="150" marker-end="url(#dh-contract-arrow)" />
<rect class="d-a" x="8" y="154" width="96" height="26" rx="5" /><text class="d-on-a" x="56" y="171" text-anchor="middle">template'in .</text>
<rect class="d-b" x="112" y="154" width="96" height="26" rx="5" /><text class="d-on-b" x="160" y="171" text-anchor="middle">page tag'leri</text>
<rect class="d-c" x="216" y="154" width="96" height="26" rx="5" /><text class="d-on-c" x="264" y="171" text-anchor="middle">failure policy</text>
<text class="d-small" x="56" y="196" text-anchor="middle">hatada</text><text class="d-small" x="56" y="209" text-anchor="middle">atılır</text>
<text class="d-small" x="160" y="196" text-anchor="middle">hatada korunur</text><text class="d-small" x="160" y="209" text-anchor="middle">cache'lenen</text><text class="d-small" x="160" y="222" text-anchor="middle">page'le saklanır</text>
<text class="d-small" x="264" y="196" text-anchor="middle">Required(): 500,</text><text class="d-small" x="264" y="209" text-anchor="middle">ErrNotFound: 404</text><text class="d-small" x="264" y="222" text-anchor="middle">WithFallback: f</text><text class="d-small" x="264" y="235" text-anchor="middle">hiçbiri: boş</text>
</svg>
<figcaption>Data handler'a request'in context'i ve render context'i verilir; üç dönüş değerinin her biri ayrı bir yere gider. Veri template'in <code>.</code>'sı olur; hata durumunda atılır. Tag'ler page'in kendi tag'lerine eklenir ve page cache'lenince onunla saklanır; hata durumunda korunur. Hata, fragment'in failure policy'sine gider: <code>Required()</code> bir fragment page'i 500 ile, hata <code>collage.ErrNotFound</code>'u wrap ediyorsa 404 ile başarısız kılar; fallback'i olan bir fragment onu render eder; diğerleri hiçbir şey render etmez.</figcaption>
</figure>

Handler hata dönse bile tag'ler korunur. Neye bağlı olduğunu belirledikten sonra
başarısız olan bir handler, page'i neyin invalidate edeceğini yine de bildirmiştir.

Handler, strateji tanımlamayan bir page'in nasıl sunulacağını da belirler. Böyle
bir page, render ettiği herhangi bir şeyin data handler'ı varsa dynamic, yoksa
static olur. Çünkü bir handler request'i, bir cookie'yi ya da saati okuyabilir ve
fonksiyonun dışındaki hiçbir şey bunu yapıp yapmadığını bilemez.

Bunun yerine ne söyleneceği handler'ın çıktısına bağlıdır:

| Handler'ın çıktısı | Söylenecek | Ne yapar |
| --- | --- | --- |
| Her okuyucu için aynıysa (dosyadan okunan bir yazı gibi) | Page'in kendisinde `Static()` ya da `Incremental(ttl)` | Page tanımladığı stratejiyi korur |
| Yalnızca path'in parametrelerine ve locale'e bağlıysa | Handler'ın fragment'inde `Static()` | O fragment'i kullanan her page'i static bırakır |
| Her okuyucu için aynı ama zaman içinde değişiyorsa (bir ölçüm gibi) | Fragment'te `Shared()` | Render'ı okuyucular arasında paylaşılabilir ve page dynamic kalır |

Ayrıntılar için [Caching](/docs/caching#a-page-that-declares-none) sayfasına bakın.

### Sabit veri: collage.Value

Program başlarken belli olan veri (bir link listesi, bir başlık, bir site adı)
handler gerektirmez. `collage.Value(v)`, template'e her render'da `v`'yi verir:

```go
type homeView struct {
	Links []link
}

content := collage.NewFragment("home-content", "pages/home.html").
	WithData(collage.Value(homeView{Links: links})).
	Build()
```

Burada render başına çekilen hiçbir şey yoktur. Bu yüzden handler'ın aksine,
strateji tanımlamayan bir page'i static bırakır.

Bir fragment'in tek bir veri kaynağı vardır. `WithData`'yı veriyle iki kez
çağırmak, register sırasında `collage.ErrConflictingData` ile reddedilir. Hata
mesajı fragment'in adını verir: `fragment "home-content" has its data set twice`.
Kastettiğiniz çağrıyı tutun, diğerini silin. v0.49.0'dan önce ikinci çağrı sessizce
kazanırdı.

nil bir `Data` (`WithData(nil)` ya da nil bir fonksiyon verilmiş bir constructor)
veri yok demektir ve hiçbir şeyle çakışmaz.

### Tag'siz handler'lar

`collage.Load`, tag bildirmeyen, yalnızca veriyi ve bir hatayı dönen bir loader
için `LoadTagged`'ın karşılığıdır:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithData(collage.Load(loadClock)).
	Build()
```

Uygulamanın bir kez kurduğu bir şey (bir servis, bir store) üzerinden okuyan bir
handler'ı, onu parametre olarak alan bir fonksiyon üretir. Fragment'in kendi
fonksiyonu da bu parametreyi ona iletir. Böylece hiçbir kod global bir değişkene
uzanmaz:

```go
func Post(posts *blog.Service) *collage.Fragment {
	return collage.NewFragment("blog-post", "pages/blog-post.html").
		WithData(postData(posts)).
		Build()
}

func postData(posts *blog.Service) collage.Data {
	return collage.Load(func(ctx context.Context, rc *collage.RenderContext) (postView, error) {
		post, err := posts.BySlug(ctx, rc.Param("slug"))
		if err != nil {
			return postView{}, err
		}
		return postView{Title: post.Title}, nil
	})
}
```

`collage add page` ve `collage add fragment` handler'ı bu biçimde, henüz parametresiz
olarak yazar (v0.73.0'dan beri). Böylece bir servis eklemek, iki fonksiyona birer
parametre eklemekten ibarettir. VS Code'da eklentinin `cdata` snippet'i böyle bir
handler yazar ve dönüş tipi olarak dosyada tanımlı `…View` tipini kullanır.

İkisi de loader hata döndüğünde veriyi atar. Hangisini seçeceğiniz şöyle
belirlenir:

| Veri | Yazılacak |
| --- | --- |
| Her render'da aynıysa | `collage.Value(v)` |
| Değişiklikleri cache'lenmiş bir page'e ulaşması gereken içerikten çekiliyorsa | `collage.LoadTagged(fn)` |
| Çekiliyorsa ve bildirilecek tag yoksa | `collage.Load(fn)` |
| Veri yoksa ve fragment yalnızca page için bir şeyler tanımlıyorsa | `collage.Effect(fn)` (bkz. [aşağısı](#handlers-that-render-nothing)) |

Tag'leri yalnızca page cache'lenmiyorsa ya da veri hiç değişmiyorsa dışarıda
bırakın. Verisi değişen cache'lenmiş bir page, kendisini invalidate edebilmeleri
için tag'lerine ihtiyaç duyar. `collage.Value`'dan bir handler'a geçmek de yalnızca
fragment'i değiştirmez: strateji tanımlamayan bir page, onunla birlikte static'ten
dynamic'e geçer.

### Not found bir hata değildir

Eksik bir kayıt ile bozuk bir veritabanı farklı hatalardır. Okuyucu da her biri için
farklı bir cevap almalıdır: ilki için 404, ikincisi için 500. Hangisi olduğunu
`collage.ErrNotFound`'u wrap ederek belirtin:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	slug := rc.Param("slug")
	post, err := store.Post(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("blog: no post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("blog: load post %q: %w", slug, err)
	}
	return post, []string{"post:" + slug}, nil
}
```

`Required()` bir fragment'te, `collage.ErrNotFound`'u wrap eden bir hata page'in
not-found page'ini 404 ile render eder. Diğer her hata ise page'in error page'ini
500 ile render eder. Ayrıntılar için
[Page'ler ve layout'lar](/docs/pages-and-layouts#not-found-and-error-pages)
sayfasına bakın.

Kolayca gözden kaçan iki koşul vardır:

- **En üste kadar `%w` ile wrap edin.** Framework kontrolü `errors.Is` ile yapar.
  Bu yüzden store'unuz ile handler'ın dönüşü arasında herhangi bir yerde kullanılan
  bir `%v`, 404'ü 500'e çevirir.
- **Bunu yalnızca required bir fragment 404'e çevirir.** Optional bir fragment'te
  bu sıradan bir hatadır ve fragment ya hiçbir şey ya da fallback'ini render eder.
  Çünkü page o fragment olmadan da gösterilmeye değerdir.

## Handler'lar ne zaman çalışır

Bir page'in data handler'ları sırayla tek tek çalışmaz. **Bir fragment kendi
template'ini render etmeden önce, slot'larındaki bütün fragment'lerin data
handler'ları aynı anda başlar.** Bir yazı, bir sidebar ve bir yorum listesinden
oluşan bir page, üçünün toplam süresini değil, en yavaşını bekler.

Garanti edilen sıra şudur:

- **Parent'ın handler'ı, child'larınkiler başlamadan önce biter.** Bir child,
  parent'ının bir key altında sakladığı değeri ve parent'ının çözdüğü bir path
  parametresini okuyabilir.
- **Sibling'ler aynı anda çalışır.** Her biri kendi goroutine'inde, belirli bir
  sıra olmadan çalışır.
- **Template'ler yine de tek tek, ağaç sırasıyla render edilir.** Çıktı bir
  request'ten diğerine aynıdır. Sıraya bağlı her şeye (örneğin `<head>`'de hangi
  title'ın kazanacağına) hangi handler'ın önce bittiği değil, ağaç karar verir.

Bunun yazdığınız koda üç etkisi vardır.

- **Handler'lar sibling'leriyle eşzamanlı çalışmaya karşı güvenli olmalıdır.**
  Paylaştıkları her şey (bir map, bir sayaç, goroutine-safe olmayan bir client),
  Go'da başka her yerde gerektirdiği özeni burada da gerektirir.
- **Paylaşılan değerler bir `collage.Key` üzerinden geçer.** Key'in `Get` ve `Set`
  metotları render'ın lock'unu tutar. Ayrıntılar
  [aşağıda](#sharing-data-between-fragments).
- **Template'in render etmediği bir slot'taki fragment de yine başlar.** Template
  biter bitmez bu fragment'in context'i cancel edilir ve hatası yok sayılır.
  Pahalı olan ve çoğu zaman atlanan bir handler'ı template'teki bir koşulun
  arkasına değil, başka bir mekanizmanın arkasına koyun.

## Render context'i

`rc *collage.RenderContext`, bir handler'ın render edilen request hakkında bildiği
her şeydir.

| Üye | Size verdiği |
| --- | --- |
| `rc.Request` | Cevaplanan `*http.Request`. Static export'ta, page'in path'i için oluşturulmuş sentetik bir `GET` |
| `rc.Locale` | URL'nin çözüldüğü locale |
| `rc.Param(name)`, `rc.PathParams` | Route'taki `{name}` placeholder'larının yakaladığı değerler |
| `rc.Page` | Render edilen page. Bir action'da, action'ın cevap verdiği URL'nin page'i (v0.33.0'dan beri). **Yalnızca okunur** |
| `key.Get(rc)`, `key.Set(rc, value)` | Tek bir render'daki fragment'ler arasında bir `collage.Key` üzerinden paylaşılan değerler |
| `rc.Context()` | Render context'inin taşıdığı context |
| `rc.HoistTitle`, `rc.HoistMeta`, `rc.HoistProperty`, `rc.HoistLink`, `rc.HoistAlternate`, `rc.HoistStylesheet`, `rc.Hoist` | Page'in `<head>`'i için tanımlar. Bkz. [Head ve SEO](/docs/head-and-seo) |
| `rc.Asset(path)` | Mount edilmiş bir dosyanın content-addressed URL'si. Bkz. [Static asset'ler](/docs/assets) |
| `rc.URL(name, params)`, `rc.ActionURL(name, params)` | Bir page'in ya da action'ın adıyla bulunan, render'ın locale'indeki URL'si (v0.37.0'dan beri). Bkz. [Go'dan link'ler](/docs/links-and-locales#links-from-go) |

Handler'ın içinde `rc.Context()`, `ctx` argümanıyla aynı context'tir ve
fragment'in timeout'unu taşır. `ctx`'i kullanın, zaten elinizde olan odur.

Render context'inin kendisiyle ilgili iki kural vardır.

- **Onu saklamayın.** Render context'i tek bir render'a aittir. Onu handler
  döndükten sonra bir goroutine'de, bir cache'te ya da bir struct'ta tutmak, çoktan
  bitmiş bir request'i tutmak demektir.
- **`rc.Page`'e yazmayın.** O, register edilmiş tek `*collage.Page`'dir ve aynı anda
  o page'i render eden bütün request'ler tarafından paylaşılır. Bir handler'dan
  onun `Paths` map'ine ya da `DependencyTags`'ine yazmak, framework'ün canlı
  state'i üzerinde bir data race'tir. `go test -race` bunu raporlar, production'da
  ise er geç veriyi bozar. Request'ten request'e değişen her şeyi döndüğünüz veriye
  ya da bir `collage.Key` altına koyun.

## Fragment'ler arasında veri paylaşımı

Bir page'deki fragment'ler çoğu zaman aynı şeye ihtiyaç duyar. Yazı page'inin
içeriği, `<head>`'i ve "bu yazarın diğer yazıları" kutusu, hepsi yazının kendisine
ihtiyaç duyar.

### collage.Key

`collage.Key`, package seviyesinde bir kez tanımlanan bir isim ve bir tiptir. `Set`
bir değeri bu key altında render'ın geri kalanı için saklar. `Get` ise o değeri
aynı tiple geri okur:

```go
var postKey = collage.NewKey[*Post]("post")

// in the parent's handler
postKey.Set(rc, post)

// in a child's handler, which starts after the parent's has returned
post, ok := postKey.Get(rc)
if !ok {
	return nil, nil, errors.New("more-by-author: no post stored under postKey")
}
```

Key altında hiçbir şey saklanmamışsa `ok` false olur. Okurken type assertion
gerekmez; yanlış tipte bir yazma ise derlenmez.

Bir key, ismi *ve* tipinden oluşur:
`NewKey[*Post]("post")` ile `NewKey[*Draft]("post")` iki ayrı key'dir ve iki ayrı
değer tutar. Farklı dosyalarda tanımlanmış iki `NewKey[*Post]("post")` ise tek bir
key'dir. Key'ler uygulamanızın kendi namespace'idir. Çakışmayacak isimler seçin.

`With`, tanımlanmış bir key'den her değer için ayrı bir key türetir:
`postKey.With(slug)` key'inin adı `post:<slug>` olur ve aynı tipi tutar. `NewKey`
boş bir isimle çağrılırsa key'in tanımlandığı satırda panic eder. Zero value olan
bir `Key` de (`NewKey` ile hiç oluşturulmamış bir struct field'ı) kullanıldığı yerde
panic eder.

`Get` ve `Set` render'ın lock'unu alır. Eşzamanlı çalışan sibling'lere karşı
güvenli olmalarının nedeni budur. v0.50.0'dan önce değerler bir `SharedData` map'i
üzerinde `rc.Set(key, value)` ve `collage.Get[T](rc, key)` ile paylaşılıyordu. Bu
üçü de kaldırıldı.

Bu yöntem parent'tan child'a doğru çalışır, çünkü parent'ın handler'ı önce biter.
Sibling'ler arasında çalışmaz. Sibling'ler aynı anda çalıştığı için biri, diğerinin
o ana kadar bir şey set etmiş olmasına güvenemez.

### collage.Once

Sibling'ler için ya da aynı veriyi çekmesi gerekebilecek herhangi bir fragment grubu
için `collage.Once` kullanın. `collage.Once` bir key için çekme işlemini render
başına en fazla bir kez çalıştırır ve sonucu isteyen her fragment'e verir:

```go
func loadAuthorCard(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	slug := rc.Param("slug")
	post, err := collage.Once(rc, postKey.With(slug), func(ctx context.Context) (*Post, error) {
		return store.Post(ctx, slug)
	})
	if err != nil {
		return Author{}, nil, err
	}
	return post.Author, []string{"post:" + slug, "author:" + post.Author.ID}, nil
}
```

İlk çağıran veriyi çeker. Diğerleri onu bekler ve onun ürettiği sonucu alır. Akla
ilk gelen alternatif şudur: `Get` ile bakmak, bulunamazsa veriyi çekmek, sonra
`Set` ile yazmak. Bu yaklaşımda kontrol ile yazma arasında bir boşluk kalır. İki
sibling de bu boşluğa düşer ve ikisi de veriyi çeker.

- **Hata da bir sonuçtur.** Bekleyen herkes aynı hatayı alır. Çekme işlemi her
  fragment için yeniden denenmez. Aksi hâlde tek bir yavaş hata birkaç hataya
  dönüşürdü.
- **Bir render boyunca geçerlidir.** Ayarlamanız ya da temizlemeniz gereken hiçbir
  şey yoktur.
- **Kendi context'i sona eren bekleyen taraf beklemeyi bırakır** ve o hatayı
  döner.
- **Değerleri kendine aittir.** `Once`'ın çektiği değer `Get` ile okunamaz, `Set`'in
  sakladığı değer de `Once`'a verilmez. Adı aynı, tipi farklı iki key iki ayrı çekme
  işlemidir. v0.50.0'dan önce bu durum `ErrOnceTypeMismatch` hatasıydı.

### collage.Cached

`Once` işi bir page içinde paylaştırır. `collage.Cached` ise işi page'ler ve
request'ler arasında paylaştırır. Böylece aynı yazarın otuz yazısı, yazarı yalnızca
bir kez çeker.

```go
var authorKey = collage.NewKey[Author]("author")

author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
	func(ctx context.Context) (Author, error) { return api.Author(ctx, id) })
```

Verdiğiniz tag'ler page'in kendi tag'lerine eklenir. Bunlardan birini invalidate
ettiğinizde hem saklanan değer hem de o değerden oluşturulmuş bütün cache'lenmiş
page'ler stale olarak işaretlenir. Render'lar arasında hiçbir şeyin tutulmadığı durumlarda (cache
kapalıyken, development'ta, bir preview'da) `collage.Cached` tam olarak `Once`
gibi davranır. Ayrıntılar [Caching](/docs/caching#caching-data-across-pages)
sayfasındadır.

Fragment'ler ayrı ayrı yenileniyorsa başvurulacak olan da `Cached`'dir. Her
[fragment path'i](/docs/forms-and-actions#a-fragment-at-its-own-url) kendi başına
bir render'dır. Bu yüzden `Once` iki fragment path'i arasında hiçbir şey paylaşmaz,
`Cached` ise paylaşır.

## Hiçbir şey render etmeyen handler'lar

Bazı fragment'ler markup render etmek için değil, page için bir şeyler tanımlamak
için vardır: bir title, bir canonical link, structured data gibi. `collage.Effect`,
yalnızca hata dönen bir handler'ı adapte eder:

```go
seo := collage.NewFragment("post-seo", "fragments/empty.html").
	WithData(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		post, err := collage.Once(rc, postKey.With(rc.Param("slug")), func(ctx context.Context) (*Post, error) {
			return store.Post(ctx, rc.Param("slug"))
		})
		if err != nil {
			return err
		}
		rc.HoistTitle(post.Title)
		rc.HoistMeta("description", post.Summary)
		return nil
	})).
	Build()
```

Fragment'in template'i hiçbir veri almaz ve handler hiçbir tag bildirmez. Handler'ın
tanımladığı şeyler değişen bir içerikten geliyorsa ve page cache'leniyorsa, o
içeriğin tag'lerinin yine de page'e ulaşması gerekir. Bu durumda tag'leri bunun
yerine sıradan bir handler'dan dönün ya da page üzerinde `WithDependencyTags` ile
ekleyin.

Program başlarken belli olan bir title (layout'taki site adı gibi) hiç handler
gerektirmez. Fragment üzerindeki `WithTitle(s)` onu tanımlar ve strateji
tanımlamayan bir page'i static bırakır. Ayrıntılar için
[Head ve SEO](/docs/head-and-seo) sayfasına bakın.

## Örnek: bir katalogdaki ürün page'i

Bir mağazanın `/products/{id}` adresindeki ürün page'i üç servisten beslenir ve her
veri kendisine uyan yolla çekilir:

| Veri | Nereden gelir | Nasıl çekilir |
| --- | --- | --- |
| Ad ve marka id'si | Katalog | Ürünün handler'ı, her render'da |
| Fiyat | Bir fiyat servisi | Aynı handler, katalogdan sonra |
| Marka: ad ve ülke | Katalog; birçok ürünün paylaştığı yavaş bir sorgu | `collage.Cached`, bir gün için, `brand:<id>` tag'iyle |
| Yorumlar | Bir yorum servisi | `collage.Once`, iki sibling fragment'ten: yorum kutusu ve page'in description'ı |

Buradaki catalog, pricing ve reviews package'ları sizin kendi client'larınızı
temsil eder. Ürünün handler'ı:

```go
var (
	brandKey   = collage.NewKey[catalog.Brand]("brand")
	reviewsKey = collage.NewKey[reviews.Summary]("reviews")
)

type productView struct {
	Name  string
	Price string
	Brand catalog.Brand
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (productView, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Product(ctx, id)
	if errors.Is(err, catalog.ErrNoProduct) {
		return productView{}, nil, fmt.Errorf("shop: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: load product %q: %w", id, err)
	}
	price, err := pricing.Price(ctx, id)
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: price of %q: %w", id, err)
	}
	// Many products share a brand: one fetch serves all of them.
	brand, err := collage.Cached(rc, brandKey.With(p.BrandID), 24*time.Hour, []string{"brand:" + p.BrandID},
		func(ctx context.Context) (catalog.Brand, error) { return catalog.LookupBrand(ctx, p.BrandID) })
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: brand %q: %w", p.BrandID, err)
	}
	return productView{Name: p.Name, Price: price, Brand: brand}, []string{"product:" + id, "price:" + id}, nil
}
```

- **Bilinmeyen bir id 404'tür.** Handler `collage.ErrNotFound`'u wrap eder ve
  fragment'i `Required()`'dır. Bu yüzden okuyucu not-found page'ini görür. Diğer
  her hata page'i 500 ile başarısız kılar.
- **Marka, bütün ürünleri için bir kez çekilir.** `brand:<id>` tag'i, handler onu
  dönmüş gibi page'in tag'lerine eklenir. Bu yüzden handler yalnızca
  `product:<id>` ve `price:<id>` döner.

Yorumlara iki yerde ihtiyaç vardır: yorum kutusunda ve puanı page'in
description'ına koyan bir fragment'te. İkisi de aynı fonksiyonu çağırır:

```go
// reviewsOf is called by two sibling fragments; Once makes it one call per render.
func reviewsOf(rc *collage.RenderContext, id string) (reviews.Summary, error) {
	return collage.Once(rc, reviewsKey.With(id), func(ctx context.Context) (reviews.Summary, error) {
		return reviews.For(ctx, id)
	})
}

func loadReviews(ctx context.Context, rc *collage.RenderContext) (reviews.Summary, []string, error) {
	id := rc.Param("id")
	summary, err := reviewsOf(rc, id)
	if err != nil {
		return reviews.Summary{}, nil, err
	}
	return summary, []string{"reviews:" + id}, nil
}

func productSEO(ctx context.Context, rc *collage.RenderContext) error {
	summary, err := reviewsOf(rc, rc.Param("id"))
	if err != nil {
		return err
	}
	rc.HoistMeta("description", fmt.Sprintf("Rated %.1f out of 5 in %d reviews.", summary.Average, summary.Count))
	return nil
}
```

İki fragment sibling'dir, bu yüzden handler'ları aynı anda başlar. `reviewsOf`'a
ilk ulaşan yorum servisini çağırır, diğeri onun sonucunu bekler. Bir `Get`,
bulunamazsa çekme ve ardından bir `Set` ise servisi iki kez çağırabilirdi.

Ürün ikisini de kendi slot'larına bağlar:

```go
reviewsBox := collage.NewFragment("product-reviews", "fragments/reviews.html").
	WithData(collage.LoadTagged(loadReviews)).
	WithTimeout(500 * time.Millisecond).
	Build()

seo := collage.NewFragment("product-seo", "fragments/empty.html").
	WithData(collage.Effect(productSEO)).
	WithTimeout(500 * time.Millisecond).
	Build()

content := collage.NewFragment("product", "pages/product.html").
	WithData(collage.LoadTagged(loadProduct)).
	WithSlotFragment("reviews", reviewsBox).
	WithSlotFragment("seo", seo).
	Required().
	Build()

page := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(content).
	WithPath("en", "/products/{id}").
	Incremental(10 * time.Minute).
	Build()
```

```html
<!-- templates/pages/product.html -->
<article>
  <h1>{{.Name}}</h1>
  <p class="price">{{.Price}}</p>
  <p class="brand">{{.Brand.Name}}, {{.Brand.Country}}</p>
  {{slot "reviews"}}
</article>
{{slot "seo"}}
```

Description, layout'taki `{{hoist "head"}}` marker'ının olduğu yere yerleşir (bkz.
[Head ve SEO](/docs/head-and-seo)). İki fragment'e aynı timeout verilir, çünkü tek
çağrı, onu ilk isteyen fragment'in context'i altında çalışır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 200" role="img" aria-labelledby="dh-share-title">
<title id="dh-share-title">Once ve Cached yan yana. Solda Once: tek bir render'da sibling fragment'ler reviews ve seo, reviews:42 key'ini ister ve tek bir çağrı ikisine de yeter; sonraki render yeniden çağırır. Sağda Cached: ürün 42'nin ve ürün 43'ün render'ları brand:acme key'ini ister ve tek bir çağrı data cache'te tutulur; 24 saat ya da brand:acme invalidate edilene kadar tazedir.</title>
<defs><marker id="dh-share-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<line class="d-rule" x1="160" y1="8" x2="160" y2="192" />
<text x="80" y="20" text-anchor="middle">Once</text>
<rect class="d-frame" x="6" y="28" width="146" height="128" rx="6" />
<text class="d-small" x="14" y="42">tek render</text>
<rect class="d-b" x="16" y="50" width="60" height="24" rx="4" /><text class="d-on-b d-small" x="46" y="66" text-anchor="middle">reviews</text>
<rect class="d-b" x="84" y="50" width="60" height="24" rx="4" /><text class="d-on-b d-small" x="114" y="66" text-anchor="middle">seo</text>
<path class="d-flow" d="M46 74C46 86 76 82 76 94" marker-end="url(#dh-share-arrow)" />
<path class="d-flow" d="M114 74C114 86 84 82 84 94" marker-end="url(#dh-share-arrow)" />
<rect class="d-chip" x="22" y="98" width="116" height="22" rx="11" /><text class="d-small" x="80" y="113" text-anchor="middle">reviews:42</text>
<text class="d-small" x="80" y="142" text-anchor="middle">tek çağrı, ortak</text>
<text class="d-small" x="80" y="176" text-anchor="middle">sonraki render:</text>
<text class="d-small" x="80" y="189" text-anchor="middle">yeniden çağrılır</text>
<text x="242" y="20" text-anchor="middle">Cached</text>
<rect class="d-frame" x="168" y="28" width="70" height="46" rx="6" /><text class="d-small" x="203" y="48" text-anchor="middle">ürün 42</text><text class="d-small" x="203" y="61" text-anchor="middle">render'ı</text>
<rect class="d-frame" x="246" y="28" width="70" height="46" rx="6" /><text class="d-small" x="281" y="48" text-anchor="middle">ürün 43</text><text class="d-small" x="281" y="61" text-anchor="middle">render'ı</text>
<path class="d-flow" d="M203 74C203 86 238 82 238 94" marker-end="url(#dh-share-arrow)" />
<path class="d-flow" d="M281 74C281 86 246 82 246 94" marker-end="url(#dh-share-arrow)" />
<rect class="d-chip" x="184" y="98" width="116" height="22" rx="11" /><text class="d-small" x="242" y="113" text-anchor="middle">brand:acme</text>
<rect class="d-soft-a" x="168" y="128" width="148" height="28" rx="6" /><text class="d-small" x="242" y="146" text-anchor="middle">data cache: tek çağrı</text>
<text class="d-small" x="242" y="176" text-anchor="middle">taze: 24 saat ya da</text>
<text class="d-small" x="242" y="189" text-anchor="middle">invalidate edilene dek</text>
</svg>
<figcaption><code>Once</code> bir çağrıyı tek bir render'ın fragment'leri arasında, <code>Cached</code> ise render'lar arasında paylaştırır. Ürün page'inde yorum kutusu ile description her render'da yorum servisine yapılan tek bir çağrıyı paylaşır; bir markanın bütün ürünleri de tek bir marka sorgusunu paylaşır. Render'lar arasında hiçbir şeyin tutulmadığı yerde (cache kapalıyken, development'ta, bir preview'da) <code>Cached</code>, <code>Once</code> gibi davranır.</figcaption>
</figure>

Her hatanın bu page'e etkisi:

- **Ürün 7 yoktur.** Required fragment'in hatası `collage.ErrNotFound`'u wrap
  eder. Sonuç, 404 ile not-found page'idir.
- **Katalog ya da fiyat servisi başarısız olur.** Ürün required olduğu için
  page'in error page'i 500 ile sunulur.
- **Marka sorgusu başarısız olur.** Henüz saklanmış bir marka yoksa bu da 500'dür.
  Stale bir marka saklanmışsa `Cached` onu sunar ve stale veriden
  oluşturulan page cache'lenmez.
- **Yorum servisi çökmüştür ya da 500 milisaniyeden yavaştır.** Servise yapılan
  tek çağrı başarısız olur ve iki fragment de aynı hatayı alır: yorum kutusu
  hiçbir şey render etmez, description tanımlanmaz ve page onlarsız 200 ile cevap
  verir.
  Başarısız bir fragment'le sunulan page degraded'dır ve hiç cache'lenmez. Bu
  yüzden sonraki request yorum servisine yeniden sorar. Page cache'te daha eski
  bir kopya duruyorsa ise o kopya `STALE` olarak işaretlenip sunulur.

Cache açıkken page taze olduğu sürece bunların hiçbiri çalışmaz: cache'lenmiş bir
page render edilmez. Bu yüzden `/products/42`'nin on dakika içindeki sonraki
okuyucusu hiçbir çağrıya mal olmaz. Bkz.
[Caching](/docs/caching#once-cached-and-the-page-cache).

## Timeout'lar ve context

Her data handler bir deadline altında çalışır. Bu deadline fragment'in
`WithTimeout(d)` değeridir. Timeout ayarlamayan bir fragment için ise
`Config.Template.Timeout` kullanılır, varsayılan değeri beş saniyedir.

Deadline, handler'ın aldığı `ctx` üzerindedir. **Handler'ı değil, context'i
sınırlar.**

Framework bir handler'ı deadline'ı geldiğinde bırakmaz, dönmesini bekler. Çünkü
bir goroutine dışarıdan durdurulamaz. Hiç dönmeyen handler'ları bırakmak, her
request için bir goroutine'in sonsuza kadar sızması anlamına gelirdi. Bu yüzden
context'ini yok sayan bir handler deadline'ını aşarak çalışabilir ve page onu
bekler.

Deadline'a uymak tek bir alışkanlığa bağlıdır: bekleyebilecek her şeye `ctx`'i
geçirin.

```go
func loadWeather(ctx context.Context, rc *collage.RenderContext) (*Weather, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherURL(rc.Locale), nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err // context.DeadlineExceeded when the timeout passed
	}
	defer res.Body.Close()

	var weather Weather
	if err := json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return nil, nil, err
	}
	return &weather, nil, nil
}
```

Süre dolduğunda context'inin hatasını dönen bir handler, diğer handler'lar gibi
başarısız olur. Bu hata yine de `errors.Is(err, context.DeadlineExceeded)` ile
eşleşir. Deadline ancak handler onu hata olarak dönerse bir hatadır. `ctx`'i yok
sayıp geç de olsa `nil` dönen bir handler başarılı olmuştur ve verisi render
edilir.

Olmasa da olur türündeki her şey için kısa bir timeout'u bir fallback ile birlikte
kullanın. Böylece page, yavaş bir servisi sizin seçtiğiniz noktada beklemeyi
bırakır.

Hata, hangi deadline'ın dolduğunu söyler. Yalnızca fragment'in kendi timeout'u
timeout olarak bildirilir: `collage: execution exceeded 5s`. Daha yukarıda biten
bir context (iptal edilen bir request ya da daha kısa bir deadline) ise varsa
nedeniyle birlikte `collage: execution stopped by context` olarak bildirilir.
v0.18.1'den önce ikisi de timeout gibi görünürdü ve iptal edilen bir request sizi
hiç yavaş olmamış bir handler'ı aramaya gönderirdi.

## Panic'ler

Bir data handler'daki panic process'i çökertmez. Panic recover edilir ve panic
değerini ve stack'i taşıyan bir `*collage.PanicError`'a dönüştürülür. Fragment da
diğer her hatada olduğu gibi bu hatayla başarısız olur. Development'ta error page
stack'i gösterir. Kendi kodunuzda bu hataya `errors.As` ile ulaşabilirsiniz.

## Template'ler nasıl kontrol edilir

Bir template verisini isimle okur: `{{.Title}}`, `{{.Author.Name}}`.
`html/template` bu isimleri ancak template çalışırken çözer. Bir `{{.Titel}}`,
page'inin her render'ını başarısız kılar. Kimsenin ziyaret etmediği bir page'de
ise hiç ortaya çıkmaz.

v0.49.0'dan beri her fragment'in verisi tipli bir constructor'dan gelir. Bu
yüzden register işlemi her template'in hangi Go tipiyle çalışacağını bilir ve
template'i bu tipe göre dolaşır. `RegisterPage`,
template'leri verilerine uymayan bir page'i reddeder. Hata mesajında page'in,
fragment'in ve dosyanın adı, satır ve sütun yer alır:

```text
collage: page "post": fragment "post-body" (post.html:1:6): {{.Titel}}: type blog.Post has no field or method Titel (did you mean Title?)
```

Kontrol yalnızca kesin olarak başarısız olacak şeyleri bildirir: her bulgu, render ona
ulaştığında `text/template`'in başarısız olacağı bir ifadedir. Artık başlangıçta
başarısız olan bir uygulamanın template'i, render edildiğinde zaten başarısız
olacaktı. Kontrol bunu daha erken bildirir ve kendinden bir kural eklemez.

### Neler dolaşılır

Bir page'in ulaştığı her fragment: layout'ları, content'i, bunların slot'larına
bağlanan her şey, fallback'ler, inline fragment'ler,
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile açılan
fragment'ler ve page'in
[not-found ve error page'leri](/docs/pages-and-layouts#not-found-and-error-pages)
olarak verilen page'ler. `{{template "partials/author.html" .Author}}` ile dahil
edilen bir partial, kendisine verilen tiple dolaşılır. Uygulama genelinde
kendisine verilen her tip için de bir kez dolaşılır.

Bir [slot resolver'ın](/docs/fragments-and-slots#slots-filled-per-render) döndüğü
fragment page render edilirken oluşturulur. Bu yüzden register işlemi onu hiç görmez
ve template'i kontrol edilmez.

`.` fragment'in veri tipiyle başlar ve template'i izler: `{{range}}` içinde
elemandır, `{{with}}` içinde değerdir. Bir `$x := …` değişkeni de kendi scope'u
boyunca verildiği tipi korur.

### Neler bildirilir

- **Tipin sahip olmadığı bir alan ya da method** veya export edilmemiş bir alan.
  Pointer'lar izlenir, embedded struct'lardan promote edilen alanlar da sayılır.
  Yakın bir isim varsa, en yakın export edilmiş isim önerilir.
- **Adreslenebilir olmayan bir değer üzerinden ulaşılan, pointer receiver'lı bir
  method.** `text/template` böyle bir method'u değerin adresi üzerinden çağırır.
  Template'e değer olarak verilen verinin adresi yoktur; alanlarının ve bir map'in
  elemanlarının da yoktur. Bir pointer'ın gösterdiği şey ve bir slice'ın
  elemanları adreslenebilirdir. `URL`, `*Post` üzerinde tanımlıysa `{{.URL}}` bir
  `Post` üzerinde başarısız olur, bir `*Post` üzerinde çalışır. Çözüm, template'e
  pointer'ı vermektir.
- **Key'i isimli bir string tipi olan bir map** (`map[Slug]Post`): `.key`, key'i
  düz bir `string` olarak arar ve böyle bir map bunu kabul etmez. Key'i `string`
  olan bir map sorun değildir. Map'te olmayan bir key de hata değildir: zinciri
  sessizce bitirir, bu yüzden `{{.Meta.absent.Name}}` hiçbir şey render etmez.
  Yalnızca map'in eleman tipini bilen kontrol, o eleman tipinde olmayan bir ismi
  yine de bildirir.
- **Yanlış biçimdeki çağrılar:** yanlış sayıda argüman verilen bir method,
  template fonksiyonu ya da built-in; argüman verilen bir alan ya da map key'i;
  bir değer ve bir hatadan fazlasını dönen bir method ya da fonksiyon; fonksiyon
  olmayan bir şey üzerinde `call`.
- **Yanlış türdeki değerler üzerinde `range`, `len`, `index` ve `slice`:** bir
  string ya da struct üzerinde range, iki değişkenle bir tamsayı üzerinde range,
  bir struct'ın `len`'i, bir struct'a `index`. String, slice ya da array dışında
  bir şeyin `slice`'ı, bir string'in üç index'le `slice`'ı ve `text/template`'in
  adresini alamadığı bir array'in `slice`'ı da bildirilir. Örneğin veri değer
  olarak verildiğinde `{{slice .Days 0}}` bildirilir; çözüm pointer'ı vermektir.
  Slice, array ve string tamsayılarla index'lenir ve slice'lanır. Map ise kendi
  key tipindeki bir key'le index'lenir: bir `map[int]Post` üzerinde
  `{{index .ByID "7"}}` bildirilir.
- **Sayı sabitleri**, `text/template`'in onlara verdiği tipi alır: `1` bir `int`,
  `1.5` bir `float64`, `1i` bir `complex128`, `'a'` bir `int`'tir. Bu yüzden
  `{{(1.5).Nope}}` ve `{{len 1}}` bildirilir. `int`'e sığmayacak kadar büyük bir
  sabit ve komut olarak kullanılan `nil` de bildirilir.
- **Parametresinin alamayacağı bir argüman** (v0.51.0'dan itibaren).
  `text/template` bir template fonksiyonunu ya da method'u çağırırken argümanı
  nasıl değerlendiriyorsa kontrol de öyle değerlendirir:
  - Bir sabit, parametrenin türüne göre değerlendirilir. Tamsayı parametresi
    tamsayı olarak yazılmış bir sayıyı alır, unsigned parametre negatif olmayan
    bir tamsayıyı alır. Float parametresi complex olmayan her sayıyı, string ya
    da bool parametresi de bir string'i ya da bool'u alır. Bu yüzden
    `{{upper 1}}`, `{{printf 1}}`, `{{uint -1}}` ve `{{i8 1.5}}` bildirilir.
    `nil` de yalnızca nil alabilen bir parametreye verilebilir.
  - Diğer her değer parametreye atanabilir olmalı, atanabilir bir şeyin
    pointer'ı olmalı ya da parametre onun pointer'ıysa adreslenebilir olmalıdır.
    Bu yüzden `int` bir alanla `{{upper .Count}}` ve veri değer olarak
    verildiğinde bir `func(*User)` için `{{edit .Owner}}` bildirilir. `*User`
    tipinde bir alanla bir `func(User)` için `{{show .Author}}` ise sorun
    değildir. Variadic bir parametre kendi eleman tipini alır. Pipe ile verilen
    değer de (`{{.Count | upper}}`) diğer argümanlar gibi değerlendirilir.
  - `call`, ilettiği değerleri `text/template`'in `call`'u gibi kontrol eder:
    parametreye atanabilir bir değer ya da başka bir tamsayı tipine
    dönüştürülen bir tamsayı geçer, başka hiçbir şey geçmez.

Render'ın hiç ulaşamayacağı şeyler bildirilmez. Koşulu bir literal olan bir
`{{if}}` ya da `{{with}}` (`true`, `false`, `0`, `1`, `""`, `"x"` ya da bunlardan
birinin `not`'u), literal'in dışarıda bıraktığı tarafı hiç çalıştırmaz. Koşul bir
değişken tanımlasa ya da bir değişkene değer atasa bile bu böyledir:
`{{if $x := false}}{{.Nope}}{{end}}` hiçbir şey bildirmez. `{{and}}`
ve `{{or}}` yalnızca sonucu belirleyen bir literal'de durur: `and` için `false`,
`0`, `""` ya da `nil`; `or` için `true`, sıfır olmayan bir sayı ya da boş olmayan
bir string. Bu yüzden `{{and 0 .Nope}}` hiçbir şey bildirmez. `{{and 1 .Nope}}` ve
`{{and .Title .Nope}}` ise `.Nope`'u bildirir, çünkü render ona ulaşabilir.

### Neler bildirilmez

- **Render'dan önce bilinemeyenler.** Bir interface (`any`, bir
  `map[string]any`'nin elemanları, bir `error`) her şeyi tutabilir. Bu yüzden
  ondan okunan hiçbir şey bildirilmez. Bir method'un ya da fonksiyonun döndüğü bir
  `reflect.Value`'dan okunanlar da bildirilmez, çünkü `text/template` onu tuttuğu
  değere açar.
- **Verisi olmayan bir fragment.** nil verinin bir alanını okumak `html/template`'te
  hata değildir: `{{.Title}}` hiçbir şey render etmez. `WithData` kullanmayan ya
  da `collage.Effect` kullanan bir fragment yine dolaşılır, ama yalnızca fonksiyon
  çağrıları kontrol edilir: argüman sayıları ve `.`'dan okunmayan argümanlar.
- **Verinin içindeki nil pointer'lar.** `Author` nil ise `{{.Author.Name}}` render
  edilirken başarısız olur. Onun nil olup olmadığı tipin değil, verinin işidir.
- **Uyup uymadığına render'ın karar verdiği bir argüman.** Interface tipinde bir
  değer, bir `reflect.Value`, `any` ya da `reflect.Value` tipinde bir parametre
  ve nil alabilen bir parametreye verilen bir map elemanı değerlendirilmez. Map'te
  olmayan bir key nil olarak gelir ve o parametre nil'i alır. Bir map key'iyle
  okunan bir değerin method'una verilen argümanlar da değerlendirilmez, çünkü key
  yoksa render o method'u hiç çağırmaz. Aralık dışındaki bir index de
  bildirilmez. Kontrol açısından bir plugin'in render fonksiyonları her şeyi
  alır, bu yüzden argümanları plugin'e bırakılır.
- **Bir tamsayı sabitinin aralığı.** `text/template` bir sabiti tamsayı
  parametresine aralık kontrolü yapmadan dönüştürür. Bu yüzden `{{i8 300}}`
  başarısız olmaz, 44 render eder.

### Hatayı okumak

`RegisterPage`, page'in kendi fragment'lerindeki bütün bulguları bir kerede,
`errors.Join` ile birleştirip dosya, satır ve sütuna göre sıralayarak döner.
Böylece on hata, on değil tek bir yeniden başlatma demektir. Page'in not-found
page'i, page'in kendisi geçtikten sonra kontrol edilir. Error page'i ise ikisi de
geçtikten sonra kontrol edilir. Bunların bulguları bir
`collage: page "post" not-found page: …` (ya da `error page: …`) hatasıyla sarılı
olarak, page'in kendi bulguları düzeltildikten sonraki başlangıçta gelir.

Her bulgu bir `*collage.TemplateTypeError`'dır ve her biri
`collage.ErrTemplateType` ile eşleşir. Bir bulgu, bir join'in yanı sıra sarmalayan
bir hatanın altında da durabilir. Bu yüzden hepsini listelemek için bütün ağacı
dolaşın:

```go
// typeErrors collects every *collage.TemplateTypeError in err's tree.
func typeErrors(err error) []*collage.TemplateTypeError {
	switch e := err.(type) {
	case nil:
		return nil
	case *collage.TemplateTypeError:
		return []*collage.TemplateTypeError{e}
	case interface{ Unwrap() []error }:
		var found []*collage.TemplateTypeError
		for _, inner := range e.Unwrap() {
			found = append(found, typeErrors(inner)...)
		}
		return found
	case interface{ Unwrap() error }:
		return typeErrors(e.Unwrap())
	}
	return nil
}
```

ve page'in register edildiği yerde:

```go
if err := app.RegisterPage(page); errors.Is(err, collage.ErrTemplateType) {
	for _, typeErr := range typeErrors(err) {
		fmt.Printf("%s:%d:%d %s\n", typeErr.Template, typeErr.Line, typeErr.Col, typeErr.Reason)
	}
}
```

Dönen hata üzerinde `errors.As` ilk bulguyu bulur. Alanları:

| Alan | Ne tutar |
| --- | --- |
| `Page`, `Fragment` | Template'in ait olduğu page ve fragment |
| `Template` | İfadenin bulunduğu dosya (fragment'in kendi dosyası ya da dahil ettiği bir partial) ya da `inline template of fragment "x"` |
| `Line`, `Col` | `text/template`'in hatayı göstereceği yer. Bu, `Expr`'in başı değil içi olabilir, örneğin bir çağrının son argümanı |
| `Expr` | Yazıldığı hâliyle ifade, `{{.Titel}}` |
| `Reason` | Neyin yanlış olduğu |
| `Suggestion` | Tipin sahip olduğu en yakın isim ya da boş |

Kontrol, register sırasında yapılan diğer kontrollerin çalıştığı yerde çalışır:
başlangıçta, dolayısıyla uygulamayı başlatan
[`collage check`](/docs/cli#collage-check)'te ve page'leri register eden bir
test'te.

`collage dev` altında, program çalışırken düzenlenen bir template yeniden
parse edilir. Ama bir sonraki yeniden başlatmaya, yani Go kodundaki bir sonraki
değişikliğe kadar tekrar kontrol edilmez.
[`collage inspect`](/docs/cli#collage-inspect), bir editörün aynı şekilde kontrol
edebilmesi için her fragment'in veri tipini ve ulaştığı tipleri yazdırır.

### Kapatmak

`WithoutTypeCheck()`, nadiren karşılaşılan yanlış bir bulgu düzeltilene kadar tek
bir fragment'in template'ini kontrolün dışında bırakır. Fragment eskisi gibi render
edilir ve `collage inspect` onu `"typeCheck": false` olarak işaretler:

```go
collage.NewFragment("post-body", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithoutTypeCheck().
	Build()
```

Bir interface dönecek şekilde tanımlanan bir handler (örneğin `collage.Load[any]`)
tipi bilinmez bırakır ve template'i dolaşılmaz. Bir fragment'in verisinin sabit bir
biçimi olmadığını bilerek söylemenin yolu budur. `collage.Value` bunun istisnasıdır:
değeri elde olduğu için, bir interface içinde tutulan değer tuttuğu tipe göre
kontrol edilir.
