---
description: Bir fragment'in verisini nasıl çektiğini anlatır: tipli handler sözleşmesi, dependency tag'ler, 404'ler, eşzamanlılık, render context'i, veri paylaşımı ve timeout'lar.
reference: DataHandler, Data, Load, RenderContext, Get, Once, Effect, ErrNotFound, PanicError
---

# Data handler'lar

Data handler, bir fragment'in template'inde render edeceği veriyi almak için
çağırdığı fonksiyondur. Request'in context'ini ve render context'ini alır. Geriye
veriyi, bu verinin geldiği dependency tag'leri ve bir hata döner. Bir page'in dış
dünyayla konuştuğu her şey (bir veritabanı, bir CMS, bir API) data handler'larda
olur, başka hiçbir yerde olmaz.

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return Post{}, nil, err
	}
	return post, []string{"post:" + post.Slug}, nil
}

content := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	Required().
	Build()
```

## Sözleşme

Handler'ı kendi tipinizle yazın ve `collage.DataHandler` ile adapte edin:

```go
func(ctx context.Context, rc *collage.RenderContext) (T, []string, error)
```

`collage.DataHandler`, `T` üzerinde generic'tir. Bu yüzden tek bir adapter her view
tipi için çalışır ve handler'ınızın döndüğü değer, template'in `.` olarak aldığı
değerin ta kendisidir. Kodunuzun hiçbir yerinde tipsiz bir değere ya da type
assertion'a gerek kalmaz. (Builder üzerinde bir method değil de bir fonksiyon
olmasının nedeni, Go method'larının type parameter alamamasıdır.)

Üç dönüş değerinin her birinin ayrı bir görevi vardır.

- **Veri:** Template'in render ettiği şeydir. Template için yazılmış bir struct,
  yani bir "view", template'e bir veritabanı satırı vermekten genellikle daha
  anlaşılırdır. Her fragment kendi verisini alır. Bir child, parent'ının verisini
  görmez.
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

Hata durumunda `collage.DataHandler` veriyi iletmez, atar. Bu, pointer tipleri
için önemlidir. Aksi hâlde bir hatayla birlikte dönen nil bir `*Post`, template'e
nil pointer tutan ama kendisi nil olmayan bir değer olarak ulaşırdı.

### Daha kısa adapter'lar: Data ve Load

Her fragment'in üç dönüş değerine de ihtiyacı yoktur. Tag bildirmeyen durumlar
için iki adapter vardır.

`collage.Data` değerin kendisini alır. Program başlarken belli olan veriler için
kullanılır: bir link listesi, bir başlık, bir site adı gibi. Yazmanız gereken bir
fonksiyon yoktur:

```go
type homeView struct {
	Links []link
}

content := collage.NewFragment("home-content", "pages/home.html").
	WithDataHandler(collage.Data(homeView{Links: links})).
	Build()
```

`collage.Load` ise tag'ler olmadan yalnızca veriyi ve bir hatayı dönen bir handler
alır:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithDataHandler(collage.Load(loadClock)).
	Build()
```

İkisi de `collage.DataHandler` gibi generic'tir. Template yine kendi tipinizi alır
ve `collage.Load` da hata durumunda veriyi aynı şekilde atar. Hangisini
seçeceğiniz şöyle belirlenir:

| Veri | Adapter |
| --- | --- |
| Her render'da aynıysa | `collage.Data(v)` |
| Dışarıdan çekiliyorsa ve page cache'lenmiyorsa ya da veri hiç değişmiyorsa | `collage.Load(fn)` |
| Dışarıdan çekiliyorsa ve değiştiğinde cache'lenmiş page düşürülmeliyse | `collage.DataHandler(fn)` |

Birinden diğerine geçmek baştan yazmak değil, yalnızca bir signature
değişikliğidir. Bir page cache'lenmeye başladığında, `Load` handler'ına tag'ler
eklenir ve handler bir `DataHandler` handler'ına dönüşür.

### Not found bir hata değildir

Eksik bir kayıt ile bozuk bir veritabanı farklı hatalardır. Okuyucu da her biri için
farklı bir cevap almalıdır: ilki için 404, ikincisi için 500. Hangisi olduğunu
`collage.ErrNotFound`'u wrap ederek belirtin:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (Post, []string, error) {
	slug := rc.Param("slug")
	post, err := store.Post(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, nil, fmt.Errorf("blog: no post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return Post{}, nil, fmt.Errorf("blog: load post %q: %w", slug, err)
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
| `rc.Page` | Render edilen page. **Yalnızca okunur** |
| `rc.Get(key)`, `rc.Set(key, value)` | Tek bir render'daki fragment'ler arasında paylaşılan değerler |
| `rc.Context()` | Render context'inin taşıdığı context |
| `rc.HoistTitle`, `rc.HoistMeta`, `rc.HoistProperty`, `rc.HoistLink`, `rc.HoistAlternate`, `rc.HoistStylesheet`, `rc.Hoist` | Page'in `<head>`'i için tanımlar. Bkz. [Head ve SEO](/docs/head-and-seo) |
| `rc.Asset(path)` | Mount edilmiş bir dosyanın content-addressed URL'si. Bkz. [Static asset'ler](/docs/assets) |

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
post, ok := collage.Get[Post](rc, "post")
if !ok {
	return moreView{}, nil, errors.New("more-by-author: no post in shared data")
}
```

Key altında hiçbir şey saklanmamışsa `ok` false olur. Saklanan değer bir `Post`
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
	post, err := collage.Once(rc, "post:"+slug, func(ctx context.Context) (Post, error) {
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

## Hiçbir şey render etmeyen handler'lar

Bazı fragment'ler markup render etmek için değil, page için bir şeyler tanımlamak
için vardır: bir title, bir canonical link, structured data gibi. `collage.Effect`,
yalnızca hata dönen bir handler'ı adapte eder:

```go
seo := collage.NewFragment("post-seo", "fragments/empty.html").
	WithDataHandler(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		post, err := collage.Once(rc, "post:"+rc.Param("slug"), func(ctx context.Context) (Post, error) {
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
yerine bir `collage.DataHandler`'dan dönün ya da page üzerinde `WithDependency` ile
ekleyin.

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
func loadWeather(ctx context.Context, rc *collage.RenderContext) (Weather, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherURL(rc.Locale), nil)
	if err != nil {
		return Weather{}, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Weather{}, nil, err // context.DeadlineExceeded when the timeout passed
	}
	defer res.Body.Close()

	var weather Weather
	if err := json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return Weather{}, nil, err
	}
	return weather, nil, nil
}
```

Süre dolduğunda context'inin hatasını dönen bir handler, diğer handler'lar gibi
başarısız olur. Bu hata yine de `errors.Is(err, context.DeadlineExceeded)` ile
eşleşir. Deadline ancak handler onu hata olarak dönerse bir hatadır. `ctx`'i yok
sayıp geç de olsa `nil` dönen bir handler başarılı olmuştur ve verisi render
edilir. Olmasa da olur türündeki her şey için kısa bir timeout'u bir fallback ile
birlikte kullanın. Böylece page, yavaş bir servisi sizin seçtiğiniz noktada
beklemeyi bırakır.

## Panic'ler

Bir data handler'daki panic process'i çökertmez. Panic recover edilir ve panic
değerini ve stack'i taşıyan bir `*collage.PanicError`'a dönüştürülür. Fragment da
diğer her hatada olduğu gibi bu hatayla başarısız olur. Development'ta error page
stack'i gösterir. Kendi kodunuzda bu hataya `errors.As` ile ulaşabilirsiniz.
