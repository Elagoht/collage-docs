---
description: Bir fragment'in verisini nasıl çektiği — tipli handler sözleşmesi, bağımlılık etiketleri, 404'ler, eşzamanlılık, render bağlamı, veri paylaşımı ve zaman aşımları.
---

# Data handler'lar

Data handler, bir fragment'in şablonunun render edeceği şeyi almak için çağırdığı
fonksiyondur. İsteğin context'ini ve render bağlamını alır; veriyi, bu verinin
geldiği bağımlılık etiketlerini ve bir hatayı döndürür. Bir sayfanın dış dünyayla —
bir veritabanıyla, bir CMS'le, bir API'yle — konuşan her şeyi data handler'larda
olur, başka hiçbir yerde değil.

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

Handler'ı kendi tipinize göre yazın ve `collage.DataHandler` ile uyarlayın:

```go
func(ctx context.Context, rc *collage.RenderContext) (T, []string, error)
```

`collage.DataHandler`, `T` üzerinde generic'tir; dolayısıyla tek bir uyarlayıcı her
görünüm tipine hizmet eder ve handler'ınızın döndürdüğü değer, şablonun `.` olarak
aldığı değerin ta kendisidir. Kodunuzda tipsiz bir değere ya da tip dönüşümüne
(type assertion) gerek kalmaz. (Builder üzerinde bir metot değil de fonksiyon
olmasının nedeni, Go metotlarının tip parametresi alamamasıdır.)

Üç sonucun her birinin bir görevi var.

- **Veri** — şablonun render ettiği her neyse. Şablon için yazılmış bir struct,
  yani bir "görünüm", şablona bir veritabanı satırı vermekten genellikle daha
  açıktır. Her fragment kendi verisini alır; bir çocuk, üst fragment'inkini görmez.
- **Etiketler** — bu verinin kurulduğu içerik parçaları, örneğin
  `"post:hello-world"` ya da `"author:ada"`. Sayfadaki her fragment'ten, sayfanın
  kendi `WithDependency` etiketleriyle birlikte toplanır ve önbellekteki sayfayla
  birlikte saklanır; böylece bir etiketi geçersiz kılmak, onu gösteren her sayfayı
  düşürür. Veri değişen hiçbir şeye bağlı değilse `nil` döndürün. Bkz.
  [Önbellek](/docs/caching).
- **Hata** — nil olmayan bir hata fragment'i başarısız kılar; bunun sayfa için ne
  anlama geldiğine fragment'in
  [hata politikası](/docs/fragments-and-slots#when-a-fragment-fails) karar verir.

Etiketler, handler hata döndürse bile tutulur. Neye bağlı olduğunu çıkarıp sonra
başarısız olan bir handler, sayfayı neyin geçersiz kılacağını yine de söylemiştir.

Hata durumunda `collage.DataHandler` veriyi aktarmak yerine atar. Bu, pointer
tipleri için önemlidir: bir hatayla birlikte döndürülen nil bir `*Post`, aksi hâlde
şablona nil bir pointer tutan, nil olmayan bir değer olarak ulaşırdı.

### Daha kısa uyarlayıcılar: Data ve Load

Her fragment'in üç sonuca da ihtiyacı yoktur. Etiket bildirmeyenler için iki
uyarlayıcı vardır.

`collage.Data` değerin kendisini alır; program başlarken sabitlenen veriler için
— bir bağlantı listesi, bir başlık, bir site adı. Yazılacak bir fonksiyon yoktur:

```go
type homeView struct {
	Links []link
}

content := collage.NewFragment("home-content", "pages/home.html").
	WithDataHandler(collage.Data(homeView{Links: links})).
	Build()
```

`collage.Load` ise etiketler olmadan, veriyi ve bir hatayı döndüren bir handler
alır:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithDataHandler(collage.Load(loadClock)).
	Build()
```

İkisi de `collage.DataHandler` gibi generic'tir; şablon yine kendi tipinizi alır
ve `collage.Load` hata durumunda veriyi aynı şekilde atar. Hangisini seçmeli:

| Veri | Uyarlayıcı |
| --- | --- |
| Her render'da aynı | `collage.Data(v)` |
| Çekiliyor; sayfa önbelleğe alınmıyor ya da veri hiç değişmiyor | `collage.Load(fn)` |
| Çekiliyor; değiştiğinde önbellekteki sayfa düşürülmeli | `collage.DataHandler(fn)` |

Birinden diğerine geçmek yeniden yazmak değil, bir imza değişikliğidir: bir sayfa
önbelleğe alınmaya başladığında `Load` handler'ı etiketlerini kazanır ve bir
`DataHandler` handler'ı olur.

### Bulunamadı bir hata değildir

Eksik bir kayıt ile bozuk bir veritabanı farklı hatalardır ve okuyucunun her biri
için farklı bir yanıt alması gerekir: ilki için 404, ikincisi için 500. Hangisi
olduğunu `collage.ErrNotFound`'u sarmalayarak söyleyin:

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

`Required()` bir fragment'te, `collage.ErrNotFound`'u sarmalayan bir hata, sayfanın
bulunamadı sayfasını 404 ile render eder; başka her hata, hata sayfasını 500 ile
render eder. Bkz. [Sayfalar ve layout'lar](/docs/pages-and-layouts#not-found-and-error-pages).

İkisi de kolayca gözden kaçan iki koşul var:

- **En üste kadar `%w` ile sarmalayın.** Framework `errors.Is` ile denetler;
  dolayısıyla deponuz ile handler'ın dönüşü arasında herhangi bir yerdeki bir `%v`,
  404'ü 500'e çevirir.
- **Onu yalnızca zorunlu bir fragment 404'e çevirir.** İsteğe bağlı bir fragment'te
  bu sıradan bir hatadır — fragment hiçbir şey ya da yedeğini render eder — çünkü
  sayfa o olmadan da göstermeye değerdir.

## Handler'lar ne zaman çalışır

Bir sayfanın data handler'ları tek tek çalışmaz. **Bir fragment şablonunu render
etmeden önce, slot'larındaki her fragment'in data handler'ı aynı anda başlar.** Bir
yazı, bir kenar çubuğu ve bir yorum listesinden oluşan sayfa, üçünün toplamını değil,
en yavaşını bekler.

Garanti edilen sıra:

- **Üst fragment'in handler'ı, çocuklarınınkiler başlamadan biter.** Bir çocuk,
  üst fragment'in paylaşılan veriye koyduğunu ve üst fragment'in çözdüğü bir yol
  parametresini okuyabilir.
- **Kardeşler aynı anda çalışır**, her biri kendi goroutine'inde, belirli bir sırası
  olmadan.
- **Şablonlar yine de tek tek, ağaç sırasıyla render edilir.** Çıktı bir istekten
  diğerine aynıdır ve sıraya bağlı her şeye — örneğin `<head>`'de hangi başlığın
  kazanacağına — hangi handler'ın önce bittiği değil, ağaç karar verir.

Bunun yazdığınız kod için üç sonucu var.

- **Handler'lar kardeşleriyle eşzamanlı çalışmaya karşı güvenli olmalıdır.**
  Paylaştıkları her şey — bir map, bir sayaç, goroutine açısından güvenli olmayan bir
  istemci — Go'nun başka herhangi bir yerinde gerektireceği özeni gerektirir.
- **Paylaşılan veri `rc.Get` ve `rc.Set` üzerinden geçer,** asla doğrudan
  `SharedData` map'i üzerinden değil — bkz. [aşağısı](#sharing-data-between-fragments).
- **Şablonun render etmediği bir slot'taki fragment de yine başlar.** Context'i,
  şablon biter bitmez iptal edilir ve hatası yok sayılır. Pahalı ve genellikle
  atlanan bir handler, şablondaki bir koşuldan başka bir şeyin arkasına konmalıdır.

## Render bağlamı

`rc *collage.RenderContext`, bir handler'ın render edilen istek hakkında bildiği her
şeydir.

| Üye | Size verdiği |
| --- | --- |
| `rc.Request` | Yanıtlanan `*http.Request`. Statik dışa aktarmada, sayfanın yolu için yapay bir `GET` |
| `rc.Locale` | URL'nin çözüldüğü locale |
| `rc.Param(name)`, `rc.PathParams` | Route'un `{name}` yer tutucularının yakaladıkları |
| `rc.Page` | Render edilen sayfa — **salt okunur** |
| `rc.Get(key)`, `rc.Set(key, value)` | Tek bir render'ın fragment'leri arasında paylaşılan değerler |
| `rc.Context()` | Render bağlamının taşıdığı context |
| `rc.HoistTitle`, `rc.HoistMeta`, `rc.HoistProperty`, `rc.HoistLink`, `rc.HoistAlternate`, `rc.HoistStylesheet`, `rc.Hoist` | Sayfanın `<head>`'i için bildirimler — bkz. [Head ve SEO](/docs/head-and-seo) |
| `rc.Asset(path)` | Mount edilmiş bir dosyanın içerik adresli URL'si — bkz. [Statik dosyalar](/docs/assets) |

Bir handler'ın içinde `rc.Context()`, fragment'in zaman aşımını taşıyan, `ctx`
argümanıyla aynı context'tir. `ctx`'i kullanın; zaten elinizde olan odur.

Render bağlamının kendisiyle ilgili iki kural.

- **Onu saklamayın.** Tek bir render'a aittir. Onu handler döndükten sonra da — bir
  goroutine'de, bir önbellekte, bir struct'ta — tutmak, bitmiş bir isteği tutmaktır.
- **`rc.Page`'e yazmayın.** O, kayıtlı tek `*collage.Page`'dir ve o sayfayı aynı anda
  render eden her istek tarafından paylaşılır. Bir handler'dan onun `SEO` map'ine ya
  da `DependencyTags`'ine yazmak, framework'ün canlı durumu üzerinde bir veri
  yarışıdır; `go test -race` bunu raporlar, production ise er geç bozar. İstekten
  isteğe değişen her şey, döndürdüğünüz veriye ya da paylaşılan veriye gider.

## Fragment'ler arasında veri paylaşmak

Bir sayfadaki fragment'ler çoğu zaman aynı şeye ihtiyaç duyar. Yazı sayfasının
içeriği, `<head>`'i ve "bu yazarın diğer yazıları" kutusu, hepsi yazıyı ister.

### rc.Set ve collage.Get

`rc.Set(key, value)` bir değeri render'ın geri kalanı için saklar,
`collage.Get[T](rc, key)` ise onu saklandığı tip olarak geri okur:

```go
// in the parent's handler
rc.Set("post", post)

// in a child's handler, which starts after the parent's has returned
post, ok := collage.Get[Post](rc, "post")
if !ok {
	return moreView{}, nil, errors.New("more-by-author: no post in shared data")
}
```

Anahtar altında hiçbir şey saklanmamışsa `ok` false olur; saklanan şey bir `Post`
değilse de öyle — çağıran için ikisi de istediği değerin orada olmadığı anlamına gelir.
Anahtarlar uygulamanızın kendi ad alanıdır; çakışmayacak adlar seçin.

`rc.Get` ve `rc.Set` render'ın kilidini alır; eşzamanlı kardeşlere karşı
güvenli olmalarının, çıplak `rc.SharedData` map'inin ise güvenli olmamasının nedeni
budur.

Bu, üst fragment'ten çocuğa doğru çalışır, çünkü üst fragment'in handler'ı önce
biter. Kardeşler arasında çalışmaz: aynı anda çalıştıkları için biri, diğerinin
henüz bir şey ayarlamış olduğuna güvenemez.

### collage.Once

Kardeşler için — ya da her biri aynı veri çekme işlemine ihtiyaç duyabilecek
herhangi fragment'ler için — `collage.Once` kullanın. Bir anahtar için bir veri
çekme işlemini render başına en fazla bir kez çalıştırır ve sonucu isteyen her
fragment'e verir:

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

İlk çağıran veriyi çeker; geri kalanlar onu bekler ve onun ürettiğini alır. Akla
gelen alternatifte — `rc.Get`, yoksa çek, `rc.Set` — denetim ile yazma arasında bir
boşluk vardır; iki kardeş de bu boşluğa düşer ve ikisi de veriyi çeker.

- **Hata da bir sonuçtur.** Bekleyen herkes onu alır; veri çekme her fragment için
  yeniden denenmez — tek bir yavaş hata böyle birkaç hataya dönüşürdü.
- **Bir render sürer.** Yapılandırılacak da çıkarılacak da hiçbir şey yoktur.
- **Kendi context'i sona eren bir bekleyen beklemeyi bırakır** ve o hatayı döndürür.
- **Bir anahtar, bir tip.** Bir anahtarı iki farklı tip olarak istemek sessizce boş
  bir değer değil, `ErrOnceTypeMismatch`'tir.

### collage.Cached

`Once` işi bir sayfanın içinde paylaştırır. `collage.Cached` ise onu sayfalar ve
istekler arasında paylaştırır: aynı yazarın otuz yazısı yazarı bir kez çeker.

```go
author, err := collage.Cached(rc, "author:"+id, time.Hour, []string{"author:" + id},
	func(ctx context.Context) (Author, error) { return api.Author(ctx, id) })
```

Etiketleri sayfanın kendi etiketlerine eklenir ve birini geçersiz kılmak hem
saklanan değeri hem de ondan kurulmuş, önbellekteki her sayfayı düşürür. Render'lar
arasında hiçbir şeyin tutulmadığı yerlerde — önbellek kapalıyken, geliştirmede, bir
önizlemede — tam olarak `Once` gibi davranır. Ayrıntılar
[Önbellek](/docs/caching#caching-data-across-pages) sayfasında.

## Hiçbir şey render etmeyen handler'lar

Bazı fragment'ler işaretleme render etmek için değil, sayfa için bir şeyler
bildirmek için vardır: bir başlık, bir canonical bağlantı, yapılandırılmış veri.
`collage.Effect`, yalnızca hata döndüren bir handler'ı uyarlar:

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

Fragment'in şablonu hiç veri almaz ve handler hiç etiket bildirmez. Bildirdiği şey
değişen içerikten geliyorsa ve sayfa önbelleğe alınıyorsa, o içeriğin etiketlerinin
yine de sayfaya ulaşması gerekir: bunun yerine onları bir `collage.DataHandler`'dan
döndürün ya da sayfada `WithDependency` ile ekleyin.

## Zaman aşımları ve context

Her data handler bir son tarih altında çalışır: fragment'in `WithTimeout(d)`'si ya
da hiç ayarlamayan bir fragment için `Config.Template.Timeout` — varsayılan olarak
beş saniye.

Son tarih, handler'ın aldığı `ctx` üzerindedir. **Handler'ı değil, context'i
sınırlar.** Framework bir handler'ı son tarihinde bırakıp gitmez; dönmesini bekler,
çünkü bir goroutine dışarıdan durdurulamaz ve hiç dönmeyen handler'ları bırakıp
gitmek, her istek için bir goroutine'i sonsuza dek sızdırırdı. Dolayısıyla context'ini
yok sayan bir handler son tarihini aşarak çalışabilir ve sayfa onu bekler.

Buna uymak tek bir alışkanlıktır: bekleyebilecek her şeye `ctx`'i geçirin.

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

Süre dolduğunda context'inin hatasını döndüren bir handler, diğerleri gibi başarısız
olur; hata yine de `errors.Is(err, context.DeadlineExceeded)` ile eşleşir. Son tarih
yalnızca handler öyle derse bir hatadır: `ctx`'i yok sayıp geç kalarak `nil`
döndüren bir handler başarılı olmuştur ve verisi render edilir. Olsa iyi olan her şeyde
kısa bir zaman aşımını bir yedekle eşleştirin; sayfa, yavaş bir servisi sizin
seçtiğiniz noktada beklemeyi bırakır.

## Panic'ler

Bir data handler'daki panic süreci çökertmez. Panic değerini ve yığın izini taşıyan
bir `*collage.PanicError`'a dönüştürülerek kurtarılır ve fragment, diğer her hatada
olduğu gibi onunla başarısız olur. Geliştirmede hata sayfası yığın izini gösterir;
kendi kodunuzda ona `errors.As` ile ulaşın.
