---
description: Form post'larını, fetch çağrılarını ve webhook'ları action'larla karşılamak ve onları request forgery'ye karşı korumak.
reference: NewAction, ActionBuilder, ActionResult, SeeOther, JSONOf, RenderPage, RenderFragment, PageBuilder.WithAction
---

# Form'lar ve action'lar

Bir page `GET` ve `HEAD` request'lerine cevap verir. Geri kalan her şey bir
**action**'dır: bir form post'u, bir `fetch()` çağrısından gelen `DELETE`, bir ödeme
sağlayıcısının webhook'u. Action, request'i alan ve neyle cevap verileceğini
söyleyen bir fonksiyondur.

Hiçbir şeyin cevap vermediği bir method ile gelen request, `Allow` header'ı taşıyan
bir `405` alır. Page, hiçbir şey olmamış gibi render edilmez. Error hook'ları bu
durumu `collage.ErrMethodNotAllowed` olarak görür. `OPTIONS` da aynı listeden
cevaplanır.

## Page üzerinde bir action

Olağan durum, bulunduğu page'e post eden bir form'dur. Page'e o method için bir
action verin:

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

Action, page'in path'lerini page'in tanımladığı her locale'de devralır. Böylece form
hem `/contact`'ta hem de `/tr/iletisim`'de aynı şekilde çalışır.

Bir action handler'ı şu şekildedir:

```go
type ActionHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error)
```

`rc.Request` request'in kendisidir ve body'si zaten sınırlandırılmıştır (bkz.
[Request body'leri](#request-bodies-are-bounded)). `rc.Locale` ve `rc.Param`, bir
data handler'da nasıl çalışıyorsa burada da öyle çalışır.

`WithAction` tek bir method için bir kısayoldur. Birden fazla method, action'a özel
bir body limiti ya da forgery kontrolünün kapalı olması gerekiyorsa action'ı
`NewAction` ile oluşturun ve `WithActionFor` ile bağlayın. Bu durumda action'ın
kendi path'lerinin yerini page'in path'leri alır:

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

## Kendi URL'sinde bir action

Bir page ile ilgisi olmayan bir action, örneğin bir JSON endpoint'i ya da bir
webhook, tek başına register edilir:

```go
err := app.RegisterAction(collage.NewAction("like").
	WithPath("en", "/api/like").
	WithMethods(http.MethodPost).
	WithHandler(like).
	Build())
```

Bir action bir page ile aynı path'i paylaşabilir. `WithAction`'ın yaptığı da tam
olarak budur. Ancak iki action aynı path'te aynı method'a cevap veremez.
`RegisterAction`; adı, path'i, method'u ya da handler'ı olmayan bir action'ı
reddeder. Daha önce alınmış bir adla gelen ikinci bir action'ı da reddeder. Her
birinin kendi hatası vardır ve hepsi [Hatalar](/docs/errors#actions) sayfasında
listelenir.

## Action neyle cevap verir

Response'u bir `ActionResult` belirler. Onu bir helper ile oluşturun:

| Helper | Response |
| --- | --- |
| `collage.SeeOther(url)` | `url`'e `303 See Other` |
| `collage.RenderPage(page)` | Bu request'in `RenderContext`'i ile render edilmiş bütün bir page |
| `collage.RenderFragment(f)` | Tek bir fragment'in markup'ı, layout olmadan |
| `collage.JSON(status, body)` | `application/json` olarak `body` |
| `collage.JSONOf(status, v)` | `application/json` olarak marshal edilmiş `v`; `(*ActionResult, error)` döner |
| `collage.NoContent(status)` | Sadece bir status, body yok |

Struct'ı kendiniz de doldurabilirsiniz. Struct'ın body üretmek için dört alanı
vardır: `Location`, `Fragment`, `Page` ve `Body`. Bu sırayla bakılır ve set edilen
ilk alan kazanır. `Status`, her birinin seçeceği status'un yerine geçer. `Header`
response'a yazılır; `Set-Cookie` için doğru yer burasıdır. `ContentType` ise
`Body` ile birlikte kullanılır. Boş bir `ContentType`, `application/octet-stream`
anlamına gelir; byte'lara bakılarak asla tahmin edilmez. `nil` bir result `204`
demektir.

```go
func like(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	n, err := store.Like(ctx, rc.Request.PostFormValue("post"))
	if err != nil {
		return nil, err
	}
	return collage.JSONOf(http.StatusOK, likeResponse{Count: n})
}
```

Handler'ın döndürdüğü bir hata `500` olur. Hata `collage.ErrNotFound`'u wrap
ediyorsa `404` olur. Header'ı kendiniz set etmediğiniz sürece action'ın response'u
`Cache-Control: no-store` ile gönderilir ve asla page cache'e girmez. Çünkü bu
response tek bir gönderimden, onu gönderen kişi için üretilmiştir.

## Başarısızlık render eder, başarı redirect eder

Bir form'a hem `RenderPage` hem de `SeeOther` ile cevap vermek doğrudur. Hangisini
seçeceğinizi, tarayıcının bundan sonra ne yapacağı belirler.

Bir POST'a page ile cevap verirseniz adres çubuğu post edilen URL'de kalır ve
history kaydı bir POST olur. Bu yüzden sayfayı yenilemek form'u tekrar gönderir.
`303` ile cevap verirseniz tarayıcının sonraki request'i başka bir yere yapılan bir
`GET` olur ve *onu* yenilemek zararsızdır. Dolayısıyla:

**Reddedilen bir gönderim, `422` status'u ile page'i render eder.** Okuyucu
hatayı düzeltip form'u tekrar gönderecektir; zaten amaç tekrar göndermesidir.
Reddin nedenini ve okuyucunun yazdıklarını tekrar önüne getirmenin yolu render
etmektir.

**Kabul edilen bir gönderim, `303` ile redirect eder.** İş yapılmıştır ve sayfayı
yenilemek onu ikinci kez yapmamalıdır.

### Validation için yeniden render

Handler ve cevap olarak verdiği page aynı `RenderContext`'i paylaşır. Handler'ın
oraya `rc.Set` ile koyduğu her şeyi page'in data handler'ları `collage.Get` ile
okuyabilir. Session yok, flash message yok, query string'de hiçbir şey yok: ikisi
de aynı request'tir.

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

Bu page de her page gibi, render hook'ları dahil render edilir. Bir plugin'in
`OnAfterRender`'ı, örneğin bir minifier ya da bir image rewriter, bir `GET`
request'inin aldığı page'i nasıl şekillendiriyorsa validation page'ini de öyle
şekillendirir. Bkz. [Plugin yazmak](/docs/writing-plugins#afterrenderhook).

Cevap olarak verdiğiniz page, `app.RegisterPage` ile **register ettiğiniz değerin
kendisi** olmalıdır. Bir page'in içeriğini layout'una yerleştiren şey register
işlemidir. Bu yüzden handler içinde oluşturulan bir page, içi boş bir layout olarak
render edilirdi. collage böyle bir page'i, page'in adını vererek
`collage.ErrUnregisteredPage` ile reddeder. Yukarıdaki örnekte `page` üzerinden
kurulan closure, action'a kendi page'ini vermenin en basit yoludur.

Bir page'e path yerine adıyla redirect etmek için
[`app.URL`](/docs/links-and-locales#links-from-go) kullanın.

## Forgery koruması

Unsafe bir method ile (`GET`, `HEAD` ve `OPTIONS` dışındaki her şey) bir action'a
gelen her request, handler çalışmadan önce bir forgery token'ı için kontrol edilir.
Geçerli bir token taşımayan request `403` alır ve handler onu hiç görmez. Error
hook'ları nedenini de görür. Token ya da cookie yoksa `collage.ErrCSRFMissing`
gelir. Token cookie'deki ile eşleşmiyorsa `collage.ErrCSRFMismatch`, token'ı bu
uygulama imzalamamışsa `collage.ErrCSRFInvalid` gelir.

### Form içinde

`{{csrfToken}}`'ı `<form>`'un içine koyun. Bu, field adı dahil hidden input'un
tamamını render eder. Yani yanlış yapabileceğiniz bir şey kalmaz:

```html
<form method="post">
  {{csrfToken}}
  ...
</form>
```

Kullanılan yöntem imzalı bir double-submit cookie'dir. Token, rastgele bir değer ile
bu değerin `Security.CSRFKey` ile atılmış imzasından oluşur. Token hem bir cookie'de
hem de form'da gönderilir. Kontrol etmek için key dışında hiçbir şey gerekmez. Session
store yoktur, instance'lar arasında paylaşılan hiçbir şey yoktur.

### Form içeren page'ler yine de cache'lenir

Bir token tek bir okuyucuya aittir, cache'lenmiş bir page ise herkes tarafından
paylaşılır. Bu yüzden `{{csrfToken}}` bir token render etmez. Bunun yerine bir marker
render eder ve cache'in sakladığı şey bu marker'dır. Her response'ta marker, çıkışta
o okuyucunun kendi token'ı ile değiştirilir. Response, token'ın cookie'si ile birlikte
`private, no-store` olarak gönderilir. Render paylaşılır, okuyucuya özel tek string
paylaşılmaz.

Bu olmasaydı, bir sitenin footer'ındaki bir bülten form'u bütün sitede cache'i
kapatırdı. Bunun değiştirdiği tek şey şudur: form içeren bir page static bir dosyaya
dönüştüğünde form'u gönderebileceği bir sunucu kalmaz. Bu yüzden
[static export](/docs/static-export) böyle bir page'i atlar ve nedenini söyler.

### Key'i ayarlayın

```go
Security: collage.SecurityConfig{
	CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")), // e.g. openssl rand -hex 32
},
```

Boş bırakırsanız başlangıçta bir key üretilir. Uygulamada unsafe bir method kabul
eden ve token doğrulayan bir action varsa collage bunu bildirir: production'da bir
uyarı, development'ta düz bir log satırı olarak. Böyle bir action'ı olmayan ya da bu
tür action'larının hepsi `WithoutCSRF` ile muaf tutulmuş bir uygulama hiçbir token'ı
kontrol etmez ve key hakkında uyarılmaz. Üretilen key ilk çalıştırma için sorun
değildir, ama deploy için yanlıştır. Üretilen key her process'te farklıdır. Bu yüzden
restart'tan önce yüklenen bir form restart'tan sonra reddedilir. Bir instance'ın
sunduğu form da bir sonraki instance tarafından reddedilir. Form içeren ve daha önceki
bir key ile saklanmış cache'lenmiş bir page, hiçbir şeyin doğrulayamayacağı bir token
ile sunulmaz, yeniden render edilir. Bkz.
[disk cache'inin namespace'i](/docs/caching#the-namespace).

### fetch ile

Form olmadan yapılan bir `fetch()`, token'ı `X-CSRF-Token` header'ında gönderir.
Token'ı page'deki herhangi bir `{{csrfToken}}` input'undan okuyun. Page tarayıcıya
ulaştığında bu input okuyucunun kendi token'ını taşır. (`_csrf` field'ın varsayılan
adıdır. Nasıl değiştirileceği için [aşağıya](#requests-that-cannot-carry-a-token)
bakın.)

```js
const token = document.querySelector('input[name="_csrf"]')?.value ?? "";
await fetch("/api/like", {
  method: "POST",
  headers: { "X-CSRF-Token": token },
  body: new URLSearchParams({ post: "hello-world" }),
});
```

Bir form'u post eden `fetch()` için ekstra bir şey gerekmez. `new FormData(form)`
`multipart/form-data` olarak gönderilir ve collage bunu URL-encoded bir body gibi
okur. Böylece hidden input diğer field'larla birlikte gider:

```js
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(form.action, { method: "POST", body: new FormData(form) });
  // The action answered with RenderFragment: the form's own fragment, re-rendered.
  form.outerHTML = await response.text();
});
```

Bir action'ın cevap olarak verdiği fragment ya da page, diğerleri gibi render edilir.
Dolayısıyla içindeki form da yeni bir token taşır. Kendini değiştiren bir form bu
sayede çalışmaya devam eder.

### Token taşıyamayan request'ler

Bir ödeme sağlayıcısının webhook'u ya da tarayıcı olmayan bir şeyin bearer token ile
çağırdığı bir API token taşıyamaz. Kontrolü sadece o action için kapatın:

```go
app.RegisterAction(collage.NewAction("stripe-webhook").
	WithPath("en", "/hooks/stripe").
	WithMethods(http.MethodPost).
	WithoutCSRF().
	WithHandler(onStripe).
	Build())
```

Ardından request'i başka bir yolla doğrulayın: bir webhook'un signature header'ı ya
da bir bearer token ile. Tarayıcının gönderdiği herhangi bir şeyde `WithoutCSRF()`
korumayı elden vermek demektir. (`Security.DisableCSRF` korumayı bütün uygulama için
kapatır. Bu sadece hiç tarayıcı form'u olmayan bir uygulama için doğrudur.)

Cookie, field ve header adları `Security.CSRFCookieName`, `CSRFFieldName` ve
`CSRFHeaderName` ile değiştirilebilir. Field'ın adını değiştirirseniz ad iki tarafta
birden değişir. `{{csrfToken}}` kontrolün okuduğu adı render eder, dolayısıyla
form'da bir değişiklik gerekmez.

## Request body'leri sınırlıdır

Bir action'ın request body'si varsayılan olarak **4 MiB** ile sınırlıdır. Bu limiti
uygulamanın tamamı için `Server.MaxBodyBytes` ile, tek bir action için
`WithMaxBodyBytes` ile değiştirin. Negatif bir değer limit olmadığı anlamına gelir.

Limiti handler'ınız uygulamaz, limit handler'ınız çalışmadan önce uygulanır. Her
handler'ın hatırlaması gereken bir limit, unutan handler'da yoktur. Anonim bir
kullanıcının bulacağı handler da tam olarak o handler'dır. Limitin ötesini okuyan
bir handler `*http.MaxBytesError` alır. Bu hatayı döndürürseniz response `413`
olur:

```go
if err := rc.Request.ParseForm(); err != nil {
	return nil, err // a body over the limit becomes a 413
}
```

Token form'dan kontrol edildiğinde kontrol önce body'yi okur. Bu yüzden limiti aşan
bir form daha orada, handler'ınız çalışmadan önce reddedilir. Bu durumda da cevap
`403` değil `413` olur, çünkü okunamayacak kadar büyük bir body forgery değildir.

## Action'ın değiştirdiğini invalidate etmek

İçeriği değiştiren bir action genellikle cache'lenmiş bazı page'leri yanlış hâle
getirir. Hangileri olduğunu result üzerinde belirtin:

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

Tag'ler response yazılmadan **önce** invalidate edilir. Bu sıra önemlidir. Okuyucu
redirect'i takip ederek az önce değiştirdiği page'e gider. Bu page, değişikliğin
stale hâle getirdiği bir cache kaydından sunulmamalıdır. Invalidate işlemi başarısız
olursa response başarısız olmaz, hata log'lanır. Çünkü değişiklik zaten
gerçekleşmiştir.

Tag'ler tam olarak [Caching](/docs/caching#dependency-tags) sayfasında anlatıldığı
gibi çalışır ve `collage.Cached` ile saklanan değerlere de ulaşır.

## Kendi URL'sinde bir fragment

Bir client framework'ü olmadan page'in bir kısmını yenilemek için iki şey gerekir:
sadece o kısımla cevap veren bir URL ve o kısmı yerine koyan birkaç satır
JavaScript. Birincisini `WithFragmentPath` sağlar:

```go
collage.NewPage("search").
	WithLayout(layout).
	WithContent(searchContent).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Dynamic().
	Build()
```

`GET /search/results?q=grid` sadece `results` fragment'ini render eder, başka hiçbir
şeyi render etmez. Fragment'in data handler'ı çalışır, kendi slot'ları doldurulur ve
failure policy'si uygulanır. Bu aynı render'dır, sadece daha aşağıdan başlar.
Etrafında bir layout olmadığı için hoist ettiği şeylerin gidecek bir yeri yoktur ve
response'u cache'lenmez.

```js
const input = document.querySelector('input[name="q"]');
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Tanımlanmamış hiçbir şeye erişilemez. Her fragment'i otomatik olarak dışarı açan bir
framework, her page'in her iç parçasını public web'e açmış olurdu.

`RenderFragment` ile cevap veren bir action ile birleştirildiğinde bir form post
edilebilir ve sadece değişen kısımla cevaplanabilir:

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

Fragment'in data handler'ı action'ın `RenderContext`'i ile çalışır. Böylece az önce
eklenen yorumu görür. Handler'ın oraya `rc.Set` ile koyduğu her şeyi de görür.
