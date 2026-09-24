---
description: Plugin sözleşmesi, Host ve ConfigHost'un sundukları, her hook ve neyi değiştirebileceği, testleriyle birlikte eksiksiz bir plugin.
---

# Plugin yazmak

Bir plugin, dört metodu olan bir Go tipidir. Geri kalan her şey — bir render'a
tepki vermek, çıktıyı yeniden yazmak, bir şablon fonksiyonu eklemek — isteğe
bağlıdır: istediğiniz hook'un interface'ini uygularsınız, framework de onu tip
doğrulamasıyla (type assertion) bulur.

Bu sayfa o yüzeyin başvuru kaynağıdır. [Plugin kullanmak](/docs/plugins) işin öbür
tarafıdır: bir uygulamanın sizin yazdığınızı nasıl kaydettiği ve yapılandırdığı.

## Sözleşme

```go
type Plugin interface {
	Name() string
	Version() string
	Init(ctx context.Context, host collage.Host) error
	Shutdown(ctx context.Context) error
}
```

- **`Name`** plugin'i tanımlar. Boş olmamalı ve uygulama içinde benzersiz olmalıdır;
  plugin'in yapılandırma bölümünün anahtarı da odur — bu yüzden bir modül yolu gibi
  okunmasını sağlayın: `acme/stamp`.
- **`Version`** plugin'inizin kendi sürümüdür, tanılama için.
- **`Init`** uygulama başladığında bir kez çalışır: uygulama kendi sayfalarını
  kaydettikten sonra ve ilk istekten önce.
- **`Shutdown`**, `Init`'in edindiği her şeyi serbest bırakır. `Init`'i çalışmış
  ya da başarılı olmuş olsun olmasın, kayıtlı her plugin için çağrılır ve birden
  fazla kez çağrılabilir; bu yüzden `Init` olmadan da güvenli ve idempotent olmalıdır
  — bkz. [Yaşam döngüsü](#lifecycle).

Hook'lar tip doğrulamasıyla bulunduğu için, adı yanlış yazılmış bir hook metodu
derleme hatası değildir — hiç çalışmayan bir hook'tur. Uygulamayı amaçladığınız her
interface'i doğrulayın:

```go
var (
	_ collage.Plugin          = (*Plugin)(nil)
	_ collage.AfterRenderHook = (*Plugin)(nil)
)
```

## İki aşama: Configure ve Init

Bazı işlerin şablonlar ayrıştırılmadan önce yapılması gerekir. `html/template`
yalnızca şablon ayrıştırılırken fonksiyon haritasında bulunan bir fonksiyonu
çağırabilir ve ayrıştırma `collage.New` içinde olur. Bu yüzden isteğe bağlı, daha
erken bir aşama vardır:

```go
type Configurer interface {
	Configure(ctx context.Context, host collage.ConfigHost) error
}
```

`Configure`, `New` içinde, plugin başına bir kez, kayıt sırasıyla ve şablonlar
ayrıştırılmadan önce çalışır. Döndürülen bir hata `New`'u iptal eder. Henüz hiçbir
şey edinilmemiştir, bu yüzden geri alma (rollback) yoktur.

`Init` daha sonra, uygulama başladığında çalışır — `Handler`, `ListenAndServe`,
`Start`, `RenderPath`, `RenderDocumentPath` ya da `DispatchCommands`'a yapılan ilk
çağrıda; buna statik build de dahildir, çünkü builder `RenderPath` üzerinden render
eder. O zamana kadar uygulama sayfalarını kaydetmiştir, dolayısıyla bir plugin
onları okuyabilir ya da kendi sayfalarını ekleyebilir.

İkisine de ihtiyaç duyan bir plugin ikisini de uygular. **`Configurer`'ı uygulayan
bir plugin `Config.Plugins` içinde verilmelidir**: `RegisterPlugin`, `New`
şablonları ayrıştırdıktan sonra çağrılır; bu yüzden böyle bir plugin'in
`Configure`'ını sessizce atlamak yerine onu `ErrConfigurerRegisteredLate` ile
reddeder.

### Her aşama nelere erişebilir

| | `ConfigHost` (Configure) | `Host` (Init) |
| --- | --- | --- |
| `DevMode`, `Logger`, `Config` | evet | evet |
| `AddTemplateFunc`, `WrapMount` | evet | — |
| `Pages`, `Page`, `InvalidateTags` | — | evet |
| `RegisterPage`, `RegisterDocument`, `Mount` | — | evet |
| `RegisterCommand` | — | evet |

`ConfigHost` bilerek daha dardır. `Configure` sırasında uygulama henüz hiçbir şey
kaydetmemiştir; sayfalar boş bir liste olurdu, geçersiz kılmanın da ulaşacağı bir
önbellek olmazdı.

### ConfigHost

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın geliştirme modunda çalışıp çalışmadığı. |
| `Logger() *slog.Logger` | Uygulamanın logger'ı. |
| `Config(v) error` | Bu plugin'in yapılandırma bölümünü `v`'ye çözer — bkz. [Yapılandırma](#configuration). |
| `AddTemplateFunc(name, fn) error` | Bir şablon fonksiyonu ekler. Ad daha önce eklenmişse — başka bir plugin tarafından ya da bu plugin tarafından daha önce — `ErrDuplicateTemplateFunc` döner. |
| `WrapMount(wrap func(fs.FS) fs.FS)` | Mount edilen her dosya sistemine, sarmalayıcıların kaydedildiği sırayla uygulanan bir dönüşüm kaydeder. |

### Host

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın geliştirme modunda çalışıp çalışmadığı. |
| `Logger() *slog.Logger` | Uygulamanın logger'ı. |
| `Config(v) error` | Bu plugin'in yapılandırma bölümünü `v`'ye çözer. |
| `Pages() []*collage.Page` | Kayıtlı her sayfa, her biri savunmacı bir kopya. |
| `Page(name) (*collage.Page, bool)` | Adıyla tek bir sayfa, savunmacı bir kopya. |
| `InvalidateTags(ctx, tags...) error` | Etiketlerden herhangi biriyle kurulmuş her önbellek girdisini düşürür. |
| `RegisterPage(page) error` | Plugin'in katkıda bulunduğu bir sayfayı kaydeder. |
| `RegisterDocument(doc) error` | Plugin'in katkıda bulunduğu bir document'ı kaydeder. |
| `Mount(prefix, fsys, opts...) error` | Bir dosya sistemini bir URL öneki altında sunar. |
| `RegisterCommand(cmd) error` | Bir komut ekler — bkz. [Komutlar](#commands). |

`Init`'in aldığı şey `*App` değildir. Bu metotları ileten ve başka hiçbir şey
yapmayan dar bir değerdir; bu yüzden bir plugin tip doğrulamasıyla
`ListenAndServe`'e, `Shutdown`'a, router'a, önbelleğe ya da şablon kümesine
ulaşamaz.

**`Host` bir plugin'in neye ulaşabileceğini sınırlar, neyi değiştirebileceğini
değil.** `Pages` ve `Page`, sayfa struct'ının ve onun `Paths`, `Redirects`, `SEO`
ve `DependencyTags` kaplarının kopyalarını döndürür; bu yüzden onları düzenlemek
uygulamanın kendi sayfasına dokunmaz. Bir kopyanın içindeki fragment işaretçileri
ise hâlâ paylaşılır ve aşağıdaki olaylar kopyayı değil, *canlı* sayfayı taşır — her
istekte bir sayfayı ve fragment ağacını kopyalamak sıcak yola (hot path) pahalıya
patlardı. Bir olayın `Page`'i üzerinden yazmak, eşzamanlı her isteğin okuduğu
sayfayı değiştirir: bu bir veri yarışıdır (data race) ve `go test -race` bunu
söyler. Sayfaları salt okunur kabul edin. Plugin'ler güvenilen koddur, bir sandbox
değil.

Bir plugin'in kaydettiği sayfalar, document'lar ve mount'lar, uygulamanın
kendilerininkiyle aynı kurallara tabidir: zaten alınmış bir ad ya da yol bir
başlangıç hatasıdır, kayıt sırasına göre sonuçlanan bir yarış değil.

## Hook'lar

| Interface | Metot | Olay | Ne zaman tetiklenir | Neyi değiştirebilir |
| --- | --- | --- | --- | --- |
| `PageResolvedHook` | `OnPageResolved` | `PageResolvedEvent` | Her sayfa isteğinde bir kez, routing'in hemen ardından — önbellek isabetleri dahil | hiçbir şeyi |
| `BeforeRenderHook` | `OnBeforeRender` | `BeforeRenderEvent` | Taze bir sayfa render'ından önce | olayda hiçbir şeyi; `ev.Context` üzerinden hoist edebilir |
| `AfterRenderHook` | `OnAfterRender` | `AfterRenderEvent` | Bir sayfa render'ı başarılı olduktan sonra | `ev.HTML` |
| `DocumentRenderedHook` | `OnDocumentRendered` | `DocumentRenderedEvent` | Bir document handler'ı gövdesini ürettikten sonra | `ev.Body` |
| `CacheWriteHook` | `OnCacheWrite` | `CacheWriteEvent` | Bir sayfa ya da document önbelleğe yazılmadan önce | `ev.Skip`, `ev.TTL`, `ev.Tags` |
| `CacheInvalidateHook` | `OnCacheInvalidate` | `CacheInvalidateEvent` | Girdiler etiketle geçersiz kılındıktan sonra | hiçbir şeyi |
| `ErrorHook` | `OnError` | `ErrorEvent` | Bir istek sunulurken oluşan bir hatada | hiçbir şeyi |

Her hook metodunun biçimi `func(ctx context.Context, ev *Event) error`'dır.

### PageResolvedHook

```go
type PageResolvedEvent struct {
	Page   *collage.Page // live — do not write through it
	Locale string
	Path   string
}
```

Bir sayfaya yönlenen her istekte, önbelleğe bakılmadan önce bir kez tetiklenir;
bu yüzden taze render'ları olduğu kadar önbellek isabetlerini de görür. Bir
document için asla tetiklenmez, statik build sırasında da tetiklenmez — build bir
istek değildir ve istekleri sayan bir plugin kimsenin istemediği render'ları
sayardı. Bir hata, isteği `"page_resolved"` aşaması altında 500 ile başarısız
kılar.

### BeforeRenderHook

```go
type BeforeRenderEvent struct {
	Context *collage.RenderContext // the render about to run
	Page    *collage.Page
	Locale  string
	Path    string
}
```

Taze bir render'dan hemen önce tetiklenir, önbellek isabetinde ise **tetiklenmez**
— `PageResolvedHook`'tan farkı budur. Sayfalar için, hata sayfaları için, bir
action'ın `RenderPage` ile yanıt verdiği sayfa için ve statik build'in render
ettiği her sayfa için tetiklenir.

Render context'ini alan tek hook budur ve nedeni hoist etmektir. Sayfaya katkıda
bulunan bir plugin, bildirimini ağaç render edilmeden önce yapmak zorundadır:

```go
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}
```

Burada yapılan bir bildirim sıfır derinliğinde durur; bu yüzden aynı anahtarı
bildiren herhangi bir fragment onun yerini alır: varsayılanı plugin, özel olanı
sayfa sağlar. Yalnızca layout'un `{{hoist "head"}}` çağırdığı yere düşer. Bir hata,
isteği `"before_render"` altında 500 ile başarısız kılar.

### AfterRenderHook

```go
type AfterRenderEvent struct {
	Page     *collage.Page
	Locale   string
	Degraded bool   // some fragment failed, fallback or not
	HTML     []byte // replace it to post-process the page
	// Data: the render's shared data, the map behind rc.Set and rc.Get
}
```

Bir sayfa render'ı başarılı olduktan sonra tetiklenir: sayfalar, hata sayfaları,
bir action'ın `RenderPage` ile yanıt verdiği sayfa ve statik build'ler için.
Sonradan işlemek için `ev.HTML`'i değiştirin; orada bıraktığınız şey sunulan
şeydir ve — bir cache-write hook'u onu atlamadıkça — önbelleğe alınan şeydir. Sonraki
plugin'ler öncekilerin ürettiğini görür.

`ev.Data` render'ın paylaşılan verisidir — fragment'lerin `rc.Set` ve `rc.Get` ile
okuyup yazdığı haritanın ta kendisi —; yani sayfanın *neyden* kurulduğudur ve
geri ayrıştıracağı markup yerine makaleyi isteyen bir plugin içindir. İçinde ne
olduğu tamamen uygulamanın kuralıdır; framework oraya hiçbir şey koymaz. Canlı
haritadır: onu okumak sorun değildir, hook'tan sonra da tutmak istek durumunu elde
tutmak demektir.

Bulunduğu yerin iki sonucu:

- **Önbellek isabetinde yeniden çalışmaz.** Çıktısı önbelleğe alınan şeydir. Her
  istekte çalışması gereken bir hook, önbelleğe alınan bir sayfayla birleştirilemez.
- **Boş bir sonuç bir hatadır.** Dağıtımdan (dispatch) sonra `ev.HTML` boşsa, istek
  boş bir sayfa sunmak yerine 500 ile başarısız olur.

Bir action'ın `RenderPage` ile yanıt verdiği sayfa da onu çalıştırır (v0.10.0'dan
itibaren; öncesinde yalnızca `BeforeRender`'ı çalıştırıyordu), dolayısıyla bir
doğrulama sayfası da diğerleri gibi küçültülür. Bir hata, isteği `"after_render"`
altında 500 ile başarısız kılar.

### DocumentRenderedHook

```go
type DocumentRenderedEvent struct {
	Document    *collage.Document
	ContentType string
	Locale      string
	Path        string
	Body        []byte // replace it to transform the document
}
```

`AfterRenderHook`'un [document'lar](/docs/documents) için karşılığı — sitemap'ler,
feed'ler, JSON. O olmasaydı, çıktıyı sonradan işleyen bir plugin sayfaları kapsar
ve geri kalan her şeyi sessizce atlardı. ETag hesaplanmadan ve gövde önbelleğe
alınmadan önce tetiklenir; dolayısıyla ürettiğiniz şey saklanan ve ETag'in
tanımladığı şeydir. Bir gövdeye dokunup dokunmayacağınıza karar vermek için
`ContentType`'a bakın. Dağıtımdan sonra boş bir gövde ya da bir hata, isteği 500
ile başarısız kılar.

Bir document `OnPageResolved`, `OnBeforeRender` ya da `OnAfterRender` dağıtmaz:
sayfası yoktur ve şablon render etmez.

### CacheWriteHook

```go
type CacheWriteEvent struct {
	Key  string        // the cache key; changing it changes nothing
	Page *collage.Page // nil for a document
	TTL  time.Duration // may be adjusted
	Tags []string      // may be adjusted
	Skip bool          // set true to suppress the write
}
```

Render edilmiş bir sayfa ya da document saklanmadan önce tetiklenir. **Bir document
için `Page` `nil`'dir** ve onu korumasız dereference eden bir hook her document
isteğinde panic'e düşer — panic kontrol altına alınır ama dağıtıldığı yazma işlemi
bırakılır, dolayısıyla document hiç önbelleğe alınmaz:

```go
func (p *Plugin) OnCacheWrite(_ context.Context, ev *collage.CacheWriteEvent) error {
	if ev.Page == nil {
		return nil // a document; Key, TTL and Tags are still valid
	}
	if ev.Page.Name == "home" {
		ev.TTL = time.Minute
	}
	return nil
}
```

Bir hata ya da `Skip` yazmayı engeller ve istek yine de başarılı olur: sayfa zaten
render edilmiştir ve onu önbelleğe almadan sunmak, bir önbellek sorununu 500'e
çevirmekten iyidir. Hata, error hook'larına `"cache_write"` altında bildirilir.

### CacheInvalidateHook

```go
type CacheInvalidateEvent struct {
	Tags []string
}
```

`InvalidateTags` bazı etiketlerin girdilerini düşürdükten sonra tetiklenir — onu
uygulama, bir action ya da bir plugin çağırmış olsun. Bir istekten değil, o
çağrıdan dağıtılır ve bir hata `InvalidateTags`'in döndürdüğüne eklenir (join).
Geçersiz kılmayı kendiniz tetiklemek için `Host.InvalidateTags`'i çağırın.

### ErrorHook

```go
type ErrorEvent struct {
	Err   error
	Page  *collage.Page // nil unless the failure was a page's own; see below
	Path  string
	Stage string
}
```

Bir istek sunulurken oluşan bir hatada tetiklenir: bir sayfa, bir document, bir
action, bir mount ya da `App.Handle` ile kaydedilmiş bir handler. `Stage` hatanın
nerede olduğunu adlandırır.

`Page` yalnızca routing'in çözümlediği bir sayfanın kendi hatasında dolu olur.
Hiçbir sayfa çözümlenmediğinde `nil`'dir; bir document, bir action — action'ın
`RenderPage` ile yanıt verdiği sayfa dahil —, bir mount ve bir `App.Handle`
handler'ı için de öyle. Bunları birbirinden ayırmak için `Path`'i okuyun ve
`Page`'in her kullanımını koruma altına alın. Framework'ün kullandığı aşamalar
`"route"`, `"not_found"`, `"page_resolved"`, `"before_render"`, `"render"`,
`"after_render"`, `"cache_write"`, `"error_page"`, `"asset"`, `"handler"` ve
`"panic"`'tir — küme kapalı bir enum değildir.

Uyarı kurmaya değer olan `"error_page"`'dir: hataları bildiren sayfanın kendisinin
başarısız olduğu ve istemcinin yine de makul görünen yerleşik bir sayfa aldığı
anlamına gelir; yani başka türlü kimse bunu fark etmezdi.

`Err`'i `errors.Is` ile sınıflandırın — hiçbir şeyle eşleşmeyen bir URL için
`collage.ErrNoRoute`, var olmayan içerik için `collage.ErrNotFound`, bir 405 için
`collage.ErrMethodNotAllowed`, reddedilen bir gönderim için `collage.ErrCSRFMissing`
ve kardeşleri, 4xx ya da 5xx yanıt veren bir mount için `collage.ErrAssetFailed`,
kurtarılmış bir panic için `collage.ErrPanic` ve geri kalanlar
[Hatalar](/docs/errors#reported-to-error-hooks) sayfasında listelenmiştir.

`OnError`'dan döndürülen bir hata loglanır ve yutulur; kalan plugin'ler olayı yine
de alır: başarısız olan bir hata işleyicisi yeni bir hata işleme turu
başlatmamalıdır.

### Dağıtım kuralları

- Hook'lar **kayıt sırasıyla** çalışır.
- Her çağrı **panic'e karşı korunur**. Panic'e düşen bir hook, hata döndürmüş bir
  hook gibi başarısız olur; süreci çökertmez.
- `OnPageResolved`, `OnBeforeRender`, `OnAfterRender` ve `OnDocumentRendered` için
  **ilk hata dağıtımı durdurur** ve isteği başarısız kılar.
- `OnCacheWrite` için ilk hata dağıtımı durdurur ve yazmayı engeller.
- `OnCacheInvalidate` için ilk hata dağıtımı durdurur ve `InvalidateTags`'ten
  döndürülür.
- `OnError` için hatalar loglanır ve dağıtım devam eder.

## Şablon fonksiyonları

Bir plugin, şablon fonksiyonunu `Configure`'dan ekler:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	return host.AddTemplateFunc("readingTime", func(words int) string {
		return fmt.Sprintf("%d min read", max(1, words/200))
	})
}
```

Bundan sonra her şablon `{{readingTime .Words}}` çağırabilir. Fonksiyon,
`html/template`'in fonksiyon haritasında kabul ettiği herhangi bir değer olabilir.

- İki kez eklenen bir ad — iki plugin tarafından ya da aynı plugin tarafından iki
  kez — ikinci `AddTemplateFunc`'tan döndürülen `ErrDuplicateTemplateFunc`'tır.
  `Configure`'ınız onu döndürürse `New` başarısız olur; hatayı olduğu gibi ilettiğinizde
  de olan budur. Uygulama bunu `Template.Funcs` ile çözemez: çakışma plugin'ler
  arasındadır ve birinin geri çekilmesi gerekir.
- Bunun dışında **uygulama kazanır**. `Config.Template.Funcs` içinde, bir plugin'in
  eklediği bir ad altındaki girdi plugin'in fonksiyonunun yerini alır: uygulama
  ikisini de görebilir ve karar verebilir.
- Yerleşik bir ad altındaki plugin fonksiyonu yerleşik olanın yerini alır — render
  başına bağlanan fonksiyonlar (`slot`, `hoist`, `asset`, `stylesheet`, `csrfToken`,
  `pageURL`, `pageURLIn`, `localeURL`) hariç; render motoru onları her seferinde
  yeniden bağlar. Bkz. [Şablon fonksiyonları](/docs/template-functions).
- `Configure`'dan çağrılmalıdır. Sonradan eklemenin bir yolu yoktur, çünkü
  ayrıştırmadan sonra eklenen bir fonksiyonu hiçbir şablon çağıramaz.

## Mount'ları sarmalamak

`WrapMount`, mount edilen her dosya sistemine uygulanan bir fonksiyon kaydeder:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	host.WrapMount(func(fsys fs.FS) fs.FS {
		return minifyingFS{inner: fsys} // your own fs.FS
	})
	return nil
}
```

Yanıtı değil dosya sistemini sarmalar, çünkü mount'lar `Range`, `If-Range` ve
kısmi yanıtları destekleyen `http.ServeContent` üzerinden sunulur. Baytları yanıt
başına değiştirmek her ofseti kaydırır ve bir range isteği, ilan edilen uzunluğu
artık tutmayan bir dosyanın yanlış dilimini döndürür. Dosyaları dönüştürülmüş
olanların *ta kendisi* olan bir sarmalayıcı bu hesabı doğru tutar.
Sarmalayıcılar kaydedildikleri sırayla çalışır ve `nil` bir sarmalayıcı yok sayılır.

## Sayfa, document ve mount eklemek

Bir plugin `Init`'ten, bir uygulamanın kullandığı builder'larla kurulmuş kendi
route'larını `Host.RegisterPage`, `Host.RegisterDocument` ve `Host.Mount` ile
ekleyebilir.

*Dosya üreten* bir plugin — yeniden boyutlandırılmış görseller, üretilmiş ikonlar —
onları bir route'tan değil, bir mount'tan sunmalıdır. Statik build her sayfa render
edildikten sonra her mount'u çıktısına kopyalar; bu yüzden sayfaların ne istediğini
kaydeden bir dosya sistemi builder'a tam olarak doğru kümeyi verir ve dışa aktarılan
sitenin arkasında hiçbir şeyin çalışmasına gerek kalmaz. Dinamik bir yoldaki
document bu şekilde listelenemez.

## Komutlar

Bir plugin, `Init`'ten bir komut ekler:

```go
type Command struct {
	Name  string // as typed on the command line
	Usage string // for your program's own help; the framework never prints it
	Short string // one line
	Run   func(ctx context.Context, args []string) error
}
```

`RegisterCommand` boş bir adı (`ErrEmptyCommandName`) ve başka bir komutun zaten
sahip olduğu bir adı (`ErrDuplicateCommand`) reddeder. `ErrAppStarted` ile asla
kapanmaz: diğer `Host` kayıt çağrıları gibi `Init` sırasında çalışır, başlangıçtan
sonra da çalışır — ama `DispatchCommands` çalıştıktan sonra kaydedilen bir komutu
kimse dağıtmaz.

`Usage` ve `Short` veridir. Ne framework ne de `collage` binary'si onları yazdırır;
yardım listesi isteyen bir program onu `app.Commands()`'tan kurar.

`collage` CLI plugin komutlarını çalıştırmaz: uygulamanızı hiç yüklemez. Onları
uygulamanın kendi `main`'i `collage.DispatchCommands` ile dağıtır; iskelesi
oluşturulmuş bir `main.go` da flag'lerden sonra kalan her kelime için bunu yapar.
Böylece plugin'inizin bir kullanıcısı `go run . <command>` çalıştırır:

```go
flag.Parse()

// ... build app and register everything ...

if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
log.Fatal(app.ListenAndServe())
```

`DispatchCommands` önce uygulamayı başlatır, çünkü komutları kaydeden `Init`'tir.
Çıkış kodları: başarı için `0`; bir başlangıç hatası, çalışıp başarısız olan bir
komut ya da `Run`'ı olmayan bir komut için `1`; nil bir app, argüman olmaması ya da
kimsenin sahiplenmediği bir ad (`ErrUnknownCommand`) için `2`. Plugin'inizin
README'sinde komutlarının uygulama üzerinden çalıştığını belirtin — v0.10.0'dan
önce iskelesi oluşturulmuş bir projenin o bloğu kendisinin eklemesi gerekir. Bkz.
[collage CLI](/docs/cli#plugin-commands).

## Yapılandırma

Bir plugin, `Config.PluginConfig`'in kendi bölümünü her iki aşamada da kullanılabilen
`host.Config` ile tipli bir struct'a okur. Önce varsayılanlarınızı ayarlayın;
`Config` uygulamanın bölümünü onların üzerine çözer:

```go
type Config struct {
	Generator string `json:"generator"`
	Disabled  bool   `json:"disabled"`
}

func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	p.cfg = Config{Generator: "collage"} // defaults
	return host.Config(&p.cfg)           // overlaid by the application's section, if any
}
```

- **Bulunmayan bir bölüm `v`'yi olduğu gibi bırakır**; böylece "yapılandırılmamış"
  ile "sıfır değerine yapılandırılmış" farklı ifadeler olarak kalır.
- **Var olan ama bozuk bir bölüm bir hatadır.** Operatör bir şey yazmıştır ve onun
  yerine varsayılanlarla çalışmak, bunun reddettiği sessiz hata olurdu.
- Bölüm, varsayılanlarınızın üzerine `json.Unmarshal` ile çözülür ve onun
  kurallarına uyar. JSON'daki bir skaler ya da slice varsayılanınızın yerini alır —
  slice birleştirilmez. Bir haritaya çözülen JSON nesnesi, girdilerini sizin
  ayarladığınız haritaya ekler, diğerlerini korur. İç içe bir struct'a çözülen JSON
  nesnesi yalnızca adını verdiği alanları ayarlar, gerisini varsayılanlarınızda
  bırakır.
- Uygulama, kayıtlı hiçbir plugin'i adlandırmayan bir anahtarı başlangıç hatası
  olarak görür (`ErrUnknownPluginConfig`). Bu yüzden `Name`'iniz yapılandırmanızın
  adresinin tamamıdır; onu değiştirmek geriye dönük uyumluluğu bozan bir
  değişikliktir.

Her anahtarı, tipini ve varsayılanını README'nizde belgeleyin. `New()`'un yanında
bir `NewWith(Config)` constructor'ı sunmak, bir uygulamanın sizi Go'da da
yapılandırmasını sağlar.

## Eksiksiz bir plugin

`acme/stamp` her sayfanın head'inde üreticiyi (generator) adlandırır, aynı adı
şablonlara sunar, sayfaları listeleyen bir komut ekler ve başarısız olan bir hata
sayfasını bildirir. Her iki aşamayı, bir render hook'unu, bir error hook'unu ve bir
komutu kullanır.

```go
// Package stamp names the generator in every page's head, offers the same name
// to templates, and adds a command that lists the application's pages.
package stamp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key of its section in plugins-config.json.
const Name = "acme/stamp"

// Config is the plugin's configuration.
type Config struct {
	// Generator is what the generator meta tag says.
	Generator string `json:"generator"`
}

// Plugin is the stamp plugin. Construct it with New.
type Plugin struct {
	cfg    Config
	logger *slog.Logger
}

// New returns the plugin with its defaults.
func New() *Plugin { return &Plugin{} }

var (
	_ collage.Plugin           = (*Plugin)(nil)
	_ collage.Configurer       = (*Plugin)(nil)
	_ collage.BeforeRenderHook = (*Plugin)(nil)
	_ collage.ErrorHook        = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure runs inside collage.New, before templates are parsed: the one
// moment a template function can still be added.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	p.cfg = Config{Generator: "collage"} // the defaults
	if err := host.Config(&p.cfg); err != nil {
		return err // a section that is present but malformed
	}
	return host.AddTemplateFunc("generator", func() string { return p.cfg.Generator })
}

// Init runs when the application starts, after it has registered its pages.
func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	p.logger = host.Logger()
	return host.RegisterCommand(collage.Command{
		Name:  "pages",
		Usage: "pages",
		Short: "List every registered page",
		Run: func(_ context.Context, _ []string) error {
			for _, page := range host.Pages() {
				fmt.Println(page.Name)
			}
			return nil
		},
	})
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBeforeRender declares the meta tag before the page renders, at depth zero,
// so a fragment that declares its own generator replaces this one.
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}

// OnError reports the one failure nobody would otherwise notice: the error page
// itself failing.
func (p *Plugin) OnError(_ context.Context, ev *collage.ErrorEvent) error {
	if ev.Stage == "error_page" {
		p.logger.Error("stamp: the error page failed", "path", ev.Path, "error", ev.Err)
	}
	return nil
}
```

Bir uygulama onu diğer plugin'ler gibi kullanır:

```go
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{stamp.New()},
	PluginConfig: pluginConfig, // {"acme/stamp": {"generator": "The Wire"}}
})
```

## Yaşam döngüsü

1. **Kayıt.** `New` içinde `Config.Plugins` ya da uygulama başlamadan önce
   `RegisterPlugin`. Ondan sonra `RegisterPlugin` `ErrAppStarted` döner — başarısız
   olan herhangi bir başlatmadan sonra da (v0.12.0'dan itibaren).
2. **Configure**, `New` içinde, onu uygulayan plugin'ler için — kayıt sırasıyla,
   ilk hatada durarak.
3. **Init**, uygulama başladığında, kayıt sırasıyla. Biri başarısız olursa başlangıç
   iptal edilir ve zaten başlatılmış her plugin ters sırayla kapatılır. Başarısız
   olan plugin kapatılmaz, çünkü başlatılmasını hiç tamamlamamıştır.
4. **Shutdown**, `App.Shutdown`'dan — `ListenAndServe` onu `SIGINT` ya da `SIGTERM`
   üzerine çağırır — ters kayıt sırasıyla. **Kayıtlı her plugin'in** `Shutdown`'ını
   çağırır; o plugin'in `Init`'i çalışmış ya da başarılı olmuş olsun olmasın: hiç
   başlamamış bir uygulama, başlatılması başarısız olmuş bir uygulama ve başarısız
   başlatmanın zaten geri aldığı plugin'ler çağrıyı alır. Bu yüzden `Shutdown`,
   `Init` olmadan ve birden fazla kez çağrılmaya karşı güvenli olmalıdır. Biri
   başarısız olsa bile her plugin sırasını alır ve hatalar birleştirilir.

   `ListenAndServe` ile plugin'ler, sunucu isteklerini boşalttıktan sonra ya da
   boşaltmadıysa `Server.ShutdownTimeout` geçtikten sonra kapatılır — son süre
   geçtikten sonra bir istek hâlâ çalışıyor olabilir. Kendinize ait bir sunucuyla
   `App`'in haberi olan bir sunucu yoktur: önce sunucunuzu durdurun, sonra
   `App.Shutdown`'ı çağırın; yoksa bir plugin, süren bir isteğin altından
   çekilebilir.

## Bir plugin'i test etmek

Bir plugin'i, bir uygulamanın onu kullanacağı şekilde test edin: plugin
`Config.Plugins` içinde olan gerçek bir `App` kurun, şablonu onu çalıştıran bir
sayfa kaydedin ve uygulamayı `httptest` ile `app.Handler()` üzerinden sürün. Hiçbir
sunucu dinlemez ve hiçbir port seçilmez.

```go
package stamp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"

	"example.com/stamp"
)

// newApp builds a one-page application with the plugin in it.
func newApp(t *testing.T, config string) *collage.App {
	t.Helper()

	root := t.TempDir()
	page := `<html><head>{{hoist "head"}}</head><body>{{generator}}</body></html>`
	if err := os.MkdirAll(filepath.Join(root, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pages", "home.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &collage.Config{
		Template: collage.TemplateConfig{Root: root},
		Plugins:  []collage.Plugin{stamp.New()},
	}
	if config != "" {
		cfg.PluginConfig = map[string]json.RawMessage{stamp.Name: json.RawMessage(config)}
	}
	app, err := collage.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	home := collage.NewPage("home").
		WithContent(collage.NewFragment("home", "pages/home.html").Build()).
		WithPath("en", "/").
		Build()
	if err := app.RegisterPage(home); err != nil {
		t.Fatalf("RegisterPage: %v", err)
	}
	return app
}

func get(t *testing.T, app *collage.App, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d: %s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestStamp_UsesItsDefaults(t *testing.T) {
	body := get(t, newApp(t, ""), "/")
	if !strings.Contains(body, `<meta name="generator" content="collage">`) {
		t.Errorf("no default meta tag in %s", body)
	}
}

func TestStamp_ReadsItsConfiguration(t *testing.T) {
	body := get(t, newApp(t, `{"generator": "my site"}`), "/")
	if !strings.Contains(body, `<meta name="generator" content="my site">`) {
		t.Errorf("no configured meta tag in %s", body)
	}
	if !strings.Contains(body, "<body>my site</body>") {
		t.Errorf("template function not applied in %s", body)
	}
}

func TestStamp_RegistersItsCommand(t *testing.T) {
	app := newApp(t, "")
	if err := app.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	commands := app.Commands()
	if len(commands) != 1 || commands[0].Name != "pages" {
		t.Errorf("commands = %v, want one named pages", commands)
	}
}

func TestStamp_RefusesLateRegistration(t *testing.T) {
	app, err := collage.New(&collage.Config{Template: collage.TemplateConfig{Root: t.TempDir()}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := app.RegisterPlugin(stamp.New()); !errors.Is(err, collage.ErrConfigurerRegisteredLate) {
		t.Errorf("RegisterPlugin: err = %v, want ErrConfigurerRegisteredLate", err)
	}
}
```

Test etmeye değer, unutması kolay birkaç şey:

- **Yapılandırılmamış durum.** Çoğu uygulama sizin için hiçbir zaman bir bölüm
  yazmayacaktır.
- **Document'lar**, `OnCacheWrite` ya da `OnDocumentRendered` uyguluyorsanız: bir
  document kaydedin ve isteyin; böylece `nil` bir `Page`, hiç önbelleğe alınmayan
  bir document olarak değil, bir testte yakalanır.
- **`-race` ile çalıştırın.** Bir olayın canlı `Page`'i üzerinden yazan bir hook,
  dedektörün bildirdiği ve başka hiçbir şeyin bildirmeyeceği bir veri yarışıdır.
- **Statik dışa aktarma**, dosya üretiyorsanız: bir `t.TempDir()` içine
  `collage.NewBuilder(app, ...)`, dışa aktarılan sitenin sunulan sitede olan her
  şeye sahip olup olmadığını gösterir. Bkz. [Test](/docs/testing) ve
  [Statik dışa aktarma](/docs/static-export).
