---
description: Fragment'ler, sundukları slot'lar, biri başarısız olduğunda ne olduğu ve render sırasında içerikten doldurulan slot'lar.
---

# Fragment'ler ve slot'lar

**Fragment**, bir sayfanın yapıldığı birimdir: bir şablon, şablonun render ettiğini
getiren isteğe bağlı bir data handler ve başka fragment'lerin yerleştiği, adı konmuş
**slot**'lar. Bir layout bir fragment'tir. Bir sayfanın içeriği, bir kenar çubuğu,
bir yazar kartı, bir yorum listesi de öyle. Bir sayfa, kökünde layout bulunan bir
fragment ağacıdır.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(collage.DataHandler(loadAuthor)).
	WithTimeout(time.Second).
	WithFallback(anonymousAuthor).
	Build()
```

## Fragment kurmak

`collage.NewFragment(name, templatePath)` bir builder başlatır; `Build()` ise
`*collage.Fragment`'i döndürür.

- **Ad**, fragment'in raporlarda nasıl anıldığıdır: hata mesajlarında, geliştirme
  hata panelinde, render metadata'sında ve trace'lerde. Sayfanın hangi parçası
  olduğunu söylesin — `"list"` değil, `"post-comments"`.
- **Şablon yolu**, uzantısıyla birlikte şablon köküne görecelidir:
  `"fragments/author.html"`, `templates/fragments/author.html` demektir. Şablon
  motorunun yüklemediği bir yol, fragment'i içeren sayfa kaydedildiğinde
  `ErrTemplateNotFound` olur — v0.11.0'dan itibaren sayfanın `WithFragmentPath` ile
  kendi URL'sinde açtığı bir fragment de buna dahildir. Yalnızca bir
  [slot resolver](#slots-filled-per-render)'ın döndürdüğü fragment daha sonra, render
  edildiğinde denetlenir. Bkz. [Şablonlar](/docs/templates).

Geri kalan her şey isteğe bağlıdır:

| Metot | Neyi ayarlar |
| --- | --- |
| `WithDataHandler(h)` | Şablonun verisini getiren fonksiyon — bkz. [Data handler'lar](/docs/data-handlers) |
| `WithSlot(name, required, allowMultiple)` | Bir slot bildirir |
| `WithSlotFragment(slot, child)` | Bir alt fragment'i bir slot'a bağlar |
| `WithSlotResolver(slot, resolve)` | Bunun yerine slot'u her render'da doldurur |
| `Required()` | Bu fragment'in başarısızlığı sayfayı başarısız kılar |
| `WithFallback(f)` | Bu fragment başarısız olduğunda ne render edileceği |
| `WithTimeout(d)` | Data handler'ının ne kadar sürebileceği |

Sayfa builder'ı gibi fragment builder'ı da hataları zinciri durdurmak yerine kayda
geçirir ve `BuildErr()` bunları döndürür. Kaydettikleri, kurduğu fragment'in üzerinde
kalır; böylece her biri — iki kez bildirilmiş bir slot, hiç bildirilmemiş bir slot'a
bağlanmış bir alt fragment, negatif bir zaman aşımı — ağacında o fragment bulunan her
sayfayı, biri `BuildErr()`'ü çağırmış olsun ya da olmasın, kayıt sırasında durdurur.
Hatayı `RegisterPage`'de değil, ona yol açan satırda almak istediğinizde onu çağırın;
bkz. [Sayfalar ve layout'lar](/docs/pages-and-layouts#building-a-page).

Data handler'ı olmayan bir fragment, şablonunu veri olmadan render eder. Bu, hiç
değişmeyen işaretlemeler — bir altbilgi, statik bir duyuru — ve tek işi slot'ları
düzenlemek olan bir layout için doğrudur.

## Slot'lar

Slot, bir fragment'in şablonunda adı konmuş bir konumdur. Fragment üzerinde bildirilir
ve şablona `{{slot "name"}}` ile yazılır:

```go
post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	WithSlot("author", true, false).
	WithSlot("related", false, true).
	WithSlotFragment("author", author).
	WithSlotFragment("related", relatedPosts).
	WithSlotFragment("related", popularPosts).
	Build()
```

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
<aside>{{slot "related"}}</aside>
```

`WithSlot(name, required, allowMultiple)` iki flag alır.

- **`required`** — slot'ta bir şey olmalıdır. Hiçbir şey bağlanmamış zorunlu bir slot
  kayıt sırasında reddedilir (`ErrRequiredSlotUnfilled`) ve şablon çalışmadan önce
  yeniden denetlenir; böylece şablon onu hiç istemese bile yakalanır.
- **`allowMultiple`** — slot birden fazla fragment tutabilir. Bunlar bağlandıkları
  sırayla, art arda render edilir. Tek fragment tutan bir slot'a ikinci bir fragment
  bağlamak `ErrSlotOccupied` kaydeder.

`WithSlotFragment`, bildirilmemiş bir slot için `ErrUnknownSlot` — önce slot'ları
bildirin — ve nil bir alt fragment için `ErrNilFragment` kaydeder. Aynı adı iki kez
bildirmek `ErrDuplicateSlot` kaydeder.

Şablonda `{{slot "name"}}`, slot'un tuttuğunu HTML olarak render eder ve bu HTML
yeniden escape edilmez: alt fragment'ler kendi değerlerini render edilirken zaten
escape etmiştir. Hiçbir şey tutmayan bir slot hiçbir şey render etmez. Fragment'in hiç
bildirmediği bir ad ise boş çıktı değil, bir **hata**dır — aksi hâlde şablondaki bir
yazım hatası, sessizce eksik kalan bir bölüm olurdu.

### Her fragment'in kendi verisi vardır

Bir alt fragment, üst fragment'inin verisini görmez. `fragments/author.html` içinde
`.`, yazı değil, `loadAuthor`'ın döndürdüğüdür. Üst fragment'inin getirdiği bir şeye
ihtiyaç duyan bir alt fragment, onu render'ın paylaşılan verisinden okur ya da
yalnızca bir kez getirilmesi için `collage.Once` veya `collage.Cached` ile kendisi
ister — bkz. [Data handler'lar](/docs/data-handlers#sharing-data-between-fragments).

### Fragment'leri yeniden kullanmak

Bir `*collage.Fragment`, pek çok sayfadaki pek çok slot'a bağlanabilir; yukarıdaki
`author` hem yazı sayfasında hem de bir arama sonuçları sayfasında görünebilir. İki
kez bağlanan bir fragment iki kez render edilir ve data handler'ı her bağlama için
bir kez çalışır.

Bir fragment kendisini içermemelidir. Bir fragment'i doğrudan ya da bir alt fragment
zinciri üzerinden kendi slot'una bağlamak, kayıt sırasında `ErrFragmentCycle` ile
reddedilir.

## Bir fragment başarısız olduğunda

Bir fragment; data handler'ı bir hata döndürdüğünde, şablonu çalıştırılamadığında,
zorunlu bir slot'u boş olduğunda ya da bunlardan biri panic ettiğinde başarısız olur
— bir data handler'daki ya da bir şablon fonksiyonundaki panic, bir
`*collage.PanicError`'a dönüştürülerek kurtarılır ve çöken bir süreç olarak değil,
bir başarısızlık olarak ele alınır.

Bundan sonra ne olacağı fragment'in **başarısızlık politikası**dır ve üç tane vardır.

```go
postContent := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	Required().
	Build()

comments := collage.NewFragment("comments", "fragments/comments.html").
	WithDataHandler(collage.DataHandler(loadComments)).
	WithFallback(collage.NewFragment("comments-unavailable", "fragments/comments-unavailable.html").Build()).
	Build()

related := collage.NewFragment("related", "fragments/related.html").
	WithDataHandler(collage.DataHandler(loadRelated)).
	Build()
```

- **`Required()`** — fragment'in başarısızlığı sayfayı başarısız kılar. Bunu sayfanın
  göstermek için var olduğu şey için kullanın. Sayfanın hata sayfası 500 ile sunulur
  — ya da hata `collage.ErrNotFound`'u sarıyorsa, bulunamadı sayfası 404 ile.
  Başarısızlık, üstündeki isteğe bağlı fragment'lerin içinden geçip yukarı taşınır:
  isteğe bağlı bir kenar çubuğunun içindeki zorunlu bir fragment yine de sayfayı
  başarısız kılar.
- **`WithFallback(f)`** — onun yerine `f` render edilir. Yedek de başarısız olursa
  fragment hiçbir şey render etmez ve sayfa yine başarılı olur: yedek, bir
  başarısızlığı sınırlamak için vardır; bu yüzden içindeki bir şey zorunlu olarak
  işaretlenmiş olsa bile kendi başarısızlığı da sınırlanır.
- **Hiçbiri** — fragment hiçbir şey render etmez ve sayfa onsuz sunulur.

`collage.ErrNotFound` yalnızca zorunlu bir fragment'te "404" anlamına gelir. İsteğe
bağlı bir fragment'te diğerleri gibi bir başarısızlıktır, çünkü sayfa onsuz da render
edilebilir.

Yedekli ya da yedeksiz, başarısız olmuş herhangi bir fragment'le sunulan sayfa
**bozulmuş** (degraded) sayılır. Okuyucuya gönderilir ama asla önbelleğe alınmaz;
böylece bir sonraki istek, başarısızlığı bir TTL boyunca sunmak yerine yeniden dener.
Statik dışa aktarma, kendisine söylenmedikçe bozulmuş bir sayfayı yazmaz.

Geliştirme modunda bir başarısızlık asla sessiz kalmaz. Hiçbir şey render etmeyen
başarısız bir fragment, çıktısının olacağı yere adını ve hatasını içeren bir HTML
yorumu bırakır; sayfa da başarısız olan her fragment'i hatasıyla — bir şablon için
dosya ve satırıyla — adlandıran bir panel taşır, bir yedek onun yerini tutmuş olsa
bile. Zorunlu bir fragment sayfanın tamamını başarısız kıldığında, geliştirme hata
sayfası o fragment'i adlandırır — içinden geçip yukarı çıktığı layout'u değil,
başarısızlığın başladığı yeri. Geliştirme modu dışında ikisi de yoktur.

### Şablonun atladığı bir slot

Bir fragment'in alt fragment'leri, art arda değil aynı anda getirsinler diye, şablon
çalışmadan önce veri getirmeye başlar. Bu da şablonun render etmemeye karar verdiği
bir slot'taki fragment'in — `{{if .ShowComments}}{{slot "comments"}}{{end}}` — handler'ının
yine de başlatılmış olduğu anlamına gelir. Şablon biter bitmez context'i iptal edilir
ve başarısızlığı yok sayılır: hiç render edilmemiş bir fragment, `Required()` olsa
bile sayfayı başarısız kılamaz.

## Zaman aşımları

`WithTimeout(d)`, fragment'in data handler'ına bir sınır koyar. Zaman aşımı
ayarlamayan bir fragment, varsayılanı beş saniye olan `Config.Template.Timeout`'u
alır; negatif bir süre `ErrInvalidTimeout` kaydeder.

```go
recommendations := collage.NewFragment("recommendations", "fragments/recommendations.html").
	WithDataHandler(collage.DataHandler(loadRecommendations)).
	WithTimeout(300 * time.Millisecond).
	WithFallback(nothingToRecommend).
	Build()
```

Zaman aşımı, doğal olarak bir yedekle eşleşir. Yavaş bir öneri servisi, sayfaya
okuyucunun sabrına değil, 300 milisaniyeye ve bir yedeğe mal olur.

Zaman aşımı handler'ı değil, handler'ın aldığı **context**'i sınırlar: `ctx`'i yok
sayan bir handler istediği kadar çalışır. Bekleyebilecek her çağrıya `ctx`'i geçirin —
bkz. [Data handler'lar](/docs/data-handlers#timeouts-and-the-context).

## Her render'da doldurulan slot'lar

Yukarıdaki her şey alt fragment'leri program başlarken bağlar. Bazı sayfalar ise
içerikleri tarafından düzenlenir: bölümlerini bir editörün bir CMS'te seçip sıraladığı
bir açılış sayfası, bir kullanıcının seçtiği widget'lardan oluşan bir pano. Bunları
başlangıçta bağlamak, her değişiklikte yeniden başlatma ve neyin nereye gittiği
konusunda veriyle uyuşması gereken kod demek olurdu.

`WithSlotResolver` bir slot'u her render'da doldurur. Resolver, render context'ini
alır ve slot'un bu sefer tuttuğu fragment'leri döndürür:

```go
type block struct {
	Kind    string
	Heading string
	Text    string
}

landing := collage.NewFragment("landing", "pages/landing.html").
	WithDataHandler(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		blocks, err := cms.Blocks(ctx, "landing")
		if err != nil {
			return err
		}
		rc.Set("blocks", blocks)
		return nil
	})).
	WithSlot("blocks", true, true).
	WithSlotResolver("blocks", func(rc *collage.RenderContext) ([]*collage.Fragment, error) {
		blocks, _ := collage.Get[[]block](rc, "blocks")
		fragments := make([]*collage.Fragment, 0, len(blocks))
		for i, b := range blocks {
			f, err := blockFragment(i, b)
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, f)
		}
		return fragments, nil
	}).
	Required().
	Build()
```

Bir resolver, program boyunca var olan fragment'leri döndürebilir ya da onları o anda
kurabilir. Her bloğun kendi verisini alması, onları kurmakla olur:

```go
func blockFragment(i int, b block) (*collage.Fragment, error) {
	switch b.Kind {
	case "hero", "text":
		return collage.NewFragment(fmt.Sprintf("block-%d-%s", i, b.Kind), "blocks/"+b.Kind+".html").
			WithDataHandler(collage.DataHandler(func(context.Context, *collage.RenderContext) (block, []string, error) {
				return b, nil, nil
			})).
			Build(), nil
	}
	return nil, fmt.Errorf("landing: unknown block kind %q", b.Kind)
}
```

Kurallar:

- **Resolver, kendi fragment'inin data handler'ından sonra çalışır**; böylece o
  handler'ın getirdiğini okuyabilir. **Döndürdüğü fragment'ler kendi handler'larını
  başlatmadan önce** de çalışır — bu handler'lar ardından, diğer alt fragment'lerinki
  gibi eşzamanlı çalışır.
- **Döndürdüğü şey slot'un kurallarına tabidir**: slot birden fazlasına izin
  vermiyorsa en fazla bir fragment, zorunluysa en az bir fragment
  (`ErrRequiredSlotEmpty`). Resolver'dan gelen bir hata ya da bir panic, onun
  fragment'ini başarısız kılar ve o fragment'in başarısızlık politikası uygulanır.
- **Önce slot'u bildirin**, `WithSlot` ile. Bir slot ya bir resolver'la ya da
  `WithSlotFragment` ile doldurulur, asla ikisiyle birden değil — ikisini karıştırmak
  `ErrSlotResolved` kaydeder.
- **Fragment'leri yalnızca render edildiklerinde denetlenir.** Kayıt onları göremez;
  bu yüzden şablonu olmayan, döndürülmüş bir fragment başlangıcı değil, o render'ı
  başarısız kılar. Builder'ının kaydettiği bir hata da aynı şekilde yakalanır:
  hatalarla kurulmuş, döndürülmüş bir fragment, slot'un sahibi olan fragment'i o
  fragment'in başarısızlık politikası altında ve hatada döndürülen fragment'in adıyla
  başarısız kılar. Bunu resolver'da ele almayı tercih ettiğinizde `BuildErr()`'ü
  kendiniz denetleyin. Şablonların hepsi başlangıçta yüklenir; bu yüzden bir
  resolver'ın kullanabileceği blok türleri kümesi yine de program tarafından
  belirlenir.

Bölümleri içerikten gelen bir sayfa, o içeriğin etiketlerini de bildirmelidir —
burada açılış sayfasının handler'ı, `collage.Effect` yerine `collage.DataHandler` ile
`"landing"`'i bir etiket olarak döndürebilir — böylece bir editör sayfayı yeniden
sıraladığında önbellekteki sayfa düşürülür. Bkz. [Önbellek](/docs/caching).

## İç içe geçme sınırı

Bir fragment ağacı, kök de sayılmak üzere en fazla **32 seviye** derin olabilir. Daha
derine inen bir sayfa, fragment'lerinin başarısızlık politikaları ne olursa olsun
`ErrMaxDepthExceeded` ile render edilemez ve hata, sınıra ulaşan fragment zincirini
adlandırır.

Gerçek sayfalar bu sınırın yanına bile yaklaşmaz. Sınır, kaydın dışarıda
bırakamayacağı tek durum için vardır: aşağıda bir yerde yeniden kendisine çözümlenen
bir fragment döndüren bir resolver. Bağlanmış bir döngü zaten kayıt sırasında
`ErrFragmentCycle` olarak reddedilir.

## Kendi URL'si olan fragment'ler

Bir fragment, etrafındaki sayfa olmadan da getirilebilir — bir `fetch()` ile
yenilenen arama sonuçları, bir form gönderildikten sonra yerine takılan bir panel. Bu,
sayfa üzerinde `WithFragmentPath(locale, pattern, fragment)` ile bildirilir ve
bildirilmedikçe hiçbir şeye bu yolla erişilemez. Böyle bir fragment kayıt sırasında
sayfayla birlikte denetlenir — şablonu, builder'ının hataları, doğrulaması. Bkz.
[Formlar ve action'lar](/docs/forms-and-actions).
