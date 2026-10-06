---
description: Bir fragment'in verisini nasıl çektiğini anlatır: handler sözleşmesi, sabit veri, dependency tag'ler, 404'ler, eşzamanlılık, render context'i, veri paylaşımı, timeout'lar ve template'lerin verilerine göre nasıl kontrol edildiği.
reference: Data, FragmentBuilder.WithData, FragmentBuilder.WithTitle, FragmentBuilder.WithoutTypeCheck, DataHandler, Load, Value, RenderContext, Get, Once, Effect, ErrNotFound, ErrConflictingData, TemplateTypeError, ErrTemplateType, PanicError
---

# Data handler'lar

Data handler, bir fragment'in template'inde render edeceği veriyi almak için
çağırdığı fonksiyondur. Request'in context'ini ve render context'ini alır. Geriye
kendi Go tipindeki veriyi, bu verinin geldiği dependency tag'leri ve bir hata
döner. Bir page'in dış dünyayla konuştuğu her şey (bir veritabanı, bir CMS, bir
API) data handler'larda olur, başka hiçbir yerde olmaz.

```go
content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.DataHandler(func(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
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
| `collage.DataHandler(fn)`, `fn` `(T, []string, error)` döner | Her render'da `fn`'i çalıştırır ve dependency tag'lerini bildirir | `T` |
| `collage.Load(fn)`, `fn` `(T, error)` döner | Her render'da `fn`'i çalıştırır, tag bildirmez | `T` |
| `collage.Value(v)` | Hiçbir şey: `v` her render'da olduğu gibi verilir | `v`'nin tipi |
| `collage.Effect(fn)`, `fn` `error` döner | Her render'da `fn`'i, tanımladıkları için çalıştırır | Hiçbir şey |

`WithData` kullanmayan bir fragment, `collage.Effect` kullanan bir fragment gibi
verisiz render edilir.

Handler, kendi tipini dönen sıradan bir fonksiyondur. Onu yukarıdaki gibi yerinde
yazabilir ya da bir isim verip geçirebilirsiniz:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	// ...
}

content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.DataHandler(loadPost)).
	Build()
```

`T` fonksiyonun imzasından gelir. Page register edilirken template de bu tipe
göre kontrol edilir (bkz.
[Template'ler nasıl kontrol edilir](#how-templates-are-checked)). Bu yüzden
handler ile template'in verisi hakkındaki varsayımı birbirinden kopamaz. Bu
şekilde yazılan bir loader'ı bir test'ten ya da bir sitemap handler'ından çağırmak
da kolaydır: bunlar assert edilmesi gereken bir değer yerine bir `*Post` alır.
Constructor'ların `WithData`'nın bir biçimi değil de ayrı fonksiyonlar olmasının
nedeni, Go method'larının type parameter alamamasıdır.

v0.49.0'dan önce bir handler `WithDataHandler` ile verilir ve
`(any, []string, error)` dönerdi. Sabit veri ise `WithData(v)` ile verilirdi. İkisi
de kaldırıldı: handler'ı gerçek tipini dönecek şekilde `collage.DataHandler` ile,
değeri de `collage.Value` ile sarın.

Bir `DataHandler`'ın üç dönüş değerinin her birinin ayrı bir görevi vardır.

- **Veri:** Template'in `.` olarak render ettiği şeydir. Template için yazılmış bir
  struct, yani bir "view", template'e bir veritabanı satırı vermekten genellikle
  daha anlaşılırdır. Her fragment kendi verisini alır. Bir child, parent'ının
  verisini görmez. Hata durumunda veri atılır. Böylece bir hatanın yanında dönen
  nil bir `*Post`, template'e varmış gibi görünen bir değer olarak hiç ulaşmaz.
- **Tag'ler:** Bu verinin oluşturulduğu içerik parçalarıdır, örneğin
  `"post:hello-world"` ya da `"author:ada"`. Page'deki her fragment'ten, page'in
  kendi `WithDependency` tag'leriyle birlikte toplanır ve cache'lenen page ile
  birlikte saklanır. Böylece bir tag'i invalidate ettiğinizde, o tag'i gösteren her
  page cache'ten düşer. Veri değişen hiçbir şeye bağlı değilse `nil` dönün.
  Ayrıntılar için [Caching](/docs/caching) sayfasına bakın.
- **Hata:** nil olmayan bir hata fragment'i başarısız kılar. Bunun page için ne
  anlama geldiğine fragment'in
  [failure policy'si](/docs/fragments-and-slots#when-a-fragment-fails) karar verir.

Handler hata dönse bile tag'ler korunur. Neye bağlı olduğunu belirledikten sonra
başarısız olan bir handler, page'i neyin invalidate edeceğini yine de bildirmiştir.

Handler, strateji tanımlamayan bir page'in nasıl sunulacağını da belirler. Böyle
bir page, render ettiği herhangi bir şeyin data handler'ı varsa dynamic, yoksa
static olur. Çünkü bir handler request'i, bir cookie'yi ya da saati okuyabilir ve
fonksiyonun dışındaki hiçbir şey bunu yapıp yapmadığını bilemez. Çıktısı her
okuyucu için aynı olan bir handler'ı (dosyadan okunan bir yazı gibi),
`Static()` ya da `Incremental(ttl)`'yi kendisi belirten bir page'e koyun. Yalnızca
path'in parametrelerini ve locale'i okuyan bir handler bunu bunun yerine kendi
fragment'inde `Static()` ile söyleyebilir. O zaman o fragment'i kullanan her page
static kalır. Çıktısı her okuyucu için aynı olan ama zaman içinde değişen bir
handler (bir ölçüm gibi) bunun yerine `Shared()` der. Render'ı okuyucular arasında
paylaşılabilir ve page dynamic kalır. Ayrıntılar için
[Caching](/docs/caching#a-page-that-declares-none) sayfasına bakın.

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
kazanırdı. nil bir `Data` (`WithData(nil)` ya da nil bir fonksiyon verilmiş bir
constructor) veri yok demektir ve hiçbir şeyle çakışmaz.

### Tag'siz handler'lar

`collage.Load`, tag bildirmeyen, yalnızca veriyi ve bir hatayı dönen bir loader
için `DataHandler`'ın karşılığıdır:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithData(collage.Load(loadClock)).
	Build()
```

İkisi de loader hata döndüğünde veriyi atar. Hangisini seçeceğiniz şöyle
belirlenir:

| Veri | Yazılacak |
| --- | --- |
| Her render'da aynıysa | `collage.Value(v)` |
| Değişiklikleri cache'lenmiş bir page'e ulaşması gereken içerikten çekiliyorsa | `collage.DataHandler(fn)` |
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
  parent'ının shared data'ya koyduğu değeri ve parent'ının çözdüğü bir path
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
- **Shared data'ya `rc.Get` ve `rc.Set` üzerinden erişilir,** hiçbir zaman
  doğrudan `SharedData` map'i üzerinden erişilmez. Ayrıntılar
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
| `rc.Get(key)`, `rc.Set(key, value)` | Tek bir render'daki fragment'ler arasında paylaşılan değerler |
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
  onun `SEO` map'ine ya da `DependencyTags`'ine yazmak, framework'ün canlı state'i
  üzerinde bir data race'tir. `go test -race` bunu raporlar, production'da ise er
  geç veriyi bozar. Request'ten request'e değişen her şeyi döndüğünüz veriye ya da
  shared data'ya koyun.

## Fragment'ler arasında veri paylaşımı

Bir page'deki fragment'ler çoğu zaman aynı şeye ihtiyaç duyar. Yazı page'inin
içeriği, `<head>`'i ve "bu yazarın diğer yazıları" kutusu, hepsi yazının kendisine
ihtiyaç duyar.

### rc.Set ve collage.Get

`rc.Set(key, value)` bir değeri render'ın geri kalanı için saklar.
`collage.Get[T](rc, key)` ise o değeri saklandığı tiple geri okur:

```go
// in the parent's handler
rc.Set("post", post)

// in a child's handler, which starts after the parent's has returned
post, ok := collage.Get[*Post](rc, "post")
if !ok {
	return nil, nil, errors.New("more-by-author: no post in shared data")
}
```

Key altında hiçbir şey saklanmamışsa `ok` false olur. Saklanan değer bir `*Post`
değilse de false olur. Çağıran taraf için ikisi de aynı anlama gelir: istediği değer
orada yoktur. Key'ler uygulamanızın kendi namespace'idir. Çakışmayacak isimler
seçin.

`rc.Get` ve `rc.Set` render'ın lock'unu alır. Eşzamanlı çalışan sibling'lere karşı
güvenli olmalarının nedeni budur. Çıplak `rc.SharedData` map'i ise güvenli değildir.

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
	post, err := collage.Once(rc, "post:"+slug, func(ctx context.Context) (*Post, error) {
		return store.Post(ctx, slug)
	})
	if err != nil {
		return Author{}, nil, err
	}
	return post.Author, []string{"post:" + slug, "author:" + post.Author.ID}, nil
}
```

İlk çağıran veriyi çeker. Diğerleri onu bekler ve onun ürettiği sonucu alır. Akla
ilk gelen alternatif şudur: `rc.Get` ile bakmak, bulunamazsa veriyi çekmek, sonra
`rc.Set` ile yazmak. Bu yaklaşımda kontrol ile yazma arasında bir boşluk kalır. İki
sibling de bu boşluğa düşer ve ikisi de veriyi çeker.

- **Hata da bir sonuçtur.** Bekleyen herkes aynı hatayı alır. Çekme işlemi her
  fragment için yeniden denenmez. Aksi hâlde tek bir yavaş hata birkaç hataya
  dönüşürdü.
- **Bir render boyunca geçerlidir.** Ayarlamanız ya da temizlemeniz gereken hiçbir
  şey yoktur.
- **Kendi context'i sona eren bekleyen taraf beklemeyi bırakır** ve o hatayı
  döner.
- **Bir key, bir tip.** Aynı key'i iki farklı tiple istemek sessizce boş bir değer
  dönmez, `ErrOnceTypeMismatch` verir.

### collage.Cached

`Once` işi bir page içinde paylaştırır. `collage.Cached` ise işi page'ler ve
request'ler arasında paylaştırır. Böylece aynı yazarın otuz yazısı, yazarı yalnızca
bir kez çeker.

```go
author, err := collage.Cached(rc, "author:"+id, time.Hour, []string{"author:" + id},
	func(ctx context.Context) (Author, error) { return api.Author(ctx, id) })
```

Verdiğiniz tag'ler page'in kendi tag'lerine eklenir. Bunlardan birini invalidate
ettiğinizde hem saklanan değer hem de o değerden oluşturulmuş bütün cache'lenmiş
page'ler düşer. Render'lar arasında hiçbir şeyin tutulmadığı durumlarda (cache
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
		post, err := collage.Once(rc, "post:"+rc.Param("slug"), func(ctx context.Context) (*Post, error) {
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
yerine sıradan bir handler'dan dönün ya da page üzerinde `WithDependency` ile
ekleyin.

Program başlarken belli olan bir title (layout'taki site adı gibi) hiç handler
gerektirmez. Fragment üzerindeki `WithTitle(s)` onu tanımlar ve strateji
tanımlamayan bir page'i static bırakır. Ayrıntılar için
[Head ve SEO](/docs/head-and-seo) sayfasına bakın.

## Timeout'lar ve context

Her data handler bir deadline altında çalışır. Bu deadline fragment'in
`WithTimeout(d)` değeridir. Timeout ayarlamayan bir fragment için ise
`Config.Template.Timeout` kullanılır, varsayılan değeri beş saniyedir.

Deadline, handler'ın aldığı `ctx` üzerindedir. **Handler'ı değil, context'i
sınırlar.** Framework bir handler'ı deadline'ı geldiğinde bırakmaz, dönmesini
bekler. Çünkü bir goroutine dışarıdan durdurulamaz. Hiç dönmeyen handler'ları
bırakmak, her request için bir goroutine'in sonsuza kadar sızması anlamına gelirdi.
Bu yüzden context'ini yok sayan bir handler deadline'ını aşarak çalışabilir ve page
onu bekler.

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
edilir. Olmasa da olur türündeki her şey için kısa bir timeout'u bir fallback ile
birlikte kullanın. Böylece page, yavaş bir servisi sizin seçtiğiniz noktada
beklemeyi bırakır.

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
ise hiç ortaya çıkmaz. v0.49.0'dan beri her fragment'in verisi tipli bir
constructor'dan gelir. Bu yüzden register işlemi her template'in hangi Go tipiyle
çalışacağını bilir ve template'i bu tipe göre dolaşır. `RegisterPage`,
template'leri verisine uymayan bir page'i reddeder. Hata page'i, fragment'i,
dosyayı, satırı ve sütunu verir:

```text
collage: page "post": fragment "post-body" (post.html:1:6): {{.Titel}}: type blog.Post has no field or method Titel (did you mean Title?)
```

Yalnızca kesin olarak başarısız olacak şeyleri bildirir: her bulgu, render ona
ulaştığında `text/template`'in başarısız olacağı bir ifadedir. Artık başlangıçta
başarısız olan bir uygulamanın template'i, render edildiğinde zaten başarısız
olacaktı. Kontrol bunu yalnızca daha erken bildirir, kendinden bir kural eklemez.

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
  Pointer'lar izlenir, gömülü struct'lardan yükseltilen alanlar da sayılır. Yakın
  bir isim varsa, en yakın export edilmiş isim önerilir.
- **Adreslenebilir olmayan bir değer üzerinden ulaşılan, pointer receiver'lı bir
  method.** `text/template` böyle bir method'u değerin adresi üzerinden çağırır.
  Template'e değer olarak verilen verinin adresi yoktur; alanlarının ve bir map'in
  elemanlarının da yoktur. Bir pointer'ın gösterdiği şey ve bir slice'ın
  elemanları adreslenebilirdir. `URL`, `*Post` üzerinde tanımlıysa `{{.URL}}` bir
  `Post` üzerinde başarısız olur, bir `*Post` üzerinde çalışır. Çözüm, template'e
  pointer'ı vermektir.
- **İsimli bir string tipiyle anahtarlanan bir map** (`map[Slug]Post`): `.key`
  anahtarı düz bir `string` olarak arar ve böyle bir map bunu kabul etmez.
  `string` ile anahtarlanan bir map sorun değildir. Map'te olmayan bir anahtar da
  hata değildir: zinciri sessizce bitirir, bu yüzden `{{.Meta.absent.Name}}` hiçbir
  şey render etmez. Yalnızca map'in eleman tipini bilen kontrol, o eleman tipinde
  olmayan bir ismi yine de bildirir.
- **Yanlış biçimdeki çağrılar:** yanlış sayıda argüman verilen bir method,
  template fonksiyonu ya da built-in; argüman verilen bir alan ya da map anahtarı;
  bir değer ve bir hatadan fazlasını dönen bir method ya da fonksiyon; fonksiyon
  olmayan bir şey üzerinde `call`.
- **Yanlış türde `range`, `len` ve `index`:** bir string ya da struct üzerinde
  range, iki değişkenle bir tamsayı üzerinde range, bir struct'ın `len`'i, bir
  struct'a `index`.

Render'ın hiç ulaşamayacağı şeyler bildirilmez. Koşulu bir literal olan bir
`{{if}}` ya da `{{with}}` (`true`, `false`, `0`, `1`, `""`, `"x"` ya da bunlardan
birinin `not`'u), literal'in dışarıda bıraktığı tarafı hiç çalıştırmaz. `{{and}}`
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
  çağrıları ve bunların argüman sayıları kontrol edilir.
- **Verinin içindeki nil pointer'lar.** `Author` nil ise `{{.Author.Name}}` render
  edilirken başarısız olur. Onun nil olup olmadığı tipin değil, verinin işidir.
- **Argüman tipleri.** Yalnızca argüman sayısı kontrol edilir. `text/template`
  bazı argümanları kendisi dönüştürür ve onu tahmin etmeye çalışmak yanlış
  alarmların kaynağı olurdu.

### Hatayı okumak

`RegisterPage`, page'in kendi fragment'lerindeki bütün bulguları bir kerede,
`errors.Join` ile birleştirip dosya, satır ve sütuna göre sıralayarak döner.
Böylece on hata, on değil tek bir yeniden başlatma demektir. Page'in not-found
page'i, page'in kendisi geçtikten sonra kontrol edilir. Error page'i ise ikisi de
geçtikten sonra kontrol edilir. Bunların bulguları bir
`collage: page "post" not-found page: …` (ya da `error page: …`) hatasıyla sarılı
olarak, page'in kendi bulguları
düzeltildikten sonraki başlangıçta gelir.

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
| `Expr` | Yazıldığı haliyle ifade, `{{.Titel}}` |
| `Reason` | Neyin yanlış olduğu |
| `Suggestion` | Tipin sahip olduğu en yakın isim ya da boş |

Kontrol, register sırasında yapılan diğer kontrollerin çalıştığı yerde çalışır:
başlangıçta, dolayısıyla uygulamayı başlatan
[`collage check`](/docs/cli#collage-check)'te ve page'leri register eden bir
test'te. `collage dev` altında, program çalışırken
düzenlenen bir template yeniden parse edilir ama bir sonraki yeniden başlatmaya,
yani Go kodundaki bir sonraki değişikliğe kadar tekrar kontrol edilmez.
[`collage inspect`](/docs/cli#collage-inspect), bir editörün aynı şekilde kontrol
edebilmesi için her fragment'in veri tipini ve ulaştığı tipleri yazdırır.

### Kapatmak

`WithoutTypeCheck()`, nadiren karşılaşılan yanlış bir bulgu düzeltilene kadar tek
bir fragment'in template'ini kontrolün dışında bırakır. Fragment eskisi gibi render
edilir ve `collage inspect` onu `"typeCheck": false` olarak işaretler:

```go
collage.NewFragment("post-body", "pages/post.html").
	WithData(collage.DataHandler(loadPost)).
	WithoutTypeCheck().
	Build()
```

Bir interface dönecek şekilde tanımlanan bir handler (örneğin `collage.Load[any]`)
tipi bilinmez bırakır ve template'i dolaşılmaz. Bir fragment'in verisinin sabit bir
biçimi olmadığını bilerek söylemenin yolu budur. `collage.Value` bunun istisnasıdır:
değeri elde olduğu için, bir interface içinde tutulan değer tuttuğu tipe göre
kontrol edilir.
