---
description: Form post'larını, fetch çağrılarını ve webhook'ları action'larla karşılamak ve onları request forgery'ye karşı korumak.
reference: NewAction, ActionBuilder, ActionResult, SeeOther, SeeOtherPage, JSON, JSONBytes, Status, RenderPage, RenderFragment, PageBuilder.WithAction, PageBuilder.WithFragmentPath, ActionBuilder.WithStreamingBody, ActionBuilder.WithBodyTimeout, ErrCSRFHeaderRequired, ErrDuplicateRoute, ErrUnknownFragmentPath, FetchHeader, LocationHeader
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

`OPTIONS`, `TRACE` ve `CONNECT` bir action'a değil, sunucuya aittir. v0.56.0'dan
beri bunlardan birini bildiren bir action register etmek `collage.ErrInvalidActionMethod`
ile başarısız olur. Hata, action'ın ve method'un adını verir. Action'ın nasıl
oluşturulduğu fark etmez: `NewAction(...).WithMethods`, bir page'in `WithAction`'ı ya da
`WithActionFor`'u aynı sonucu verir.

`OPTIONS`'a router kendisi cevap verir. Forgery kontrolü de `OPTIONS`'ı `GET` gibi
güvenli bir method sayar. Bu yüzden onu sahiplenen bir action, hiçbir token kontrol
edilmeden state değiştirirdi. CORS preflight'ı bir middleware'in işidir ve
router'dan önce cevaplanır.

Bir form post'u her seferinde aynı kontrollerden, aynı sırayla geçer. Her biri
handler'ınız çalışmadan önce cevap verebilir. Handler çalıştıktan sonra ise
response'u onun döndürdüğü şey belirler:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 404" role="img" aria-labelledby="fsub-title">
<title id="fsub-title">Bir form post'u önce page'in guard'larına, sonra 413 ile cevap veren body limitine, ardından 403 ile cevap veren origin ve token kontrolüne uğrar. Sonra action'ın handler'ı çalışır. Geçersiz bir gönderime page ve 422 ile cevap verilir. Collage-Fetch header'ını gönderen bir script'ten gelen geçerli bir gönderime bir fragment ve 200 ile cevap verilebilir. SeeOther ile cevaplanan geçerli bir gönderim, Location taşıyan bir 303'tür; request Collage-Fetch taşıyorsa Collage-Location taşıyan bir 204'tür.</title>
<defs><marker id="fsub-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">form POST</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="52" width="170" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">page guard'ları</text>
<line class="d-flow" x1="200" y1="65" x2="221" y2="65" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="54" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="68" text-anchor="middle">guard'ın cevabı</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="98" width="170" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">body limiti</text>
<line class="d-flow" x1="200" y1="111" x2="221" y2="111" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="100" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="114" text-anchor="middle">413 · çok büyük</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="144" width="170" height="36" rx="6" /><text x="115" y="160" text-anchor="middle">origin, sonra token</text><text class="d-small" x="115" y="173" text-anchor="middle">yalnızca unsafe method'lar</text>
<line class="d-flow" x1="200" y1="162" x2="221" y2="162" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="151" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="165" text-anchor="middle">403 · ret</text>
<line class="d-flow" x1="115" y1="180" x2="115" y2="196" marker-end="url(#fsub-arrow)" />
<rect class="d-c" x="30" y="200" width="170" height="26" rx="6" /><text class="d-on-c" x="115" y="217" text-anchor="middle">handler'ınız</text>
<path class="d-flow" d="M95 226C95 244 50 240 50 258" marker-end="url(#fsub-arrow)" />
<path class="d-flow" d="M115 226C115 244 150 240 150 258" marker-end="url(#fsub-arrow)" />
<path class="d-flow" d="M135 226C135 244 260 240 260 258" marker-end="url(#fsub-arrow)" />
<rect class="d-a" x="4" y="262" width="92" height="42" rx="5" /><text class="d-on-a" x="50" y="279" text-anchor="middle">geçersiz</text><text class="d-on-a d-small" x="50" y="295" text-anchor="middle">Page · 422</text>
<rect class="d-b" x="100" y="262" width="100" height="42" rx="5" /><text class="d-on-b" x="150" y="279" text-anchor="middle">geçerli, fetch</text><text class="d-on-b d-small" x="150" y="295" text-anchor="middle">RenderFragment</text>
<rect class="d-frame" x="204" y="262" width="112" height="42" rx="5" /><text x="260" y="279" text-anchor="middle">geçerli</text><text class="d-small" x="260" y="295" text-anchor="middle">SeeOther</text>
<line class="d-flow" x1="50" y1="304" x2="50" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="4" y="322" width="92" height="34" rx="6" /><text class="d-small" x="50" y="336" text-anchor="middle">422</text><text class="d-small" x="50" y="349" text-anchor="middle">form + hatalar</text>
<line class="d-flow" x1="150" y1="304" x2="150" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="100" y="322" width="100" height="34" rx="6" /><text class="d-small" x="150" y="336" text-anchor="middle">200</text><text class="d-small" x="150" y="349" text-anchor="middle">yalnızca o kısım</text>
<line class="d-flow" x1="240" y1="304" x2="240" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="204" y="322" width="98" height="34" rx="6" /><text class="d-small" x="253" y="336" text-anchor="middle">Collage-Fetch yok</text><text class="d-small" x="253" y="349" text-anchor="middle">303 · Location</text>
<line class="d-flow" x1="309" y1="304" x2="309" y2="358" marker-end="url(#fsub-arrow)" />
<rect class="d-chip-hit" x="204" y="362" width="112" height="34" rx="6" /><text class="d-hit-text d-small" x="260" y="376" text-anchor="middle">Collage-Fetch: 204</text><text class="d-hit-text d-small" x="260" y="389" text-anchor="middle">Collage-Location</text>
</svg>
<figcaption>Guard'lar, body limiti ve forgery kontrolü handler çalışmadan önce, bu sırayla cevap verebilir; forgery kontrolü token'a bakmadan önce request'in nereden geldiğine bakar. Handler'dan sonra reddedilen bir gönderim 422 ile yine page'dir. Geçerli bir gönderim bir redirect'tir; collage bunu Collage-Fetch taşıyan bir request için Collage-Location taşıyan bir 204'e çevirir. Handler isterse cevap yalnızca değişen fragment da olabilir.</figcaption>
</figure>

## Page üzerinde bir action

Olağan durum, bulunduğu page'e post eden bir form'dur. Page'e o method için bir
action verin:

```go
collage.NewPage("contact").
	WithLayouts(layout).
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

Bir page'deki action'a page'in URL'si üzerinden ulaşılır, bu yüzden page'in
[guard](/docs/pages-and-layouts#private-pages-guards)'ları onu da kapsar: private
bir page'deki bir form, page'in reddettiği bir okuyucu tarafından gönderilemez.
Guard, body okunmadan ve forgery kontrolünden önce çalışır.

Bir action'ın kendi guard'ı da olabilir: `WithGuard`, bir layout'un aldığı
`GuardFunc`'ın aynısını alır (v0.70.0'dan beri). Page üzerindeki bir action'da,
page'in guard'larından sonra sorulur: formu kimin göreceğine page, kimin
gönderebileceğine action karar verir:

```go
WithActionFor(collage.NewAction("delete").
	WithMethods(http.MethodPost).
	WithGuard(guards.Admin). // readers see the page; only admins delete
	WithHandler(deletePost).
	Build())
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

Page'i olmadığı için page guard'ları da yoktur. Onu kimin çağırabileceğini
söylemek için `WithGuard(guards.SignedIn)` ile kendi guard'ını verin
(v0.70.0'dan beri): bir page'inkiler gibi, body okunmadan ve forgery
kontrolünden önce sorulur. `collage add action api/like --path /api/like --guard
signedIn` onu bu şekilde yazar, [`collage routes`](/docs/cli#collage-routes) da
hangi action'ların kendi guard'ı olduğunu gösterir.

Ona post eden bir form path'i yazmak yerine action'ı adıyla belirtir. Böylece
action taşındığında form eski adrese post etmeye devam etmez (v0.36.0'dan beri):

```html
<form method="post" action="{{actionURL "like"}}">{{csrfToken}}…</form>
```

`actionURL` parametreleri `pageURL` gibi alır ve render'ın locale'ini takip eder.
Go tarafında render'ın locale'i için `rc.ActionURL("like", params)`
(v0.37.0'dan beri), başka bir locale için
`app.ActionURL("like", locale, params)` kullanılır. Bir page'in
action'ı register edildiği adla bulunur (yukarıdaki `WithAction` için
`"contact:POST"`), ama aynı URL'ye `pageURL` ile de ulaşılır. Ayrıntılar için
[`actionURL`](/docs/template-functions#actionurl) bölümüne bakın.

Bir action bir page ile aynı path'i paylaşabilir. `WithAction`'ın yaptığı da tam
olarak budur. Ancak page'in cevap verdiği `GET` ve `HEAD` için bu geçerli değildir.
v0.18.0'dan beri bir page'in ya da bir document'ın path'inde bu iki method'dan
birine cevap veren bir action, `collage.ErrDuplicateRoute` ile reddedilir. İkisinin
hangi sırayla register edildiği fark etmez. Önceden böyle bir action önce eşleşir ve
page'i tek kelime söylemeden gizlerdi. İki action da aynı path'te aynı method'a
cevap veremez.

`RegisterAction`; adı, path'i, method'u ya da handler'ı olmayan bir action'ı da
reddeder. Daha önce alınmış bir adla gelen ikinci bir action'ı da reddeder. Her
birinin kendi hatası vardır ve hepsi [Hatalar](/docs/errors#actions) sayfasında
listelenir.

## Action neyle cevap verir

Response'u bir `ActionResult` belirler. Onu bir helper ile oluşturun:

| Helper | Response |
| --- | --- |
| `collage.SeeOther(url)` | `url`'e `303 See Other` |
| `collage.SeeOtherPage(rc, name, params)` | Adıyla belirtilen bir page'e `303 See Other`; path'ini `rc.URL` kurar. `(*ActionResult, error)` döner (v0.68.0'dan beri) |
| `collage.RenderPage(page)` | Bu request'in `RenderContext`'i ile render edilmiş bütün bir page |
| `collage.RenderFragment(f)` | Tek bir fragment'in markup'ı, layout olmadan |
| `collage.JSON(status, v)` | `application/json` olarak marshal edilmiş `v`; `(*ActionResult, error)` döner |
| `collage.JSONBytes(status, body)` | Zaten encode edilmiş `body`, `application/json` olarak |
| `collage.Status(status)` | Sadece bir status, body yok |

v0.60.0'dan önce `collage.JSON` encode edilmiş byte'ları, `JSONOf` ise değeri
alıyordu. Artık `JSON` değeri, `JSONBytes` byte'ları alır; `NoContent` ise `Status`
oldu.

Struct'ı kendiniz de doldurabilirsiniz. Struct'ın body üretmek için dört alanı
vardır: `Location`, `Fragment`, `Page` ve `Body`. Bunlardan birini set edersiniz:
v0.60.0'dan beri ikisini ya da daha fazlasını set eden bir result, birinin sessizce
kazanması yerine `collage.ErrAmbiguousActionResult` olarak bildirilen bir `500`'dür.
Diğer alanlar bunlarla birlikte kullanılır:

- `Status`, her birinin seçeceği status'un yerine geçer.
- `Header` response'a yazılır; `Set-Cookie` için doğru yer burasıdır.
- `ContentType`, `Body` ile birlikte kullanılır. Boş bir `ContentType`,
  `application/octet-stream` anlamına gelir; byte'lara bakılarak asla tahmin
  edilmez.

`nil` bir result `204` demektir.

```go
func like(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	n, err := store.Like(ctx, rc.Request.PostFormValue("post"))
	if err != nil {
		return nil, err
	}
	return collage.JSON(http.StatusOK, likeResponse{Count: n})
}
```

Handler'ın döndürdüğü bir hata `500` olur. Hata `collage.ErrNotFound`'u wrap
ediyorsa `404` olur.

Header'ı kendiniz set etmediğiniz sürece action'ın response'u
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
yenilemek onu ikinci kez yapmamalıdır. Redirect edilen page'de "kaydedildi" demek
için [elagoht/flash](/docs/plugins#elagohtflash), tek seferlik bir mesajı redirect'in
öbür tarafına taşır.

Bir script'in `fetch` ile gönderdiği bir form da redirect'i izlerdi. Redirect'in
gösterdiği page'i indirir, sonra oraya gider ve page ikinci kez render edilirdi.

v0.20.0'dan beri `Collage-Fetch` header'ını (`collage.FetchHeader`) taşıyan bir
request'e redirect yerine `204 No Content` ile cevap verilir ve hedef
`Collage-Location` (`collage.LocationHeader`) içinde gönderilir. Script de
yalnızca bir kez gider. Action eskisi gibi `SeeOther` döndürür. collage-live'ın
form'ları bu header'ı gönderir.

v0.71.0'dan beri bu yolla yalnızca bir path ya da bir `http`/`https` adresi
iletilir. Script `location.assign` ile gider. Bu, bir `Location`'ı izleyen
tarayıcının çalıştırmayı reddettiği bir `javascript:` URL'sini sayfanın içinde
çalıştırır. Bu yüzden `javascript:`, `data:` ya da başka bir hedef, `fetch`'in
izlemediği `303`'ün kendisiyle cevaplanır.

Header yalnızca bir redirect'i değiştirir. Cevap bir page ya da fragment ise header
olsa da olmasa da aynıdır. Bir script'e page'den daha azıyla cevap vermek isteyen
handler, [aşağıdaki iş listesinin](#example-a-team-task-list)
yaptığı gibi header'ı kendisi okur.

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
		WithData(collage.Load(contactData)).
		Build()

	return collage.NewPage("contact").
		WithLayouts(layouts.Master()).
		WithContent(form).
		WithPath("en", "/contact").
		WithAction("POST", sendContact).
		Build()
}

var contactFormKey = collage.NewKey[contactView]("contact:form")

func sendContact(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	email := strings.TrimSpace(rc.Request.PostFormValue("email"))
	if !strings.Contains(email, "@") {
		contactFormKey.Set(rc, contactView{Error: "That does not look like an email address.", Email: email})
		// rc.Page is the page this action answers on: the contact page.
		return &collage.ActionResult{Status: http.StatusUnprocessableEntity, Page: rc.Page}, nil
	}
	if err := messages.Send(ctx, email, rc.Request.PostFormValue("message")); err != nil {
		return nil, err
	}
	return collage.SeeOther("/contact/thanks"), nil
}

func contactData(_ context.Context, rc *collage.RenderContext) (contactView, error) {
	// Set by the action when this render is its answer; empty on an ordinary GET.
	view, _ := contactFormKey.Get(rc)
	return view, nil
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
`collage.ErrUnregisteredPage` ile reddeder.

Bir page'in URL'sindeki action'da `rc.Page`, o page'in register edilmiş değeridir
(v0.33.0'dan beri). Bu yüzden yukarıdaki action'a hiçbir şey verilmesi gerekmez;
bir closure'ın yakalayacağı `var page` de yoktur. Öncesinde bu iş böyle
yapılıyordu.

Kendi URL'sinde register edilmiş bir action'ın orada page'i yoktur ve
`rc.Page`'i `nil`'dir. Böyle bir action, istediği page'in register edilmiş
değeriyle cevap verir; bu page action'dan önce oluşturulur. Dev mode'da body'siz
bir `422` ile cevaplanan form post'u uyarı olarak log'lanır. Çünkü okuyucu boş bir
sayfa görür ve bunun olağan sebebi `nil` bir `rc.Page`'dir.

Bir page'e path yerine adıyla redirect etmek için `collage.SeeOtherPage`
kullanın (v0.68.0'dan beri). Böylece path'i değişen bir page, redirect'i eski path'e
yönelmiş hâlde bırakamaz. Path'i
[`rc.URL`](/docs/links-and-locales#links-from-go) ile, request'in locale'inde ve her
değeri kendi path segment'i içinde tutarak kurar. Ad bilinmiyorsa hata döner.
Bunun için de action'a `*App` taşımak gerekmez:

```go
return collage.SeeOtherPage(rc, "post", map[string]string{"slug": slug})
```

v0.68.0'dan önce path'i `rc.URL` ile (v0.37.0'dan beri) kurun ve `SeeOther`'a verin.

Okuyucuyu geldiği yere geri gönderen bir login, o yeri `next` query parametresinden
okur ve bu parametre okuyucunun girdisidir. `Location` olmadan önce
`collage.SafeRedirect`'ten geçirin (v0.44.0'dan beri):

```go
return &collage.ActionResult{
	Location: collage.SafeRedirect(rc.Request.URL.Query().Get("next"), "/"),
}, nil
```

`next` bu sitede bir path ise onu, değilse fallback'i döndürür. Böylece
`next=https://evil.example` `/`'de biter. Bkz.
[Redirect'ler sitede kalır](/docs/security#redirects-stay-on-the-site).

Bir form'un bunun etrafında genellikle ihtiyaç duyduklarını üç plugin karşılar:

- [elagoht/validate](/docs/plugins#elagohtvalidate) alanları kontrol eder ve
  reddedilen bir gönderime page ve 422 ile cevap verir; okuyucunun girdisi ve her
  alanın mesajı page'de yer alır.
- [elagoht/honeypot](/docs/plugins#elagohthoneypot) spam'i CAPTCHA olmadan
  reddeder.
- [elagoht/session](/docs/plugins#elagohtsession) kimin giriş yaptığını imzalı bir
  cookie'de tutar.

## Forgery koruması

Unsafe bir method ile (`GET`, `HEAD` ve `OPTIONS` dışındaki her şey) bir action'a
gelen her request, handler çalışmadan önce bir forgery token'ı için kontrol edilir.
Geçerli bir token taşımayan request `403` alır ve handler onu hiç görmez. Error
hook'ları nedenini de görür:

| Hata | Neden |
| --- | --- |
| `collage.ErrCSRFMissing` | Token ya da cookie yok |
| `collage.ErrCSRFMismatch` | Token cookie'deki ile eşleşmiyor |
| `collage.ErrCSRFInvalid` | Token'ı bu uygulama imzalamamış |
| `collage.ErrCSRFCrossOrigin` | Tarayıcı, gönderimin başka bir origin'den geldiğini söylüyor (v0.34.0'dan beri) |

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

### Request nereden geldi

İmzalı bir token, saldırgandan saklanan bir secret değildir: form içeren her page,
onu yükleyen herkese bir token verir. Domain'iniz için cookie yazabilen biri (kardeş
bir subdomain, düz HTTP üzerinde bir man-in-the-middle) bir ziyaretçinin
tarayıcısına kendi token'ını yerleştirip onu gönderebilir. Bu yüzden v0.34.0'dan
beri token tek başına kontrol edilmez.

Tarayıcı bir request'in nereden geldiğini `Sec-Fetch-Site`'ta, eski bir
tarayıcıysa `Origin`'de (`Host` ile karşılaştırılarak) söyler. Tarayıcının başka bir origin'den geldiğini işaretlediği bir request, ne
taşırsa taşısın `403` ile reddedilir. Kardeş bir subdomain de başka bir origin
sayılır, çünkü cookie'yi yerleştirebilecek olan tam da odur. İki header'ı da
taşımayan bir request tarayıcıdan gelmemiştir ve token kontrolüne geçer.

Başka bir origin'de olup buraya post etmesi amaçlanan bir form (kendi subdomain'inde
duran bir admin paneli gibi) `Security.CSRFTrustedOrigins`'te belirtilir:

```go
Security: collage.SecurityConfig{
	CSRFKey:            []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	CSRFTrustedOrigins: []string{"https://admin.example.com"},
},
```

Her entry `scheme://host[:port]` biçimindedir. Büyük-küçük harf ve varsayılan port
önemli değildir. Wildcard ise uygulama başlarken reddedilir; her origin'i tek tek
belirtin. Token cookie'si `SameSite=Lax` olarak ayarlanır. Bu yüzden başka bir
*site*'taki güvenilen bir origin onu hiçbir zaman göndermez. Bir origin'e güvenmek,
aynı sitenin başka bir origin'i içindir.

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
kontrol etmez ve key hakkında uyarılmaz.

Üretilen key ilk çalıştırma için sorun değildir, ama deploy için yanlıştır.
Üretilen key her process'te farklıdır. Bu yüzden restart'tan önce yüklenen bir form
restart'tan sonra reddedilir. Bir instance'ın sunduğu form da bir sonraki instance
tarafından reddedilir. Form içeren ve daha önceki bir key ile saklanmış cache'lenmiş
bir page, hiçbir şeyin doğrulayamayacağı bir token ile sunulmaz, yeniden render
edilir. Bkz.
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

`WithStreamingBody()` ile kurulmuş bir action bunun istisnasıdır. Böyle bir action
token'ı yalnızca header'dan okur, bu yüzden form post eden bir `fetch()` bile
`X-CSRF-Token`'ı ayarlar. [Streaming body'ler](#streaming-bodies) bölümüne bakın.

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

Limiti handler'ınız uygulamaz, limit handler'ınız çalışmadan önce uygulanır.
v0.34.0'dan beri [middleware](/docs/middleware-and-apis)'inizden de önce, request'in
yönlendirildiği action'ın limitiyle uygulanır.

- Body'yi okuyan ya da parse eden bir middleware onu sınırlı olarak okur.
- Body'si limitine zaten ulaşmış bir action çalışmadan `413` ile cevap verir.
- `app.Handle` ile mount edilmiş bir handler'a dokunulmaz.

Her handler'ın hatırlaması gereken bir limit, unutan handler'da yoktur.
Anonim bir kullanıcının bulacağı handler da tam olarak o handler'dır. Limitin ötesini
okuyan bir handler `*http.MaxBytesError` alır. Bu hatayı döndürürseniz response `413`
olur:

```go
if err := rc.Request.ParseForm(); err != nil {
	return nil, err // a body over the limit becomes a 413
}
```

v0.57.0'dan beri body'si limite ulaşmış bir handler hangi hatayı döndürürse
döndürsün cevap `413` olur. Bu hata okumanın kendi hatası da olabilir, "upload
failed" gibi handler'ın kendi hatası da. Eskiden ikincisi `500` olurdu. Handler'ın
hatası log'da ve error hook'larına verilen hatada `413`'ün yanında korunur ve
`errors.As` onun üzerinde `*http.MaxBytesError`'ı yine bulur. Bu durum hata olarak
değil, uyarı olarak log'lanır.

Token form'dan kontrol edildiğinde kontrol önce body'yi okur. Bu yüzden limiti aşan
bir form daha orada, handler'ınız çalışmadan önce reddedilir. Bu durumda da cevap
`403` değil `413` olur, çünkü okunamayacak kadar büyük bir body forgery değildir.

Parser'ın memory'de tuttuğu 32 MiB'tan büyük bir multipart body, dosyalarını diske
taşır. v0.34.0'dan beri bu dosyalar, handler çalışmış da olsa request reddedilmiş de
olsa, action cevap verdikten sonra silinir. v0.71.0'dan beri bu, form'u kim parse
etmiş olursa olsun geçerlidir: bir page'in guard'ı, bir action'ın guard'ı, bir data
handler ya da bir middleware. Bir action'ın guard'ları da body'yi zaten sınırlanmış
olarak görür.

v0.71.0'dan beri `Content-Length`'i limitin üstünde olan bir body, forgery kontrolü
ve handler çalışmadan önce `413` ile reddedilir. Böylece forgery kontrolü olmayan bir
action hiçbir zaman yarıda kesilmiş bir form üzerinde çalışmaz. Uzunluğunu
söylemeyen chunked bir body limite takılırsa, başarılı gibi cevaplanmak yerine
handler'dan sonra reddedilir.

Kendi boyut kuralınız (örneğin 5 MB'ı aşan bir fotoğrafı form'da bir mesajla
reddetmek) ancak action'ın limiti bu kuralın üzerindeyse işler. Aksi halde collage
önce `413` ile cevap verir ve mesajınız hiç gösterilmez. Varsayılan limitle 4,5
MB'lık bir fotoğraf, form 5 MB'a izin verdiğini söylese bile sunucu için fazla büyük
diye reddedilir. Böyle bir action'a dosya ve form'un geri kalanı için yer açın:
`WithMaxBodyBytes(maxPhoto + 64<<10)`.

## Streaming body'ler

Büyük bir upload alan action, body'sini bir stream olarak okur ve body'yi ondan önce
hiçbir şey okumamalıdır. v0.57.0'dan beri `WithStreamingBody()` bunu garanti eder:

```go
app.RegisterAction(collage.NewAction("upload").
	WithPath("en", "/upload").
	WithMethods(http.MethodPost).
	WithStreamingBody().
	WithMaxBodyBytes(2 << 30).
	WithHandler(upload).
	Build())
```

- **Body okunmamış olarak gelir.** Handler'dan önce hiçbir şey body'den bir form
  parse etmez. `rc.Request.Body` gönderildiği hâliyle byte'lardır,
  `rc.Request.Form` ve `rc.Request.MultipartForm` ise nil'dir. Body'yi handler
  okur, örneğin `rc.Request.MultipartReader()` ile her seferinde bir part.
  [elagoht/uploads](/docs/plugins#elagohtuploads) plugin'i bunu sizin yerinize
  yapar.
- **Token header'da gelir, yalnızca orada.** Token'ı bir form field'ında bulmak
  body'yi parse etmek demektir, bu yüzden field'a bakılmaz. `X-CSRF-Token` (ya da
  `Security.CSRFHeaderName`) taşımayan bir request `403` ile reddedilir. Error
  hook'larına nedeni, header'ı adıyla belirten `collage.ErrCSRFHeaderRequired` ile
  bildirilir. Okuyucu yalnızca 403'ü görür. Cookie `HttpOnly` olduğu için değeri
  `{{csrfToken}}`'ın render ettiği input'tan okuyun (adı `_csrf` ya da
  `Security.CSRFFieldName`):

  ```js
  const token = form.querySelector('input[name="_csrf"]').value;
  await fetch(form.action, {
    method: "POST",
    headers: {"X-CSRF-Token": token},
    body: new FormData(form),
  });
  ```

- **Düz bir HTML form'u buna post edemez.** Form gönderen bir tarayıcının header
  ayarlamanın bir yolu yoktur. Bu yüzden token'ı kimsenin okumadığı bir field'da
  kalır ve request reddedilir. Yukarıdaki gibi `fetch()` ile gönderin.
  `WithoutCSRF()` kontrolü yine tamamen kapatır ve body yine okunmamış olarak
  gelir.
- **`MaxBodyBytes` yine geçerlidir ve artırılmalıdır.** 4 MiB'lık varsayılan limit
  streaming bir body'yi de diğerleri gibi sınırlar ve handler okurken uygulanır.
  Limitin ötesini okuyup başarısız olan bir handler, hangi hatayı döndürürse
  döndürsün `413` ile cevaplanır.
- **Form'u parse eden her şey stream'i elden verir.** `rc.Request.FormValue`,
  `ParseMultipartForm`, `validate.Form` gibi bir helper: her biri body'yi handler'dan
  önce sonuna kadar okur. Hiçbir şey başarısız olmaz. net/http bütün body'yi,
  `MaxBodyBytes`'a kadar, `os.TempDir()` altındaki geçici dosyalara yazar (birçok
  Linux sunucusunda bu bir `tmpfs`, yani bellektir) ve upload artık stream
  edilmez. Body'yi bundan sonra kendisi okuyan bir handler onu tükenmiş bulur. Bunun
  yerine handler'ın stream'den okuduğu field'ları validate edin. Bir plugin'in
  `BeforeActionEvent.Form()`'u böyle bir action için tek byte okumadan
  `collage.ErrStreamingBody` döndürür.
  [Plugin yazmak](/docs/writing-plugins#beforeactionhook) sayfasına bakın.
- **Yarıda kesilen bir upload client'ın hatasıdır.** Client gittiği ya da bağlantı
  koptuğu için body'nin okunması başarısız olur ve handler ardından bir hata
  döndürürse action `400` ile cevap verir. Okuma deadline'ı geçtiyse cevap `408`
  olur. Bu durum debug seviyesinde log'lanır ve hiçbir error hook'u çağrılmaz, çünkü
  iptal edilen bir upload, bir upload'un bitmesinin en sıradan yoludur. Kural olarak
  bu cevabı kimse almaz. Giden client artık yoktur, `WithBodyTimeout` altında da
  write deadline'ı read deadline'ı ile birlikte geçmiştir. Bu yüzden status
  metrikler ve access log içindir. Bu, streaming olsun olmasın her action için
  geçerlidir.

Register sırasında `POST`, `PUT` ya da `PATCH`'ten hiçbirine cevap vermeyen bir
action'daki `WithStreamingBody()`, `collage.ErrStreamingBodyMethod` ile reddedilir.

### Uzun upload'lar ve timeout'lar

Sunucunun deadline'ları page'lere göre boyutlanmıştır. `ReadTimeout` (varsayılan
15s), body dahil bütün request'in okunmasını sınırlar. `WriteTimeout` (30s) ise
request'in header'ları gelir gelmez başlar, bu yüzden yalnızca response'u değil
upload'u da sınırlar. Bundan yavaş bir upload başarısız olur. `WriteTimeout`'u aşıp
artırılmış bir `ReadTimeout`'u aşmayan bir upload daha da kötüdür: handler dosyayı
kaydeder, client ise cevap alamaz ve yeniden dener.

`WithBodyTimeout(d)` (v0.57.0'dan beri) tek bir action'a kendi deadline'ını verir.
O action'a gelen bir request için iki deadline'ı da şimdi artı `d` ile değiştirir.
Böylece sitenin geri kalanı kısa deadline'larını korur. Bunu page'in guard'larından
sonra ve body'yi herhangi bir şey okumadan önce yapar. Sıradan bir action'da bu,
body'yi okuyan forgery kontrolünden öncedir. Kontrolü yalnızca bir header okuyan
streaming bir action'da ise kontrolden sonradır. Böylece token taşımayan bir request
sunucunun deadline'ları altında reddedilir:

```go
app.RegisterAction(collage.NewAction("upload").
	WithPath("en", "/upload").
	WithMethods(http.MethodPost).
	WithStreamingBody().
	WithMaxBodyBytes(2 << 30).
	WithBodyTimeout(30 * time.Minute).
	WithHandler(upload).
	Build())
```

`d`'yi bütün upload'u ve ardından gelen cevabı kapsayacak şekilde seçin: en az
`MaxBodyBytes` bölü hizmet vermek istediğiniz en yavaş bağlantı. Varsayılan değer
olan sıfır, sunucunun deadline'larını korur. Negatif bir değer register sırasında
`collage.ErrNegativeBodyTimeout` ile reddedilir. Deadline'lar
`http.ResponseController` üzerinden ayarlanır. Bunun için sunucu ile action
arasındaki her `ResponseWriter` wrapper'ının `Unwrap`'i implement etmesi gerekir.
collage'ın kendi wrapper'ları ve `elagoht` plugin'lerininkiler bunu yapar. Biri
yapmıyorsa request sunucunun deadline'ları altında çalışır ve bir kez uyarı
log'lanır.

Bu kontrolleri geçen bir request bağlantısını `d` süresince tutabilir. Form'un
olduğu page'i açan herkes ise token'ı bedavaya alır. Upload'u bir page guard'ı ya da bir auth
middleware'i ile koruyun. Bunlar bütün bu adımlardan önce çalışır. Handler'ın
içindeki bir kontrol ise `d` verildikten sonra çalışır. Öndeki bir proxy'nin
nelere ihtiyaç duyduğu [Deployment](/docs/deployment#uploads) sayfasındadır.

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

Page'ler de, onları oluşturan `collage.Cached` değerleri de stale olarak
işaretlenir. İkisi de tutulur (page'ler v0.63.0'dan beri). v0.64.0'dan beri
invalidate edilen page'ler ardından [warm edilir](/docs/caching#warming-after-an-invalidation)
ve önce redirect'in kendi page'i gelir.

`Location` aynı origin'deyse onun varyantı, redirect yazılmadan önce taze veriyle
yeniden render edilir. Burada varyant, query'sinin cache key'deki hâli ve page'in
her `Vary` header varyantı demektir. Böylece redirect yeni page'e, taze bir
`HIT`'e varır. Form'u gönderenin
varacağı varyant, yani `Vary` değerlerini onun request'inin bildirdiği varyant, önce
gelir ve son warm'ından beri kimse okumamış olsa bile warm edilir.

Bunu yalnızca bu action'ın response'u bekler. Bekleme şu üçünden en kısası kadar
sürer:

- 5 saniye;
- request'in deadline'ı;
- response'un write deadline'ından kalan süre, eksi redirect'i yazmak için yarım
  saniye. Write deadline, action kendi `WithBodyTimeout`'unu ayarlıyorsa odur,
  ayarlamıyorsa request'in geldiği andan itibaren `Server.WriteTimeout`'tur.

Hiç süre kalmadıysa response hiç beklemez. O ana kadar warm edilmeyen key'ler
arka plandaki turda kalır.

Bunu yalnızca `ActionResult.InvalidateTags` yapar. `app.InvalidateTags`'i (ya da
`PurgeTags`'i) handler'ının içinde kendisi çağıran bir action'ın page'leri de, her
invalidation'ınki gibi arka planda warm edilir. Ama redirect'i, page'i beklenmeden
yazılır. Bu yüzden form'u gönderen kişi page'e `STALE` olarak varabilir ya da
page'in render'ının bedelini o öder.

Page yalnızca render zamanında bitemediğinde stale olarak sunulur
(`X-Collage-Cache: STALE`, `Cache-Control: no-store`): upstream yavaş ya da
kapalıdır, bekleme süresi dolmuştur ya da render başarısız olmuştur. Redirect o
zaman da gönderilir ve bir `Cached` değerinin yerini alacak çekme işlemi arka
planda çalışır. Upstream'i kapalıyken de cevap vermeyi sürdüren bir site'ın bedeli
budur.

Bir action ne olursa olsun kendi değişikliğini hemen göstermek zorundaysa,
bunun yerine handler'dan o tag'ler için `app.PurgeTags` çağırın. Bu durumda hiçbir
şey tutulmaz, bu yüzden başarısız bir çekme işlemi page'i de başarısız kılar.

Temiz bir URL'ye ya da query'sinde yalnızca page'in `WithCacheParams` ile adını
verdiği parametreler olan bir URL'ye redirect edin. Bunları tanımlayan bir page,
key'inden diğer her query parametresini çıkarır. Bu yüzden `/posts?saved=1`,
`/posts` olarak warm edilir. Hiçbirini tanımlamayan bir page ise ham query'ye göre
key oluşturur: `/posts?saved=1` hiçbir tarifin tutmadığı ayrı bir key'dir ve onun
için hiçbir şey warm edilmez.

## Kendi URL'sinde bir fragment

Bir client framework'ü olmadan page'in bir kısmını yenilemek için iki şey gerekir:
sadece o kısımla cevap veren bir URL ve o kısmı yerine koyan birkaç satır
JavaScript. Birincisini `WithFragmentPath` sağlar:

```go
collage.NewPage("search").
	WithLayouts(layout).
	WithContent(searchContent).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Dynamic().
	Build()
```

`GET /search/results?q=grid` sadece `results` fragment'ini render eder, başka hiçbir
şeyi render etmez. Fragment'in data handler'ı çalışır, kendi slot'ları doldurulur ve
failure policy'si uygulanır. Bu aynı render'dır, sadece daha aşağıdan başlar.
Etrafında bir layout yoktur ve response'u framework tarafından asla cache'lenmez.

"Yok" ile "bozuk" burada da page'deki gibi birbirinden ayrı kalır. Data handler'ı
`collage.ErrNotFound`'u wrap eden required bir fragment `404` ile cevap verir.
Başka her hatanın cevabı `500`'dür. Fragment path bir action gibi cevaplandığı için
hatası düz metindir, asla cache'lenmez ve içinde page'den hiçbir şey yoktur. Onu
çeken bir script'in okuyabileceği de budur. v0.39.2'den önce fragment path ikisinde
de `500` ile cevap veriyordu. Bu yüzden bir script "görme yetkiniz yok" ile
"bozuk"u birbirinden ayıramıyordu.

```js
const input = document.querySelector('input[name="q"]');
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Tanımlanmamış hiçbir şeye erişilemez. Her fragment'i otomatik olarak dışarı açan bir
framework, her page'in her iç kısmını public web'e açmış olurdu.

Bir fragment path, onu açan page'in bir kısmıdır. v0.34.0'dan beri önce page'in
[guard](/docs/pages-and-layouts#private-pages-guards)'larından geçer: en dıştaki
layout'tan içe doğru, ardından content fragment'inin guard'ı. Sonra fragment'in kendi
guard'ı çalışır. Private bir page'in fragment path'i de private'tır.

v0.34.0'dan önce yalnızca fragment'in kendi guard'ı çalışıyordu. Bu yüzden
içeriği kendi URL'sinde açılmış guard'lı bir layout, o içeriği herkese sunuyordu.
Bir fragment, page'den daha sıkı olması gereken bir kısım için hâlâ kendi
guard'ını taşıyabilir:

```go
results := collage.NewFragment("results", "fragments/results.html").
	WithData(collage.LoadTagged(search)).
	WithGuard(requireAdmin). // after the page's own guards
	Build()
```

Bir fragment path'i de diğer route'lar gibi sahiplenilir. Bir page'in, bir
document'ın ya da başka bir page'in fragment path'iyle veya bir redirect'in
kaynağıyla aynı yazılan bir fragment path'i register sırasında reddedilir. İkisinin
hangi sırayla geldiği fark etmez. Aksi hâlde page'i tek kelime söylemeden gizlerdi.

**Her fragment path'i kendi başına bir render'dır.** Bir page içinde aynı yavaş
değere ihtiyaç duyan fragment'ler onu `collage.Once` ile paylaşır: render başına tek
bir fetch. Kısım kısım yenilenen bir page ise birkaç render'dır ve `Once` bunlar
arasında hiçbir şey paylaşmaz. Aynı veriyi okuyan fragment'ler için
`collage.Cached` kullanın. `Cached` değeri TTL'i boyunca ya da tag'lerinden biri
invalidate edilene kadar render'lar arasında saklar:

```go
var statsKey = collage.NewKey[monitor.Stats]("system:stats")

stats, err := collage.Cached(rc, statsKey, time.Second, nil,
	func(ctx context.Context) (monitor.Stats, error) { return monitor.Collect(ctx) })
```

Bu örnekteki gibi bir ölçüme tag verilmez: onu TTL'i güncel tutar, onu okuyan her
fragment de kendi tag'ini döner. Tag verilirse her fragment bütün tag'lere bağlı
olur ve bir kısmı invalidate etmek hepsini yeniden render eder.

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

### Ona link vermek

Path'i bir kez, `WithFragmentPath` içinde yazın. Ona giden her link'i, `pageURL`'in
page'ler için yaptığı gibi adıyla oluşturun (v0.18.0'dan beri):

```html
<div data-live="{{fragmentURL "search" "results"}}">{{slot "results"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

`fragmentURL` render'ın locale'ini kullanır ve varsayılan locale'e fallback yapar.
`fragmentURLIn "tr" "search" "results"` locale'i kendisi belirtir.
`app.FragmentURL(page, fragment, locale, params)` ise Go tarafındaki karşılığıdır.
Page'in açmadığı bir fragment `collage.ErrUnknownFragmentPath` olur. Geri kalanı
[Link'ler ve locale'ler](/docs/links-and-locales#a-fragments-url) sayfasındadır.

### Hoist ettikleri

Tek başına cevap olarak verilen bir fragment'in etrafında layout yoktur.
Fragment'in kendi yazdığı bir `{{hoist}}` marker'ı page'deki gibi doldurulur. Başka
herhangi bir alana hoist ettiği şeyler ise markup'tan önce gelir (v0.18.0'dan beri).
Örneğin `{{stylesheet}}` ile istenen bir stylesheet ya da bir title böyledir. Her
item için etkisiz bir `<template>` gönderilir:

```html
<template data-collage-hoist="head" data-collage-key="stylesheet:/static/chart.css"><link rel="stylesheet" href="/static/chart.css"></template>
<section>…the fragment…</section>
```

Key, page'in kendi head'inin tekrarları ayıklarken kullandığı key'dir. Böylece bir
script `document.head`'e orada olmayanı ekleyebilir. Bu kanalı yok sayan bir client
template element'leri ekler. Bunlar hiçbir şey render etmez ve hiçbir şey
çalıştırmaz.

### Revalidation

v0.18.0'dan beri bir GET'e verilen cevap bir `ETag` taşır. Bu, body'nin gönderildiği
hâliyle hash'idir. Cevap ayrıca `Cache-Control: private, no-cache` taşır. Eşleşen
bir `If-None-Match` ile gelen request'e body olmadan `304` ile cevap verilir. Render
yine de çalışır. Kazanılan şey, kablodaki body ve client'ın yapacağı iştir. Birkaç
saniyede bir yenilenen bir panel için bu, işin çoğudur. `Cache-Control`'ü kendisi
set eden bir handler kendi değerini korur; ancak bir `PersonaliseHook` body'yi
kişisel hale getirdiyse (v0.43.0'dan beri) handler ne set etmiş olursa olsun
cevap `private, no-cache` olur.

### Tarayıcıdan yenilemek

Framework bir client script'i ile gelmez. [collage-live](/docs/plugins#elagohtlive)
bunu yapan bir plugin'dir. `data-collage-fragment` ile işaretlenmiş element'leri
belli aralıklarla ya da sunucu bir event stream üzerinden bir değişiklik
gönderdiğinde yeniler. Yukarıdaki hoist kanalını ve ETag'i de uygular. Konuştuğu
protokol bu bölümdür. Bu yüzden htmx ya da kendi yazdığınız bir script de aynı
sunucuyla çalışır.

`fetch` ile gönderdiği bir form, [yukarıdaki kuralı](#failure-renders-success-redirects)
kısım kısım izler: cevap form'un hedefine yalnızca başarı durumunda ya da `422`
geldiğinde yerleştirilir. Bu yüzden bir action, bir gönderimin validation'dan
geçmediğini form'un fragment'ini hatalarla birlikte yeniden ve 422 status'uyla
döndürerek söyler (collage-live v0.2.0'dan beri):

```go
var problemsKey = collage.NewKey[[]string]("problems")

if problems := validate(rc); len(problems) > 0 {
	problemsKey.Set(rc, problems)
	result := collage.RenderFragment(formFragment)
	result.Status = http.StatusUnprocessableEntity
	return result, nil
}
```

Başka her hata (bir 500, reddedilen bir forgery token) hedefi olduğu gibi bırakır
ve onu bir hata page'iyle doldurmak yerine stale olarak işaretler. Bir redirect,
tarayıcının izleyeceği gibi ve tek request'te izlenir (collage-live v0.2.1'den
beri). Form [`Collage-Fetch`](#failure-renders-success-redirects) gönderir, collage
da redirect yerine action'ın nereye redirect ettiğiyle cevap verir. Aksi halde
`fetch` redirect'i izler ve page'i indirirdi, ardından sayfaya gidildiğinde page
yeniden çekilirdi.

## Örnek: bir ekibin iş listesi

Çok kullanıcılı bir web uygulaması her ekip için `/teams/{team}/tasks` adresinde
bir iş listesi tutar. Listenin altında bir "görev ekle" form'u vardır. Form üç
şekilde çalışmalıdır:

| Gönderim | Cevap |
| --- | --- |
| Boş ya da fazla uzun bir başlık | `422` ve form yeniden; içinde hata ve yazılan değer |
| Düz bir form post'undan gelen geçerli bir görev | Listeye geri `303`; böylece sayfayı yenilemek hiçbir şeyi iki kez eklemez |
| `Collage-Fetch` gönderen bir script'ten gelen geçerli bir görev | `200` ve yalnızca liste fragment'i; yerinde değiştirilir |

Liste iki yerde kullanılan tek bir fragment'tir: page'in slot'unda ve
`WithFragmentPath` ile kendi URL'sinde. Data handler'ı okuduğu veriyi ekibe göre
tag'ler.
Böylece bir görev eklemek o tag'i invalidate eder ve ekibin her üyesi yeni görevi
bir sonraki yüklemede görür.

```go
// Tag is the dependency tag of one team's task list.
func Tag(team string) string { return "tasks:" + team }

type newTask struct {
	Team  string
	Error string
	Title string
}

var newTaskKey = collage.NewKey[newTask]("tasks:new")

func TasksPage() *collage.Page {
	list := collage.NewFragment("task-list", "fragments/task-list.html").
		WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) ([]store.Task, []string, error) {
			team := rc.Param("team")
			tasks, err := store.Tasks(ctx, team)
			return tasks, []string{Tag(team)}, err
		})).
		Build()
	form := collage.NewFragment("task-form", "fragments/task-form.html").
		WithData(collage.Load(func(_ context.Context, rc *collage.RenderContext) (newTask, error) {
			// Set by the action when this render is its answer; empty on a GET.
			view, _ := newTaskKey.Get(rc)
			view.Team = rc.Param("team")
			return view, nil
		})).
		Build()
	content := collage.NewFragment("tasks", "pages/tasks.html").
		WithSlotFragment("list", list).
		WithSlotFragment("form", form).
		Build()

	return collage.NewPage("tasks").
		WithLayouts(layouts.Master()).
		WithContent(content).
		WithPath("en", "/teams/{team}/tasks").
		WithFragmentPath("en", "/teams/{team}/tasks/list", list).
		Incremental(time.Hour).
		WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			team := rc.Param("team")
			fetched := rc.Request.Header.Get(collage.FetchHeader) != ""

			title := strings.TrimSpace(rc.Request.PostFormValue("title"))
			if title == "" || utf8.RuneCountInString(title) > 200 {
				newTaskKey.Set(rc, newTask{Error: "A title is 1 to 200 characters.", Title: title})
				if fetched {
					// A script swaps the form alone.
					result := collage.RenderFragment(form)
					result.Status = http.StatusUnprocessableEntity
					return result, nil
				}
				// A browser gets the whole page back, its input in the form.
				return &collage.ActionResult{Status: http.StatusUnprocessableEntity, Page: rc.Page}, nil
			}

			if err := store.AddTask(ctx, team, title); err != nil {
				return nil, err
			}

			result := collage.RenderFragment(list)
			if !fetched {
				var err error
				result, err = collage.SeeOtherPage(rc, "tasks", map[string]string{"team": team})
				if err != nil {
					return nil, err
				}
			}
			// The team's cached page renders again, for every member.
			result.InvalidateTags = []string{Tag(team)}
			return result, nil
		}).
		Build()
}
```

Form fragment'i `<form>`'un tamamıdır, bu yüzden bir script onu `422`'nin geri
gönderdiğiyle değiştirebilir. Liste fragment'i de `<ul>`'un tamamıdır:

```html
<form id="task-form" method="post" action="{{pageURL "tasks" "team" .Team}}">
  {{csrfToken}}
  {{with .Error}}<p class="error">{{.}}</p>{{end}}
  <input name="title" value="{{.Title}}">
  <button type="submit">Add</button>
</form>
```

```html
<ul id="task-list">
  {{range .}}<li>{{.Title}}</li>{{else}}<li>No tasks yet.</li>{{end}}
</ul>
```

Birkaç satırlık bir script form'u `Collage-Fetch` ile post eder. Listener
document üzerindedir, bu yüzden bir `422` form'u değiştirdikten sonra da çalışmaya
devam eder:

```js
document.addEventListener("submit", async (event) => {
  const form = event.target;
  if (form.id !== "task-form") return;
  event.preventDefault();
  const response = await fetch(form.action, {
    method: "POST",
    headers: { "Collage-Fetch": "1" },
    body: new FormData(form),
  });
  if (response.status === 200) {
    document.querySelector("#task-list").outerHTML = await response.text();
    form.reset();
  } else if (response.status === 422) {
    form.outerHTML = await response.text();
  }
});
```

Örneğin dayandığı noktalar şunlardır; her biri yukarıda anlatıldı:

- **Forgery token'ı form'la birlikte gider.** `new FormData(form)`,
  `{{csrfToken}}`'ın render ettiği hidden input'u taşır. Bir `422`'den sonra
  cevap olarak verilen fragment de yeni bir token render eder.
- **Page cache'lenir, form yine çalışır.** `Incremental(time.Hour)` page'i içindeki
  token marker'ıyla cache'ler; her okuyucu kendi token'ını alır.
- **Önce invalidation gelir.** Tag, response yazılmadan önce invalidate edilir.
  Böylece redirect yeni listeye varır.
- **Kimin post edebileceğine page karar verir.** Yalnızca ekibin üyelerini içeri
  alan bir [guard](/docs/pages-and-layouts#private-pages-guards) page'i, action'ını
  ve `/teams/{team}/tasks/list`'i aynı şekilde kapsar.

Script kapalıyken aynı form yine post edilir, redirect edilir ve liste yeniden
yüklenir.
[Page'in bir kısmını güncellemek](/docs/update-part-of-a-page) rehberi bu yapıda
bir projeyi adım adım kurar; orada liste kendini de yeniler.
