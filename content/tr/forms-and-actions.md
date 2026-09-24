---
description: Form gönderimlerini, fetch çağrılarını ve webhook'ları action'larla karşılamak ve onları istek sahteciliğine karşı korumak.
---

# Formlar ve action'lar

Bir sayfa `GET` ve `HEAD` isteklerine yanıt verir. Geri kalan her şey — bir form
gönderimi, bir `fetch()` çağrısından gelen `DELETE`, bir ödeme sağlayıcısının
webhook'u — bir **action**'dır: isteği alan ve neyle yanıt verileceğini söyleyen bir
fonksiyon.

Hiçbir şeyin yanıt vermediği bir metotla gelen istek, hiçbir şey olmamış gibi render
edilen sayfayı değil, `Allow` header'ı taşıyan bir `405` alır; hata hook'ları bunu
`collage.ErrMethodNotAllowed` olarak görür. `OPTIONS` da aynı listeden yanıtlanır.

## Bir sayfadaki action

Olağan durum, üzerinde bulunduğu sayfaya gönderilen bir formdur. Sayfaya o metot
için bir action verin:

```go
collage.NewPage("contact").
	WithLayout(layout).
	WithContent(contactForm).
	WithPath("en", "/contact").
	WithAction("POST", sendMessage).
	Build()
```

```html
<form method="post" action="{{pageURL "contact"}}">
  {{csrfToken}}
  <input type="email" name="email">
  <textarea name="message"></textarea>
  <button type="submit">Send</button>
</form>
```

Action, sayfanın yollarını sayfanın bildirdiği her locale'de devralır; böylece form
hem `/contact`'ta hem `/tr/iletisim`'de aynı şekilde çalışır.

Bir action handler'ı şu biçimdedir:

```go
type ActionHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error)
```

`rc.Request` istektir ve gövdesi zaten sınırlandırılmıştır (bkz.
[İstek gövdeleri](#request-bodies-are-bounded)); `rc.Locale` ve `rc.Param` bir data
handler'da nasıl çalışıyorsa öyle çalışır.

`WithAction`, tek bir metot için bir kısaltmadır. Birden fazla metot, kendine ait bir
gövde sınırı ya da sahtecilik denetiminin olmaması için action'ı `NewAction` ile
kurun ve `WithActionFor` ile bağlayın — action'ın kendi yolları sayfanınkilerle
değiştirilir:

```go
collage.NewPage("post").
	WithContent(post).
	WithPath("en", "/posts/{slug}").
	WithActionFor(collage.NewAction("post-edit").
		WithMethods(http.MethodPut, http.MethodDelete).
		WithMaxBodyBytes(64 << 10).
		WithHandler(editPost).
		Build()).
	Build()
```

## Kendi URL'sindeki action

Bir sayfayla ilgili olmayan bir action — bir JSON endpoint'i, bir webhook — tek
başına kaydedilir:

```go
err := app.RegisterAction(collage.NewAction("like").
	WithPath("en", "/api/like").
	WithMethods(http.MethodPost).
	WithHandler(like).
	Build())
```

Bir action bir sayfayla aynı yolu paylaşabilir — `WithAction`'ın yaptığı tam olarak
budur — ama iki action aynı yolda aynı metoda yanıt veremez. Adı, yolu, metodu ya da
handler'ı olmayan bir action'ı `RegisterAction` reddeder; zaten alınmış bir adla
gelen ikinci bir action'ı da. Her birinin kendi hatası vardır ve hepsi
[Hatalar](/docs/errors#actions) sayfasında listelenir.

## Action neyle yanıt verir

Yanıtı bir `ActionResult` belirler. Onu bir yardımcı fonksiyonla oluşturun:

| Yardımcı | Yanıt |
| --- | --- |
| `collage.SeeOther(url)` | `url`'ye `303 See Other` |
| `collage.RenderPage(page)` | Bu isteğin `RenderContext`'iyle render edilmiş bütün bir sayfa |
| `collage.RenderFragment(f)` | Layout olmadan, tek bir fragment'in işaretlemesi |
| `collage.JSON(status, body)` | `application/json` olarak `body` |
| `collage.JSONOf(status, v)` | `application/json` olarak marshal edilmiş `v`; `(*ActionResult, error)` döner |
| `collage.NoContent(status)` | Yalnızca bir durum kodu, gövde yok |

Ya da struct'ı kendiniz doldurun. Struct'ın gövde üretmenin dört yolu vardır —
`Location`, `Fragment`, `Page` ve `Body` — ve bu sırayla, ilk ayarlanan kazanır.
`Status`, her birinin seçeceği durum kodunun yerine geçer; `Header` yanıta yazılır
(`Set-Cookie`'nin yeri burasıdır) ve `ContentType`, `Body` ile birlikte kullanılır.
Boş bir `ContentType` `application/octet-stream` demektir; asla byte'lardan tahmin
edilmez. `nil` bir sonuç `204`'tür.

```go
func like(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	n, err := store.Like(ctx, rc.Request.PostFormValue("post"))
	if err != nil {
		return nil, err
	}
	return collage.JSONOf(http.StatusOK, likeResponse{Count: n})
}
```

Handler'dan dönen bir hata `500`'dür; `collage.ErrNotFound`'u sarmalıyorsa `404`.
Bir action'ın yanıtı, header'ı kendiniz ayarlamadıkça `Cache-Control: no-store` ile
gönderilir ve asla sayfa önbelleğine girmez: tek bir gönderimden, onu gönderen kişi
için üretilmiştir.

## Başarısızlık render eder, başarı yönlendirir

`RenderPage` da `SeeOther` da bir forma verilebilecek iyi yanıtlardır. İkisi
arasındaki seçimi, tarayıcının ardından ne yapacağı belirler.

Bir POST'a sayfayla yanıt vermek, adres çubuğunu gönderilen URL'de ve geçmiş
kaydını bir POST olarak bırakır; bu yüzden sayfayı yenilemek formu yeniden gönderir.
`303` ile yanıt vermek ise tarayıcının bir sonraki isteğini başka bir yere yapılan bir
`GET` yapar ve *onu* yenilemek zararsızdır. Dolayısıyla:

**Reddedilen bir gönderim sayfayı render eder**, `422` durum koduyla. Okuyucu
düzeltip yeniden gönderecektir — amaç zaten yeniden göndermektir — ve reddin nedenini
ve yazdıklarını okuyucunun önüne geri getirmenin yolu render etmektir.

**Kabul edilen bir gönderim yönlendirir**, `303` ile. İş yapılmıştır ve bir yenileme
onu ikinci kez yapmamalıdır.

### Doğrulama için yeniden render

Handler ve yanıt olarak verdiği sayfa tek bir `RenderContext` paylaşır. Handler'ın
oraya `rc.Set` ile koyduğu her şeyi sayfanın data handler'ları `collage.Get` ile
okuyabilir. Oturum yok, flash mesajı yok, query string'de hiçbir şey yok: ikisi aynı
istektir.

```go
type contactView struct {
	Error string
	Email string
}

func ContactPage() *collage.Page {
	form := collage.NewFragment("contact-form", "pages/contact.html").
		WithDataHandler(collage.DataHandler(contactData)).
		Build()

	var page *collage.Page
	page = collage.NewPage("contact").
		WithLayout(layouts.Layout()).
		WithContent(form).
		WithPath("en", "/contact").
		WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			email := strings.TrimSpace(rc.Request.PostFormValue("email"))
			if !strings.Contains(email, "@") {
				rc.Set("contact:form", contactView{Error: "That does not look like an email address.", Email: email})
				result := collage.RenderPage(page)
				result.Status = http.StatusUnprocessableEntity
				return result, nil
			}
			if err := messages.Send(ctx, email, rc.Request.PostFormValue("message")); err != nil {
				return nil, err
			}
			return collage.SeeOther("/contact/thanks"), nil
		}).
		Build()
	return page
}

func contactData(_ context.Context, rc *collage.RenderContext) (contactView, []string, error) {
	// Set by the action when this render is its answer; empty on an ordinary GET.
	view, _ := collage.Get[contactView](rc, "contact:form")
	return view, nil, nil
}
```

```html
<form method="post" action="{{pageURL "contact"}}">
  {{csrfToken}}
  {{with .Error}}<p class="error">{{.}}</p>{{end}}
  <input type="email" name="email" value="{{.Email}}">
  <textarea name="message"></textarea>
  <button type="submit">Send</button>
</form>
```

Sayfa, her sayfa gibi, render hook'ları dahil render edilir: bir plugin'in
`OnAfterRender`'ı — bir küçültücü (minifier), bir görsel yeniden yazıcı — doğrulama
sayfasını da bir `GET`'in aldığı sayfayı biçimlendirdiği gibi biçimlendirir. Bkz.
[Plugin yazmak](/docs/writing-plugins#afterrenderhook).

Yanıt olarak verdiğiniz sayfa, `app.RegisterPage` ile **kaydettiğiniz değer**
olmalıdır. Bir sayfanın içeriğini layout'una yerleştiren şey kayıttır; bu yüzden
handler içinde kurulan bir sayfa, hiçbir şeyin etrafında duran bir layout olarak
render edilirdi. collage böyle bir sayfayı, adını vererek `collage.ErrUnregisteredPage`
ile reddeder. Yukarıdaki `page` üzerindeki closure, action'a kendi sayfasını vermenin
en basit yoludur.

Bir sayfaya yolu yerine adıyla yönlendirmek için
[`app.URL`](/docs/links-and-locales#links-from-go) kullanın.

## Sahteciliğe karşı koruma

Güvenli olmayan bir metotla (`GET`, `HEAD` ve `OPTIONS` dışındaki her şey) bir
action'a gelen her istek, handler çalışmadan önce bir sahtecilik token'ı açısından
denetlenir. Geçerli bir token taşımayan istek `403` alır ve handler onu hiç görmez.
Hata hook'ları nedenini görür: token ya da çerez yoksa `collage.ErrCSRFMissing`,
token çerezinkiyle eşleşmiyorsa `collage.ErrCSRFMismatch`, token'ı bu uygulama
imzalamamışsa `collage.ErrCSRFInvalid`.

### Formda

`{{csrfToken}}`'ı `<form>`'un içine koyun. Alan adı dahil gizli input'un tamamını
render eder; yani yanlış yapılabilecek bir şey kalmaz:

```html
<form method="post">
  {{csrfToken}}
  ...
</form>
```

Şema, imzalı bir double-submit çerezidir: token rastgele bir değer ve bu değerin
`Security.CSRFKey` ile atılmış imzasıdır; hem bir çerezde hem de formda gönderilir
ve denetlemek için anahtardan başka hiçbir şey gerekmez. Oturum deposu yok, örnekler
(instance) arasında paylaşılan hiçbir şey yok.

### Formlu sayfalar yine de önbelleğe alınır

Bir token tek bir okuyucuya aittir, önbellekteki bir sayfa ise herkesle paylaşılır.
Bu yüzden `{{csrfToken}}` bir token render etmez: bir işaretçi render eder ve
önbelleğin sakladığı şey bu işaretçidir. Her yanıtta işaretçi, çıkış sırasında o
okuyucunun kendi token'ıyla değiştirilir ve yanıt, token'ın çereziyle birlikte
`private, no-store` olarak gönderilir. Render paylaşılır; okuyucuya özgü tek dize
paylaşılmaz.

Bu olmasaydı, bir sitenin alt bilgisindeki bir bülten formu bütün sitede önbelleği
kapatırdı. Değiştirdiği tek şey şu: formlu bir sayfa statik bir dosyaya
dönüştüğünde formu gönderecek bir sunucusu kalmaz; bu yüzden
[statik dışa aktarma](/docs/static-export) böyle bir sayfayı atlar ve nedenini söyler.

### Anahtarı ayarlayın

```go
Security: collage.SecurityConfig{
	CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")), // e.g. openssl rand -hex 32
},
```

Boş bırakırsanız başlangıçta bir anahtar üretilir. collage bunu söyler — production'da
bir uyarıyla, geliştirmede düz bir log satırıyla — ama yalnızca uygulamada güvenli
olmayan bir metodu kabul eden ve token doğrulayan bir action varsa; böyle bir action'ı
olmayan ya da bu tür action'larının hepsi `WithoutCSRF` ile muaf tutulmuş bir uygulama
hiçbir token denetlemez ve anahtar konusunda uyarılmaz. Bu, ilk çalıştırma için
sorun değildir ama yayına almak için yanlıştır: üretilen anahtar her süreçte
farklıdır; bu yüzden yeniden başlatmadan önce yüklenen bir form sonrasında reddedilir,
bir örneğin sunduğu form da bir sonrakinde reddedilir. Daha önceki bir anahtarla
saklanmış, form içeren önbellekteki bir sayfa, hiçbir şeyin doğrulayamayacağı bir
token'la sunulmak yerine yeniden render edilir — bkz.
[disk önbelleğinin namespace'i](/docs/caching#the-namespace).

### fetch'ten

Form olmadan yapılan bir `fetch()`, token'ı `X-CSRF-Token` header'ında gönderir.
Token'ı sayfadaki herhangi bir `{{csrfToken}}` input'undan okuyun — sayfa
geldiğinde bu input okuyucunun kendi token'ını taşır. (`_csrf` alanın varsayılan
adıdır; değiştirmek için [aşağıya](#requests-that-cannot-carry-a-token) bakın.)

```js
const token = document.querySelector('input[name="_csrf"]')?.value ?? "";
await fetch("/api/like", {
  method: "POST",
  headers: { "X-CSRF-Token": token },
  body: new URLSearchParams({ post: "hello-world" }),
});
```

Form gönderen bir `fetch()` için ek bir şey gerekmez. `new FormData(form)`
`multipart/form-data` olarak gönderilir ve collage bunu URL-encoded bir gövde gibi
okur; böylece gizli input diğer alanlarla birlikte gider:

```js
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(form.action, { method: "POST", body: new FormData(form) });
  // The action answered with RenderFragment: the form's own fragment, re-rendered.
  form.outerHTML = await response.text();
});
```

Bir action'ın yanıt olarak verdiği fragment ya da sayfa diğerleri gibi render edilir;
dolayısıyla içindeki form da taze bir token taşır: kendini yenisiyle değiştiren bir
form çalışmaya devam eder.

### Token taşıyamayan istekler

Bir ödeme sağlayıcısının webhook'u ya da tarayıcı olmayan bir şeyin bearer token ile
çağırdığı bir API token taşıyamaz. Denetimi yalnızca o action için kapatın:

```go
app.RegisterAction(collage.NewAction("stripe-webhook").
	WithPath("en", "/hooks/stripe").
	WithMethods(http.MethodPost).
	WithoutCSRF().
	WithHandler(onStripe).
	Build())
```

Ardından isteğin kimliğini başka bir yolla doğrulayın — bir webhook'un imza header'ı,
bir bearer token. Bir tarayıcının gönderdiği herhangi bir şeyde `WithoutCSRF()`
korumayı elden çıkarmak demektir. (`Security.DisableCSRF` korumayı bütün uygulama
için kapatır; bu yalnızca hiç tarayıcı formu olmayan bir uygulama için doğrudur.)

Çerez, alan ve header adları `Security.CSRFCookieName`, `CSRFFieldName` ve
`CSRFHeaderName` ile değiştirilebilir. Adı değiştirilen alan her iki tarafta da
değişir: `{{csrfToken}}` denetimin okuduğu adı render eder, dolayısıyla formda
değişiklik gerekmez.

## İstek gövdeleri sınırlıdır

Bir action'ın istek gövdesi varsayılan olarak **4 MiB** ile sınırlıdır. Bunu uygulama
için `Server.MaxBodyBytes` ile, tek bir action için `WithMaxBodyBytes` ile
değiştirin; negatif bir değer sınır yok demektir.

Sınırı handler'ınız uygulamaz; sınır, handler'ınız çalışmadan önce uygulanır. Her
handler'ın hatırlaması gereken bir sınır, unutan handler'da bulunmayan bir sınırdır —
ve anonim bir çağıranın bulacağı handler da tam olarak odur. Sınırın ötesini okuyan
bir handler bir `*http.MaxBytesError` alır; bu hatayı döndürmek `413` ile yanıt
verir:

```go
if err := rc.Request.ParseForm(); err != nil {
	return nil, err // a body over the limit becomes a 413
}
```

Token formdan denetlendiğinde denetim önce gövdeyi okur; bu yüzden aşırı büyük bir
form orada, handler'ınız çalışmadan önce reddedilir — yine `403` ile değil, `413`
ile: okunamayacak kadar büyük bir gövde sahtecilik değildir.

## Bir action'ın değiştirdiğini geçersiz kılmak

İçeriği değiştiren bir action genellikle önbellekteki bazı sayfaları yanlış hâle
getirir. Hangileri olduğunu sonuçta belirtin:

```go
func publish(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	slug := rc.Param("slug")
	if err := store.Publish(ctx, slug); err != nil {
		return nil, err
	}
	result := collage.SeeOther("/posts/" + slug)
	result.InvalidateTags = []string{"post:" + slug, "blog:posts"}
	return result, nil
}
```

Etiketler yanıt yazılmadan **önce** geçersiz kılınır. Asıl mesele bu sıradır: okuyucu
yönlendirmeyi izleyerek az önce değiştirdiği sayfaya gider ve o sayfa, değişikliğin
bayatlattığı bir önbellek girdisinden sunulmamalıdır. Geçersiz kılma başarısız
olursa yanıt başarısız olmaz, hata loglanır — değişiklik zaten gerçekleşmiştir.

Etiketler tam olarak [Önbellekleme](/docs/caching#dependency-tags) sayfasında
anlatıldığı gibi çalışır ve `collage.Cached` ile saklanan değerlere de ulaşır.

## Kendi URL'sindeki fragment

Bir istemci framework'ü olmadan sayfanın bir kısmını yenilemek iki şey gerektirir:
yalnızca o kısımla yanıt veren bir URL ve onu yerine koyacak birkaç satır JavaScript.
Birincisi `WithFragmentPath`'tir:

```go
collage.NewPage("search").
	WithLayout(layout).
	WithContent(searchContent).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Dynamic().
	Build()
```

`GET /search/results?q=grid` `results` fragment'ini render eder, başka hiçbir şeyi
değil. Data handler'ı çalışır, kendi slot'ları doldurulur ve hata politikası
uygulanır — aynı render'dır, yalnızca daha aşağıdan başlatılmıştır. Etrafında layout
olmadığından hoist ettiklerinin gidecek bir yeri yoktur ve yanıtı önbelleğe alınmaz.

```js
const input = document.querySelector('input[name="q"]');
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Bildirilmemiş hiçbir şeye erişilemez. Her fragment'i otomatik olarak dışa açan bir
framework, her sayfanın her iç parçasını herkese açık web'e koymuş olurdu.

`RenderFragment` ile yanıt veren bir action'la birleştiğinde, bir form gönderilebilir
ve yalnızca değişen kısımla yanıtlanabilir:

```go
WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	if err := store.AddComment(ctx, rc.Param("slug"), rc.Request.PostFormValue("body")); err != nil {
		return nil, err
	}
	result := collage.RenderFragment(comments)
	result.InvalidateTags = []string{"comments:" + rc.Param("slug")}
	return result, nil
})
```

Fragment'in data handler'ı action'ın `RenderContext`'iyle çalışır; böylece az önce
eklenen yorumu — ve handler'ın oraya `rc.Set` ile koyduğu her şeyi — görür.
