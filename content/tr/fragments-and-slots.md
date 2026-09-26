---
description: Fragment'ler, sundukları slot'lar, bir fragment başarısız olduğunda ne olduğu ve render sırasında içerikten doldurulan slot'lar.
reference: NewFragment, FragmentBuilder, Fragment, FragmentBuilder.WithFallback, FragmentBuilder.WithSlot, FragmentBuilder.WithData, FragmentBuilder.Static, FragmentBuilder.Shared, SlotResolverFunc, ErrUnknownSlot
---

# Fragment'ler ve slot'lar

**Fragment**, bir page'in yapı taşıdır. Bir template'ten, template'in render ettiği
veriyi çeken isteğe bağlı bir data handler'dan ve başka fragment'lerin yerleştiği
isimli **slot**'lardan oluşur. Layout da bir fragment'tir. Bir page'in içeriği, bir
sidebar, bir yazar kartı ve bir yorum listesi de birer fragment'tir. Bir page, kökünde
layout'un durduğu bir fragment ağacıdır.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(loadAuthor).
	WithTimeout(time.Second).
	WithFallback(anonymousAuthor).
	Build()
```

## Fragment oluşturmak

`collage.NewFragment(name, templatePath)` bir builder başlatır. `Build()` ise
`*collage.Fragment`'i döner.

- **İsim**, fragment'in raporlarda nasıl görüneceğini belirler: hata mesajlarında,
  development error panel'inde, render metadata'sında ve trace'lerde. İsim,
  fragment'in page'in hangi kısmı olduğunu anlatmalıdır: `"list"` değil,
  `"post-comments"`.
- **Template path'i** template root'una göredir ve uzantıyı da içerir:
  `"fragments/author.html"`, `templates/fragments/author.html` dosyasıdır. Template
  engine'in yüklemediği bir path, fragment'i içeren page register edilirken
  `ErrTemplateNotFound` hatası verir. v0.11.0'dan beri bu kural, page'in
  `WithFragmentPath` ile kendi URL'sinde açtığı fragment'ler için de geçerlidir.
  Yalnızca bir [slot resolver](#slots-filled-per-render)'ın döndüğü fragment daha
  sonra, render edilirken kontrol edilir. Bkz. [Template'ler](/docs/templates).

Geri kalan her şey isteğe bağlıdır:

| Metot | Neyi ayarlar |
| --- | --- |
| `WithDataHandler(h)` | Template'in verisini çeken fonksiyonu. Bkz. [Data handler'lar](/docs/data-handlers) |
| `WithData(v)` | Handler yerine, program başlarken sabitlenen veriyi |
| `WithTitle(s)` | Page'in `<title>`'ını, handler olmadan. Bkz. [Head ve SEO](/docs/head-and-seo) |
| `WithSlot(name, required, allowMultiple)` | Bir slot'u required yapar ya da tek fragment'le sınırlar |
| `WithSlotFragment(slot, child)` | Bir child fragment'i bir slot'a bağlar |
| `WithSlotResolver(slot, resolve)` | Slot'u bunun yerine her render'da doldurur |
| `Required()` | Bu fragment başarısız olursa page de başarısız olur |
| `WithFallback(f)` | Bu fragment başarısız olduğunda neyin render edileceğini |
| `WithTimeout(d)` | Data handler'ının ne kadar sürebileceğini |
| `Static()` | Data handler'ının aynı URL'e gelen her request'e aynı sonucu döndürdüğünü, böylece page'i dynamic yapmadığını. v0.17.0'dan beri |
| `Shared()` | Data handler'ının belli bir anda her okuyucuya aynı sonucu döndürdüğünü, ama bu sonucun zamanla değiştiğini. Böylece render'ı birçok okuyucuya gönderilebilir ve page'in stratejisine dokunulmaz. v0.19.0'dan beri |

Page builder'da olduğu gibi fragment builder da hataları zinciri durdurmadan kaydeder
ve `BuildErr()` bunları döner. Kaydedilen hatalar, builder'ın oluşturduğu fragment'in
üzerinde kalır. İki kez kısıtlanmış bir slot, teke sınırlanmış bir slot'a bağlanan
ikinci bir child ya da negatif bir timeout gibi hataların her biri, ağacında bu
fragment bulunan her page'i register sırasında durdurur. Birinin `BuildErr()`'ü
çağırıp çağırmadığı fark etmez. Hatayı `RegisterPage`'de değil de ona yol açan
satırda almak istediğinizde `BuildErr()`'ü çağırın. Bkz. [Page'ler ve
layout'lar](/docs/pages-and-layouts#building-a-page).

Data handler'ı olmayan bir fragment, template'ini veri olmadan ya da
`WithData(v)`'nin her render'da verdiği değerle render eder. Bu, hiç değişmeyen
markup için doğru seçimdir: bir footer, static bir duyuru ya da bir link listesi
gibi. Tek işi slot'ları yerleştirmek olan bir layout için de doğrudur. Page'i
cache'lenebilir de tutar: strateji tanımlamayan bir page, render ettiği bir şeyin
data handler'ı ya da slot resolver'ı olmadıkça static'tir. Bkz.
[Caching](/docs/caching#a-page-that-declares-none). Handler'ı yalnızca path'in
parametrelerini ve locale'i okuyan bir fragment `Static()` der ve handler'ı artık
hesaba katılmaz. Handler'ı her okuyucu için aynı olan ama zaman içinde aynı
kalmayan bir fragment (bir ölçüm gibi) `Shared()` der. Bu, page hakkında hiçbir şey
söylemez ve handler yine hesaba katılır. İki sözden birini bozan bir handler, bir
okuyucunun verisini başka bir okuyucuya gönderir. `WithData` ile
`WithDataHandler`'ı birlikte ayarlamak, register sırasında `ErrConflictingData`
hatası verir.

## Slot'lar

Slot, bir fragment'in template'inde `{{slot "name"}}` ile yazılan isimli bir
konumdur. Slot'u tanımlamak için template'te onu çağırmak yeterlidir. Fragment,
child'ları ona ismiyle bağlar:

```go
post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(loadPost).
	WithSlot("author", true, false).
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

Bir slot istediği sayıda fragment tutar. Bu fragment'ler bağlandıkları sırayla, art
arda render edilir. Yukarıdaki `related` iki fragment tutar. Hiçbir şey bağlanmamış
bir slot hiçbir şey render etmez. Bir layout'ta da tanımlama gerekmez: register
işlemi page'in içeriğini layout'un `"content"` slot'una yerleştirir.

`WithSlot(name, required, allowMultiple)`, bu davranışı istemediğiniz durumlar
içindir. Yukarıdaki `author` doldurulmak zorundadır ve tek fragment tutar.
`WithSlot`, kısıtladığı bağlamalardan önce de sonra da gelebilir.

- **`required`**: Slot'ta bir şey olmak zorundadır. Hiçbir şey bağlanmamış required
  bir slot, register sırasında reddedilir (`ErrRequiredSlotUnfilled`). Template
  çalışmadan önce de tekrar kontrol edilir. Böylece template bu slot'u hiç
  kullanmasa bile hata yakalanır.
- **`allowMultiple`** false verildiğinde: Slot en fazla bir fragment tutar. İkinci
  bir fragment bağlamak `ErrSlotOccupied` kaydeder.

`WithSlotFragment`, nil bir child için `ErrNilFragment` kaydeder. `WithSlot`'u aynı
isimle iki kez çağırmak `ErrDuplicateSlot` kaydeder.

Template'te `{{slot "name"}}`, slot'un tuttuğu içeriği HTML olarak render eder ve bu
HTML tekrar escape edilmez. Çünkü child'lar kendi değerlerini render edilirken zaten
escape etmiştir.

Bir **bağlamanın iki tarafından birindeki yazım hatası**, örneğin
`{{slot "sidebar"}}` karşısında `WithSlotFragment("sidbar", ...)`, register işlemini
`ErrUnknownSlot` ile başarısız kılar. Hata, slot'u ve template'in gerçekten çağırdığı
slot'ları adlarıyla belirtir. Template'inin hiç çağırmadığı bir slot'a bağlanan bir
fragment zaten hiçbir zaman render edilemez. Template'in include ettiği bir
template'teki ya da tanımladığı bir block'taki çağrılar da sayılır. Bir slot'u
literal dışında bir şeyle, `{{slot .Which}}` ile adlandıran bir template herhangi
birini çağırabilir. Bu yüzden o fragment'in bağlamaları kontrol edilmez.

### Her fragment'in kendi verisi vardır

Bir child, parent'ının verisini görmez. `fragments/author.html` içinde `.`, post
değil, `loadAuthor`'ın döndüğü değerdir. Parent'ın çektiği bir şeye ihtiyacı olan
bir child, bunu render'ın shared data'sından okur. Ya da veriyi `collage.Once`
veya `collage.Cached` ile kendisi ister, böylece veri yalnızca bir kez çekilir.
Bkz. [Data handler'lar](/docs/data-handlers#sharing-data-between-fragments).

### Fragment'leri yeniden kullanmak

Bir `*collage.Fragment`, birçok page'deki birçok slot'a bağlanabilir. Yukarıdaki
`author` hem post page'inde hem de bir arama sonuçları page'inde görünebilir. İki kez
bağlanan bir fragment iki kez render edilir ve data handler'ı her bağlama için bir kez
çalışır.

Bir fragment kendisini içeremez. Bir fragment'i doğrudan ya da bir child zinciri
üzerinden kendi slot'una bağlarsanız, bu register sırasında `ErrFragmentCycle` ile
reddedilir.

## Bir fragment başarısız olduğunda

Bir fragment şu durumlarda başarısız olur: data handler'ı bir hata döndüğünde,
template'i çalıştırılamadığında, required bir slot'u boş kaldığında ya da data
handler veya template panic ettiğinde. Bir data handler'daki ya da bir template
fonksiyonundaki panic recover edilir ve bir `*collage.PanicError`'a çevrilir. Bu
durum çöken bir process olarak değil, bir başarısızlık olarak ele alınır.

Bundan sonra ne olacağını fragment'in **failure policy**'si belirler. Üç policy
vardır.

```go
postContent := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(loadPost).
	Required().
	Build()

comments := collage.NewFragment("comments", "fragments/comments.html").
	WithDataHandler(loadComments).
	WithFallback(collage.NewFragment("comments-unavailable", "fragments/comments-unavailable.html").Build()).
	Build()

related := collage.NewFragment("related", "fragments/related.html").
	WithDataHandler(loadRelated).
	Build()
```

- **`Required()`**: Fragment başarısız olursa page de başarısız olur. Bunu, page'in
  asıl göstermek için var olduğu içerik için kullanın. Page'in error page'i 500 ile
  sunulur. Hata `collage.ErrNotFound`'u wrap ediyorsa not-found page'i 404 ile
  sunulur. Başarısızlık, üstündeki optional fragment'lerden geçerek yukarı taşınır.
  Optional bir sidebar'ın içindeki required bir fragment de page'i başarısız kılar.
- **`WithFallback(f)`**: Fragment'in yerine `f` render edilir. Fallback de başarısız
  olursa fragment hiçbir şey render etmez ve page yine başarılı olur. Fallback bir
  başarısızlığı sınırlamak için vardır. Bu yüzden içindeki bir şey required olarak
  işaretlenmiş olsa bile fallback'in kendi başarısızlığı da sınırlanır.
- **Hiçbiri**: Fragment hiçbir şey render etmez ve page onsuz sunulur.

`collage.ErrNotFound` yalnızca required bir fragment'te "404" anlamına gelir. Optional
bir fragment'te diğerleri gibi sıradan bir başarısızlıktır, çünkü page onsuz da
render edilebilir.

Fallback'li ya da fallback'siz, başarısız olmuş herhangi bir fragment'le sunulan
page **degraded** sayılır. Degraded page okuyucuya gönderilir ama asla cache'lenmez.
Böylece sonraki request, hatayı bir TTL boyunca sunmak yerine yeniden dener.
Static export, açıkça söylenmedikçe degraded bir page'i yazmaz.

Development'ta bir hata asla sessiz kalmaz. Hiçbir şey render etmeyen, başarısız
olmuş bir fragment, çıktısının olacağı yere ismini ve hatasını içeren bir HTML yorumu
bırakır. Page ayrıca başarısız olan her fragment'i hatasıyla birlikte listeleyen bir
panel taşır. Template hatalarında panel dosyayı ve satırı da gösterir. Bir fallback
fragment'in yerini tutmuş olsa bile panel yine görünür. Required bir fragment bütün
page'i başarısız kıldığında, development error page'i o fragment'in ismini verir.
Yani hatanın içinden geçerek yukarı çıktığı layout'u değil, başladığı yeri gösterir.
Development dışında bunların hiçbiri yoktur.

### Template'in atladığı bir slot

Bir fragment'in child'ları, template çalışmadan önce veri çekmeye başlar. Böylece
verileri art arda değil, aynı anda çekerler. Bu, template'in render etmemeye karar
verdiği bir slot'taki fragment'in de handler'ının başlatılmış olduğu anlamına gelir:
`{{if .ShowComments}}{{slot "comments"}}{{end}}`. Template biter bitmez bu
fragment'in context'i cancel edilir ve hatası yok sayılır. Hiç render edilmemiş bir
fragment, `Required()` olsa bile page'i başarısız kılamaz.

## Timeout'lar

`WithTimeout(d)`, fragment'in data handler'ına bir süre sınırı koyar. Timeout
ayarlamayan bir fragment, `Config.Template.Timeout` değerini alır. Bu değerin
varsayılanı beş saniyedir. Negatif bir süre `ErrInvalidTimeout` kaydeder.

```go
recommendations := collage.NewFragment("recommendations", "fragments/recommendations.html").
	WithDataHandler(loadRecommendations).
	WithTimeout(300 * time.Millisecond).
	WithFallback(nothingToRecommend).
	Build()
```

Timeout, fallback ile doğal olarak birlikte kullanılır. Yavaş bir öneri servisi,
page'e okuyucunun sabrına değil, 300 milisaniyeye ve bir fallback'e mal olur.

Timeout handler'ı değil, handler'ın aldığı **context**'i sınırlar. `ctx`'i yok sayan
bir handler istediği kadar çalışır. Bekleyebilecek her çağrıya `ctx`'i geçin. Bkz.
[Data handler'lar](/docs/data-handlers#timeouts-and-the-context).

## Her render'da doldurulan slot'lar

Buraya kadar anlatılan her şey, child'ları program başlarken bağlar. Bazı page'lerin
düzenini ise içerikleri belirler. Örneğin bölümlerini bir editörün CMS'te seçip
sıraladığı bir landing page ya da kullanıcının seçtiği widget'lardan oluşan bir
dashboard. Bunları başlangıçta bağlamak, her değişiklikte yeniden başlatma gerektirir.
Ayrıca neyin nereye gideceği konusunda veriyle uyumlu kalması gereken bir kod
yazmanız gerekir.

`WithSlotResolver`, bir slot'u her render'da doldurur. Resolver, render context'ini
alır ve slot'un bu render'da tutacağı fragment'leri döner:

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

Bir resolver, program boyunca var olan fragment'leri dönebilir ya da onları o anda
oluşturabilir. Her block'un kendi verisini alması, fragment'leri o anda oluşturarak
sağlanır:

```go
func blockFragment(i int, b block) (*collage.Fragment, error) {
	switch b.Kind {
	case "hero", "text":
		return collage.NewFragment(fmt.Sprintf("block-%d-%s", i, b.Kind), "blocks/"+b.Kind+".html").
			WithData(b).
			Build(), nil
	}
	return nil, fmt.Errorf("landing: unknown block kind %q", b.Kind)
}
```

Kurallar şunlardır:

- **Resolver, kendi fragment'inin data handler'ından sonra çalışır.** Böylece o
  handler'ın çektiği veriyi okuyabilir. **Döndüğü fragment'ler kendi handler'larını
  başlatmadan önce çalışır.** Bu handler'lar da ardından, diğer child'larda olduğu
  gibi eşzamanlı çalışır.
- **Resolver'ın döndüğü değer slot'un kurallarına tabidir.** Slot birden fazla
  fragment'e izin vermiyorsa en fazla bir fragment dönebilir. Slot required ise en az
  bir fragment dönmelidir (`ErrRequiredSlotEmpty`). Resolver'ın döndüğü bir hata ya
  da resolver'daki bir panic, resolver'ın ait olduğu fragment'i başarısız kılar ve
  o fragment'in failure policy'si uygulanır.
- **Yalnızca bir resolver'a bağlanan slot isteğe bağlıdır ve istediği sayıda fragment
  tutar.** Ondan önce ya da sonra çağrılan `WithSlot`, onu required ya da tek
  fragment'lik yapar. Yukarıdaki `blocks` bu şekilde required'dır. Bir slot ya bir
  resolver ile ya da `WithSlotFragment` ile doldurulur, ikisiyle birden asla
  doldurulmaz. İkisini karıştırmak `ErrSlotResolved` kaydeder.
- **Bir resolver, tıpkı bir data handler gibi, strateji tanımlamayan page'i dynamic
  yapar.** Resolver'ın döndüğü şey request'e bağlı olabilir.
- **Resolver'ın fragment'leri yalnızca render edilirken kontrol edilir.** Register
  işlemi bu fragment'leri göremez. Bu yüzden template'i olmayan bir fragment
  döndürülürse uygulama başlarken değil, o render sırasında hata oluşur. Builder'ın
  kaydettiği bir hata da aynı şekilde yakalanır. Hatalarla oluşturulmuş bir fragment
  döndürülürse, slot'un sahibi olan fragment başarısız olur. Bu durumda o fragment'in
  failure policy'si uygulanır ve hata, döndürülen fragment'in ismini içerir. Hatayı
  resolver'ın içinde ele almak isterseniz `BuildErr()`'ü kendiniz kontrol edin.
  Template'lerin hepsi başlangıçta yüklenir. Bu yüzden bir resolver'ın
  kullanabileceği block türleri yine de program tarafından belirlenir.

Bölümleri içerikten gelen bir page, o içeriğin tag'lerini de bildirmelidir. Buradaki
örnekte landing handler'ı, `collage.Effect` yerine `WithDataHandler`'ın kendi
biçiminde yazılıp `"landing"`'i tag olarak dönebilir. Böylece bir editör page'i
yeniden sıraladığında cache'teki page atılır. Bkz. [Caching](/docs/caching).

## İç içe geçme sınırı

Bir fragment ağacı, root dahil en fazla **32 seviye** derinliğinde olabilir. Daha
derine inen bir page, fragment'lerinin failure policy'leri ne olursa olsun
`ErrMaxDepthExceeded` ile render edilemez. Hata, sınıra ulaşan fragment zincirini
isim isim gösterir.

Gerçek page'ler bu sınırın yakınına bile gelmez. Sınır, register işleminin önceden
engelleyemediği tek bir durum için vardır: döndüğü bir fragment'in, aşağıda bir
yerde yine kendisine resolve olduğu bir resolver. Bağlanmış bir döngü zaten register
sırasında `ErrFragmentCycle` olarak reddedilir.

## Kendi URL'si olan fragment'ler

Bir fragment, etrafındaki page olmadan da getirilebilir. Örneğin bir `fetch()` ile
yenilenen arama sonuçları ya da bir form gönderildikten sonra yerine konan bir panel.
Bu, page üzerinde `WithFragmentPath(locale, pattern, fragment)` ile tanımlanır.
Tanımlanmayan hiçbir şeye bu yolla erişilemez. Böyle bir fragment, register sırasında
page ile birlikte kontrol edilir: template'i, builder'ının kaydettiği hatalar ve
validation'ı. Bkz. [Form'lar ve action'lar](/docs/forms-and-actions).
