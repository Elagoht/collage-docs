---
description: collage CLI'ın bütün komutları (new, add, dev, build, export, serve, inspect, check, plugin, version, help), flag'leri ve her birinin neyi çalıştırdığı.
reference: Main, Env, Env.PluginConfig, Env.Static, DispatchCommands, Command, ErrUnknownCommand, FragmentBuilder.WithoutTypeCheck, App, Registrable
---

# collage CLI

`collage` komutu yeni projeleri scaffold eder ve var olan projelerinizi yönetir.
Bir projeyi development modunda çalıştırır, deploy edeceğiniz binary'yi derler ve
projeyi static dosyalar olarak export eder.

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

CLI, uygulamanızı kendi içine asla link etmez. Zaten edemez de, çünkü uygulamanız
sizin kodunuzdur. `dev`, `build`, `export`, `inspect` ve `check` komutları, `go` aracını bulunduğunuz
dizinde tıpkı elle çalıştıracağınız gibi çalıştırır. Bu sayfanın geri kalanı her
komutun tam olarak neyi çalıştırdığını anlatır.

## Kullanım

```sh
collage <command> [flags]
```

| Komut | Ne yapar |
| --- | --- |
| `new` | Yeni bir collage projesi scaffold eder |
| `add` | Bulunduğunuz projeye bir page, fragment, action, document, layout, guard, middleware ya da command yazar (v0.40.0'dan beri) |
| `dev` | Bulunduğunuz dizindeki projeyi development modunda çalıştırır |
| `build` | Bulunduğunuz dizindeki projeyi deploy edeceğiniz binary'ye derler |
| `export` | Bulunduğunuz dizindeki projeyi static dosyalara render eder |
| `serve` | Bir static export'u, bir static host'un sunacağı şekilde sunar |
| `routes` | Bulunduğunuz dizindeki projenin cevap verdiği her URL'yi, her birinin karşılaştığı guard'larla listeler (v0.69.0'dan beri) |
| `inspect` | Bulunduğunuz dizindeki projenin nelerden oluştuğunu JSON olarak yazdırır |
| `check` | Bulunduğunuz dizindeki projenin template'lerini render etmeden kontrol eder (v0.40.0'dan beri) |
| `plugin` | Projeye plugin ekler, çıkarır ya da bir plugin repository'sini kontrol eder (v0.58.0'dan beri) |
| `version` | collage CLI'ın sürümünü yazdırır |
| `help` | Bir komutun yardımını gösterir ya da bütün komutları listeler |

`collage help <command>` o komutun kendi kullanım bilgisini yazdırır.
`collage <command> -h` de aynı şeyi yapar. Flag'ler Go'nun `flag` sözdizimini
kullanır: `-out dist` ile `-out=dist` aynıdır, `--out` da çalışır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 380" role="img" aria-labelledby="cli-life-title">
<title id="cli-life-title">Bir projenin collage komutları boyunca hayatı. collage new, projeyi bu dizinde scaffold eder. Proje üzerinde çalışırken: collage add bir page, fragment, action ya da document yazar; collage dev projeyi localhost:6060'ta çalıştırır ve Go kodu değiştiğinde yeniden build eder; collage check template'lerin link'lerini kontrol eder; collage plugin add ve del, plugins.go'yu ve go.mod'u düzenler; collage inspect projeyi JSON olarak yazdırır. Yayına almak için iki yol vardır. collage build, bin/ içine tek bir static binary derler; -i ile yanına bir image dosyası ve bir servis unit'i de yazar. Binary bir sunucuya gider ve orada HOST ile PORT'u okur. Ya da collage export, HTML ve static dosyaları dist/ içine render eder; collage serve dist/'i localhost:4000'de gösterir, bir static host da aynı dosyaları sunar.</title>
<defs><marker id="cli-life-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="100" y="8" width="120" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">collage new</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#cli-life-arrow)" />
<rect class="d-frame" x="6" y="50" width="308" height="122" rx="8" />
<text x="16" y="67">bu dizindeki proje</text>
<rect class="d-cmd" x="14" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="62" y="92" text-anchor="middle">collage add</text>
<rect class="d-cmd" x="112" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="160" y="92" text-anchor="middle">collage dev</text>
<rect class="d-cmd" x="210" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="258" y="92" text-anchor="middle">collage check</text>
<text class="d-small" x="62" y="114" text-anchor="middle">Go dosyası yazar</text>
<text class="d-small" x="160" y="114" text-anchor="middle">:6060'ta çalışır</text>
<text class="d-small" x="258" y="114" text-anchor="middle">link kontrolü</text>
<rect class="d-cmd" x="14" y="124" width="146" height="24" rx="12" /><text class="d-cmd-text" x="87" y="140" text-anchor="middle">collage plugin add/del</text>
<rect class="d-cmd" x="170" y="124" width="136" height="24" rx="12" /><text class="d-cmd-text" x="238" y="140" text-anchor="middle">collage inspect</text>
<text class="d-small" x="87" y="162" text-anchor="middle">plugins.go ve go.mod</text>
<text class="d-small" x="238" y="162" text-anchor="middle">JSON olarak yazdırır</text>
<line class="d-flow" x1="84" y1="172" x2="84" y2="186" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="236" y1="172" x2="236" y2="186" marker-end="url(#cli-life-arrow)" />
<rect class="d-cmd" x="14" y="190" width="140" height="24" rx="12" /><text class="d-cmd-text" x="84" y="206" text-anchor="middle">collage build</text>
<rect class="d-cmd" x="166" y="190" width="140" height="24" rx="12" /><text class="d-cmd-text" x="236" y="206" text-anchor="middle">collage export</text>
<line class="d-flow" x1="84" y1="214" x2="84" y2="228" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="236" y1="214" x2="236" y2="228" marker-end="url(#cli-life-arrow)" />
<rect class="d-folder" x="14" y="232" width="140" height="34" rx="4" /><text x="84" y="246" text-anchor="middle">bin/&lt;name&gt;</text><text class="d-small" x="84" y="259" text-anchor="middle">tek bir static binary</text>
<rect class="d-folder" x="166" y="232" width="140" height="34" rx="4" /><text x="236" y="246" text-anchor="middle">dist/</text><text class="d-small" x="236" y="259" text-anchor="middle">HTML ve static dosyalar</text>
<line class="d-hitline" x1="66" y1="266" x2="66" y2="284" />
<rect class="d-folder" x="14" y="284" width="104" height="34" rx="4" /><text class="d-small" x="66" y="298" text-anchor="middle">-i: image dosyası</text><text class="d-small" x="66" y="311" text-anchor="middle">ve servis unit'i</text>
<line class="d-hitline" x1="218" y1="266" x2="218" y2="284" />
<rect class="d-cmd" x="166" y="284" width="104" height="24" rx="12" /><text class="d-cmd-text" x="218" y="300" text-anchor="middle">collage serve</text>
<text class="d-small" x="218" y="322" text-anchor="middle">localhost:4000</text>
<line class="d-flow" x1="140" y1="266" x2="140" y2="336" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="292" y1="266" x2="292" y2="336" marker-end="url(#cli-life-arrow)" />
<rect class="d-pill" x="14" y="340" width="140" height="34" rx="17" /><text x="84" y="354" text-anchor="middle">bir sunucu</text><text class="d-small" x="84" y="367" text-anchor="middle">HOST ve PORT'u okur</text>
<rect class="d-pill" x="166" y="340" width="140" height="34" rx="17" /><text x="236" y="354" text-anchor="middle">bir static host</text><text class="d-small" x="236" y="367" text-anchor="middle">aynı dosyaları sunar</text>
</svg>
<figcaption>Bir projenin komutlar boyunca hayatı. <code>collage new</code> projeyi bir kez scaffold eder. Proje üzerinde çalışırken <code>add</code>, <code>dev</code>, <code>check</code>, <code>plugin add</code> ile <code>del</code> ve <code>inspect</code>, bulunduğunuz dizindeki projeye etki eder. Yayına almak için <code>collage build</code> bir sunucu için tek bir binary derler; binary <code>HOST</code> ve <code>PORT</code> üzerinde dinler. <code>-i</code> ise yanına bir image dosyası ve bir servis unit'i yazar. Ya da <code>collage export</code> projeyi <code>dist/</code> içine render eder; <code>collage serve</code>, siz yüklemeden önce <code>dist/</code>'i bir static host'un sunacağı gibi gösterir. <code>collage plugin check</code> şemada yoktur: bir plugin'in kendi repository'sinde çalışır.</figcaption>
</figure>

### Çıkış kodları

| Kod | Anlamı |
| --- | --- |
| `0` | Başarılı. `help`, `collage -h` (v0.11.0'dan beri) ve `collage <command> -h` de bu kodla çıkar |
| `1` | Komut doğru parse edildi ama işini yapamadı |
| `2` | Kullanım hatası: komut verilmedi, komut bilinmiyor, flag hatalı ya da beklenmeyen bir argüman var |

## collage new

```sh
collage new <name> [--template minimal|demo] [--with plugins] [--dir path] [--module path] [--force]
```

`<name>` adında, çalıştırılmaya hazır yeni bir proje scaffold eder.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--template name` | `minimal` | Scaffold edilecek proje: `minimal` ya da `demo` (v0.14.2'den beri; v0.32.0'dan beri varsayılan `minimal`, öncesinde `demo`) |
| `--with plugins` | yok | Proje yazıldıktan sonra eklenecek plugin'ler: `@prod` gibi [paketler](#bundles) ve `elagoht/session` gibi plugin'ler, virgülle ayrılmış ya da birden fazla `--with` ile (v0.59.0'dan beri) |
| `-dir path` | `./<name>` | Projenin scaffold edileceği dizin |
| `-module path` | `<name>` | `go.mod`'a yazılan module path |
| `-force` | kapalı | Dizin boş olmasa da scaffold eder |

```sh
collage new myblog                                   # into ./myblog, module "myblog", one page
collage new myblog --template demo                   # the demos and their tests
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
collage new myblog --with @blog,@prod                # iki paketin plugin'leriyle birlikte
```

`--with`, hiçbir şey yazılmadan önce okunur: var olmayan bir paket ya da
okunamayan bir plugin adı kullanım hatasıdır ve dizin oluşturulmaz. Proje scaffold
edildikten sonra `new`, projeye CLI'ın kendi collage sürümünü ekler, ardından her
plugin'i [`collage plugin add`](#collage-plugin) gibi ekler. Başarısız olan plugin
raporlanır ve dışarıda bırakılır; proje ve diğer plugin'ler kalır, komut `1` ile
çıkar. Eklenen her plugin'in sizden beklediği şeyler (setup notu, template
fonksiyonları ve dokümantasyonu) en sonda, bir arada yazdırılır.

Scaffold edilen `main.go`, `BASE_URL`'ü `Config.BaseURL`'e okur (v0.59.0'dan beri):
sitemap'in, canonical link'in ya da feed'in bir path'in önüne koyduğu origin. Bunu
yazan bir plugin eklenene kadar boş kalabilir.

Flag'leri tek ya da çift tireyle yazabilirsiniz. Flag'ler addan önce de sonra da
gelebilir. Tam olarak bir ad vermeniz gerekir. Hiç ad vermemek ya da birden fazla
ad vermek kullanım hatasıdır. Hedef dizin varsa ve boş değilse, `-force`
vermediğiniz sürece komut reddedilir. `-force` verdiğinizde ise scaffold'un
yazdığı dosyalar, dizindeki aynı adlı dosyaların yerine geçer. v0.69.0'dan beri
komut, yerine geçtiği dosyaları da sayar: aralarındaki `go.mod`, sizin module
path'iniz ve dependency'lerinizdir. Bulunduğunuz dizine scaffold ettiğinizde artık
`cd .` demez.

**Her projede** şunlar bulunur:

- `go.mod`;
- elli satır kadar bir `main.go`: kendisine verilen `collage.Env`'den config'i ve
  static mount'u kuran `newApp` ile onu `collage.Main`'e veren bir `main`
  (v0.68.0'dan beri). `collage.Main`, [aşağıda](#the-contract-with-maingo)
  anlatılan CLI sözleşmesini yerine getirir ve
  [plugin komutlarını](#plugin-commands) dispatch eder;
- `files`'ı tanımlayan `embed.go` ve `embed_dev.go`: template'ler, static dosyalar
  ve `plugins-config.json` embed edilir, development build'inde ise proje dizininin
  kendisi kullanılır (bkz. [aşağısı](#development-builds-embed-nothing));
- boş bir `plugins.go` ile `plugins-config.json` ve bir `.env.example`
  (v0.68.0'dan beri her projede);
- bütün route'ları register eden bir `routes.go`;
- sitenin adını `WithTitle` ile veren ve hiç slot tanımlamayan bir layout
  (böylece bir page'in kendi başlığı sitenin başlığının yerine geçer);
- template'ine proje adını `collage.Value` ile veren bir ana sayfa (içinde hiçbir
  şey veri çekmediği için `Static()` demeden static'tir);
- `static/`, bir `.gitignore` ve bir README.

**`--template demo`**, projeye canlı demolardan oluşan bir page ekler:
JSON dönen bir action, kendi page'ine post eden bir form, kendi URL'si olan bir
fragment ve bir JSON document. Demo projesi ayrıca bir not-found page'i,
[`collagetest`](/docs/testing) ile yazılmış, her biri için testleri ve bir
favicon'u da ekler.

**Yerleşim alana göredir** (v0.40.0'dan beri), yani gerçek bir uygulamanın
yerleşimidir:

```text
pages/<area>/<name>.go            a page: its layouts, content, path, actions
fragments/layouts/master.go       Master(), the layout every page wraps itself in
fragments/pages/<area>/<name>.go  each page's content, mirroring pages/
actions/<area>.go                 action builders, one file per area
actions/funcs/<area>.go           their handlers
documents/<name>.go               routes that are not HTML
data/<domain>/                    state, by domain
templates/                        the HTML kept in files
```

`pages/` ve `fragments/pages/` alan alan birbirini yansıtır. İçlerindeki her
package'ın adı `pages` ya da `fragments`'tır. Bu yüzden bir page dosyası kendi
içeriğini `fragments "<module>/fragments/pages/<area>"` olarak import eder;
`routes.go` da her alanın page'lerini `demopages` gibi bir alias ile import eder.
Küçük fragment'lar markup'larını inline tutar: data handler'larının yanında bir
`collage.InlineHTML` const'u olarak. Layout ve daha büyük page'ler ise markup'larını
`templates/` içinde tutar. Bir page'in `WithActionFor` ile eklediği bir action,
page'le aynı adı taşır ve kendine ait bir path'i yoktur. `routes.go` her şeyi tek
bir [`app.Register`](/docs/pages-and-layouts#registration) çağrısıyla register eder.

**Minimal proje**, v0.32.0'dan beri varsayılan projedir ve bir projenin olabileceği en yalın hâldir. İçinde tek bir
page'i saran layout vardır. Bu page `<h1>Hello from {{.Name}}</h1>` satırından
ibarettir; `{{.Name}}`, template'e `collage.Value` ile verilen proje adıdır.
Bunun yanında arka plan ve metin rengini dark mode dahil ayarlayan bir stylesheet
bulunur. Başka hiçbir şey yoktur: test yoktur, not-found page'i de yoktur. Siz bir
not-found page'i register edene kadar collage bilinmeyen adreslere kendi sade 404'üyle
cevap verir.

İş bittiğinde sonraki adımları yazdırır:

```sh
cd myblog
go mod tidy
collage dev
```

## collage add

```sh
collage add <kind> <[area/]name> [flags]
```

Scaffold'un yerleşimine uygun bir page, fragment, action ya da document yazar ve
onu `routes.go` içinde register eder (v0.40.0'dan beri). v0.69.0'dan beri bir
layout, bir guard, bir middleware ya da bir command da yazar:

```sh
collage add page blog/post          # pages/blog/post.go + fragments/pages/blog/post.go
collage add page blog/post --file   # its template in templates/pages/blog/post.html
collage add fragment blog/sidebar   # fragments/blog/sidebar.go, for a slot
collage add fragment layouts/partials/siteFooter --file
                                    # fragments/layouts/partials/siteFooter.go
                                    # + templates/layouts/partials/siteFooter.html
collage add action blog/comment     # actions/blog.go + actions/funcs/blog.go
collage add action blog/ping --path /api/ping
collage add document feed --path /feed.xml --type application/xml
collage add action blog/like --page blog/post   # attached to pages/blog/post.go
collage add layout panel            # fragments/layouts/panel.go, nested
collage add guard admin             # guards/admin.go + fragments/layouts/admin.go
collage add page admin/users --layout panel --guard admin
collage add middleware loadUser     # middleware/loadUser.go + app.Use in routes.go
collage add command seed            # commands/seed.go + app.RegisterCommand
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--file` | kapalı | Template'i inline değil, template kökü altındaki bir dosyada tutar |
| `--path pattern` | `/<area>/<name>` | URL. Path'i olmayan bir action, eklendiği page'de cevap verir |
| `--name name` | son parça | Page'in, action'ın, document'ın ya da command'ın adı |
| `--locale code` | `Locale.Default` ya da `en` | Path'in ait olduğu locale |
| `--type type` | `text/plain; charset=utf-8` | Bir document'ın content type'ı |
| `--layout name` | yok | Bir page'in layout'u: `add layout <name>`'in yazdığı layout, sitenin layout'unun içinde (v0.69.0'dan beri) |
| `--guard name` | yok | Bir page'in guard'ı: `add guard <name>`'in yazdığı layout, sitenin layout'unun ve varsa `--layout`'un içinde (v0.69.0'dan beri) |
| `--page target` | yok | Bir action'ın page'i, `[area/]name`: action'ı o page'e ekler (v0.69.0'dan beri) |
| `--dir path` | `.` | Proje |

- **Bir page**, `layouts.Master()` ile (ya da `fragments/layouts` içinde argüman
  almayan ilk layout ile) sarılmış bir `pages/<area>/<name>.go` dosyasıdır. İçeriği
  `fragments/pages/<area>/<name>.go` içindedir: bir template, bir view struct'ı ve
  page'in başlığını belirten, tipli bir `collage.Load` handler'ı.
- **Bir action** için `actions/<area>.go` dosyasının sonuna bir builder,
  `actions/funcs/<area>.go` dosyasının sonuna da bir handler eklenir. Bu dosyalar
  yoksa oluşturulur. `--path` verildiğinde action kendine ait bir URL'de cevap verir
  ve register edilir. Verilmediğinde ise eklendiği page'de cevap verir ve komut,
  eklemeniz gereken `.WithActionFor(...)` satırını yazdırır.
- **`--page <[area/]name>`** (v0.69.0'dan beri), yeni action'ı nasıl ekleyeceğinizi
  yazdırmak yerine onu o page'e kendisi ekler: page'in `NewPage` zincirini bitiren
  `Build()`'den önce `.WithActionFor(actions.Like())` satırını koyar. Birden fazla
  zincir içeren bir page dosyasına dokunmaz, eklenecek satırı yazdırır. `--path`
  ile birlikte verilemez.
- **Bir document**, `documents/<name>.go` dosyasıdır.
- **Bir layout** (v0.69.0'dan beri), sitenin kendi layout'unun içine yerleşip
  page'in içeriğini saran bir layout'tur: `fragments/layouts/<name>.go`,
  içinde `<div class="<name>">{{slot "content"}}</div>`. Template'i inline olur,
  `--file` ile bir dosyada durur. `--layout <name>` yeni bir page'i onunla sarar.
  `add layout`'un ya da `add guard`'ın yazdığı layout'lar `//collage:nested` ile
  işaretlenir, guard taşıyan bir layout da iç içe sayılır: bunların hiçbiri, yeni
  bir page'in içine konduğu site layout'u sanılmaz.
- **Bir guard** (v0.69.0'dan beri) iki dosyadır. `guards/<name>.go`,
  [guard](/docs/pages-and-layouts#private-pages-guards)'ın kendisidir: siz kontrolünü
  yazana kadar herkesi `403` ile reddeden bir `GuardFunc`. Bir request'i geçirmek
  için `nil, nil`, başka yere göndermek için bir redirect dönün.
  `fragments/layouts/<name>.go` ise kendisi bir şey çizmeyen, yalnızca guard'ı
  taşıyan bir layout'tur, çünkü guard, page'in layout zincirinde durduğu yerde
  çalışır. `--guard <name>` yeni bir page'i onunla sarar:
  `WithLayouts(layouts.Master(), layouts.Admin())`. Page'in render'ları, ona
  eklenmiş action'lar ve fragment path'leri önce guard'a sorulur. Kendi `--path`'i
  olan bir action'ın page'i, dolayısıyla guard'ı da yoktur; erişimi handler'ında
  kontrol eder.
- **Bir middleware** (v0.69.0'dan beri), her request'i bir sonrakine geçiren bir
  `func(http.Handler) http.Handler` olan `middleware/<name>.go` dosyasıdır.
  [`app.Use`](/docs/middleware-and-apis#middleware-appuse) ile, zaten var olan
  middleware'lerden sonra register edilir; ilk register edilen en dıştadır.
  Routing'den önce çalışır, guard'lar ise ondan sonra. Bu yüzden bir guard,
  middleware'in request context'ine koyduğunu okuyabilir: giriş yapmış kullanıcıyı
  middleware'de yükleyin, kararı guard'da verin.
- **Bir command** (v0.69.0'dan beri), `app.RegisterCommand` ile register edilen bir
  `collage.Command` olan `commands/<name>.go` dosyasıdır: `go run . <name>` ya da
  build edilmiş binary'nin ardından bu ad, sunucu yerine onu çalıştırır. `--name`
  ona Go adı olmayan bir ad verir, örneğin `user-add`. Projede başka bir
  `collage.Command`'ın kullandığı bir ad reddedilir. Bir service'e ihtiyacı olan
  middleware ya da command, onu sizin ekleyeceğiniz bir argüman olarak alır.
  İkisi de register fonksiyonuna kendi satırı olarak eklenir:
  `if err := app.Use(...); err != nil { return err }`. Bunun için fonksiyonun
  `*collage.App` alması ve `error` dönmesi gerekir; değilse eklenecek satır
  yazdırılır.
- **Bir fragment**, bir page'in ya da layout'un bir slot'u için page'i olmayan bir
  page içeriğidir. Komut, onu bağlayan `.WithSlotFragment(...)` satırını yazdırır.
  v0.68.0'dan beri path'inin gösterdiği yere, `fragments/<area>/<name>.go`'ya
  yazılır; template'i de template kökü altında `<area>/<name>`'dedir. Öncesinde
  `fragments/pages/` altına yazılırdı. Package'ı o dizinde zaten olan package'tır,
  yoksa dizinin adını alır: `fragments/layouts/partials`, `partials` olur. Bir
  page'in içeriğinin yanında duracak bir fragment
  `collage add fragment pages/blog/related` ile yazılır.

Son parça bir Go adıdır ve yazıldığı gibi kalır (v0.68.0'dan beri):
`blog/storyDetail`, `layouts/partials/siteFooter`. `--name` başka bir ad
vermedikçe komutun yazdığı her şey, yani dosya, varsayılan URL ve ad, bu parçanın
adını alır. Böylece `blog/storyDetail`, `/blog/storyDetail` adresindeki
`"storyDetail"` page'idir. Fonksiyonu da ilk harfi büyütülmüş aynı addır:
`StoryDetail()`. Tireli bir ad Go adı olmadığı için reddedilir. Son parçadan
önceki dizinler yalnızca bir path'tir ve tire içerebilir: `user-settings/profile`.
Package'ları ve import alias'ları tireleri atar: `usersettings`. Path'in locale'i,
`main.go` onu nasıl yazıyorsa `Locale.Default`'tur; orada bir literal değilse
`en`'dir; ya da `--locale` ile verilendir. Bir template dosyası `Template.Root`
altına, `Extension`'ıyla birlikte yazılır; bunlar da aynı şekilde okunur. Komut hangi
locale'i kullandığını söyler.

Komut, `routes.go`'nun zaten tuttuğu listeye register eder. Bu liste bir
`app.Register(...)` çağrısı olabilir; yeni öğe kendi türündeki son öğeden sonra
gelir. Ya da `RegisterPage` ile üzerinde dolaşılan bir `[]*collage.Page{...}`
literal'i olabilir; `Document` ve `Action` için de aynısı geçerlidir. Komut dosyayı
yerinde düzenler, böylece yorumlar oldukları yerde kalır. Alanın package'ını da
scaffold'un kullandığı türden bir alias ile import eder. Böyle bir liste yoksa orada
hiçbir şeyi değiştirmez ve eklenmesi gereken satırı yazdırır. Constructor'lar
argüman almaz: bir servise ihtiyaç duyan bir page'e o servis elle eklenir.

**Hiçbir şeyin üzerine yazılmaz.** Hiçbir şey yazılmadan önce her şey hesaplanır.
Var olan bir dosya, projenin zaten tanımladığı bir page, action ya da document adı
ya da bir dosyanın gireceği package'da zaten bulunan bir identifier, komutu hiçbir
şey yazmadan durdurur.

## collage dev

```sh
collage dev
```

Bulunduğunuz dizindeki projeyi build eder ve `COLLAGE_DEV=1` ayarlanmış olarak
çalıştırır. Go kodu her değiştiğinde projeyi yeniden build edip yeniden başlatır.
Hiçbir flag ya da argüman almaz. Durdurmak için Ctrl-C'ye basın. Interrupt
sinyali programa da ulaşır ve program production'da nasıl kapanıyorsa öyle
kapanır.

Tarayıcı programla değil, `collage dev` ile konuşur. `collage dev`, programınızın
okuyacağı şekilde `HOST` ve `PORT` üzerinde dinler (v0.32.0'dan beri
varsayılan olarak `localhost:6060`, öncesinde `localhost:3000`). Her request'i programa iletir. Programı ise `HOST` ve `PORT`'u
kendine ait bir loopback adresine ayarlayarak başlatır. Cevap verecek bir
program olmadığında [hataları tarayıcıda gösterebilmesi](#errors-in-the-browser)
bu sayede mümkündür.

v0.34.0'dan beri `collage dev` yalnızca `Host`'u bu makineyi adlandıran request'lere
cevap verir: `localhost` ya da onun altındaki bir ad, bir IP adresi, `example.com`
ya da `*.test` gibi ayrılmış bir ad (v0.56.0'dan beri), başlatıldığı `HOST` veya
`COLLAGE_DEV_HOST` içindeki bir ad. Diğer her şeyi `403` ile reddeder.

Başka bir sitedeki bir page, kendi adının `127.0.0.1`'e çözümlenmesini
sağlayabilir (DNS rebinding) ve böylece `collage dev` ile same-origin olur. Bu
durumda page'lerinizi, stack'leriyle birlikte development hata sayfalarını ve
programın çıktısını okuyabilir. Gönderdiği `Host` ise kendi adıdır ve bunu
değiştiremez.

`collage dev`'e başka bir adla erişmek için onu `HOST` bu ada ayarlanmış olarak
başlatın ya da adı `COLLAGE_DEV_HOST` içine yazın. Bu, virgülle ayrılmış bir
listedir. Örneğin `app,mybox.lan`, bir docker-compose servisini ve bir LAN adını
verir. `collage dev` onu shell'den ya da environment dosyasından okur ve kendi
proxy'sine uygular. Programı da listeye kendi `HOST`'unu ekleyerek başlatır,
çünkü program loopback'te dinler ama tarayıcının `Host`'unu görür. Program da
development modunda aynı kuralı uygular.

v0.56.0'dan beri `collage dev`, kendi proxy'si loopback dışında bir adrese
bağlandığında ayrıca bir Warn log'lar, çünkü development sayfalarına o zaman
başka makinelerden erişilebilir. Bkz.
[Deployment](/docs/deployment#development-mode-stays-on-this-machine).

Programın yazdıkları terminale programın yazdığı gibi ulaşır, tek bir farkla:
programın kendi adresi `collage dev`'in adresiyle değiştirilir. Böylece
`collage: listening` satırı açmanız gereken adresi gösterir (v0.32.0'dan beri).

Renkli bir terminalde program `FORCE_COLOR=1` ile başlatılır ve
[varsayılan logger](/docs/configuration#logger) bunu dikkate alır. Programın çıktısı
`collage dev`'e bir pipe üzerinden ulaşsa da satırları saatini, renkli işaretlerini
ve soluk attribute'larını korur. `collage dev`'in kendi satırları da aynı biçimdedir.
`NO_COLOR` ikisini de kapatır.

```text
16:10:23 • collage dev: serving http://localhost:6060
16:10:24 • collage: listening  addr=localhost:6060
16:10:29 • collage dev: change detected, rebuilding
16:10:29 ✗ collage dev: build failed; the last good build is still serving
```

Scaffold edilen `main.go`, `COLLAGE_DEV=1` ayarlı olduğunda development modunu açar
ve kendisine verilen `HOST` ve `PORT` üzerinde dinler: `collage.Main` ikisini de
`newApp`'e verdiği `Env`'e koyar. Development modu template'leri
ve static dosyaları her request'te diskten okur. Bu yüzden onları düzenlediğinizde
rebuild gerekmez ve rebuild yapılmaz; bunun yerine tarayıcınızdaki page kendini
yeniler. Programınızın diskten kendisi okuduğu içerik de (örneğin Markdown), dizini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtildiğinde page'i
yeniler (v0.10.0'dan beri). Development modunun başka neleri değiştirdiğini
[Template'ler](/docs/templates#reloading-in-development) sayfasında bulabilirsiniz.

### Development build'leri hiçbir şey embed etmez

v0.46.0'dan beri `collage dev`, `go build -tags collage_dev` ile build eder.
Scaffold da `//go:embed` satırını `//go:build !collage_dev` ile başlayan
`embed.go` dosyasında tutar. Yanındaki `embed_dev.go` (`//go:build collage_dev`)
aynı değişkeni tanımlar. Development modu `templates/` ve `static/`
dizinlerini zaten diskten okur, yani bir development build'inin embed edeceği bir
şey yoktur. `collage build`, `collage export` ve düz bir `go build` tag vermez ve
eskisi gibi embed eder.

v0.68.0'dan beri iki dosya tek bir dosya sistemi tanımlar: `files`. `embed.go`'da bu
`//go:embed all:templates all:static plugins-config.json`'dır; `embed_dev.go`'da
ise proje dizininin kendisidir, `os.DirFS(".")`. Böylece `files` verilen bir plugin
(bir kart template'i, bir favicon kaynağı, Markdown yazılar) boş bir dosya sistemi
yerine düzenlemekte olduğunuz kopyayı okur. (v0.68.0'dan önce iki dosya
`templatesFS` ve `staticFS`'i tanımlıyordu ve development build'indekiler boştu.)

Sebebi Go build cache'idir. Dosya embed eden bir paket bu dosyaları derlenmiş
hâlinde taşır. Projenizde herhangi bir yerdeki değişiklik `main`'i yeniden derler.
Bu yüzden dosyaları embed eden her build, `templates/` ve `static/`'in bir kopyasını
daha cache'e yazar: her kaydetmede bir tane, beş gün boyunca saklanır. 30 MB'lık bir
`static/`, cache'i her kaydetmede yaklaşık 30 MB büyüttü; birkaç yüz kaydetme
gigabaytlar eder.

v0.46.0'dan önce scaffold edilmiş bir proje hâlâ `main.go` içinde embed eder ve
`collage dev` başlarken bunu söyler:

```text
16:10:23 ! collage dev: main.go embeds files into every development build, and each build stores another copy of them in the Go build cache; move the //go:embed lines into a file constrained with //go:build !collage_dev (see collage help dev)  patterns=all:static all:templates
```

Düzeltmek için `//go:embed` değişkenlerini `main.go`'dan (ve import'larındaki
`"embed"`'i) çıkarıp kendilerine ait iki dosyaya taşıyın. Yeni bir projeye gelen
dosyalar şunlardır:

```go
// embed.go
//go:build !collage_dev

package main

import "embed"

//go:embed all:templates all:static plugins-config.json
var files embed.FS
```

```go
// embed_dev.go
//go:build collage_dev

package main

import (
	"io/fs"
	"os"
)

var files fs.FS = os.DirFS(".")
```

Hâlâ `templatesFS` ve `staticFS`'i okuyan bir `main.go` bu iki değişkeni tutar.
Her biri kendi `//go:embed` satırıyla tanımlanır, `embed_dev.go`'da ise boş olarak
tanımlanır. `files`'a geçmek için `main.go`, `embed.go` ve `embed_dev.go`'yu yeni bir
projeden kopyalayın.

Cache'te zaten birikmiş olanı `go clean -cache` geri kazandırır. Editörünüzdeki
`gopls` tag'siz build ettiği için `embed_dev.go`'yu açtığınızda dosyanın build
constraint'ler yüzünden dışarıda kaldığını söyler; bu beklenen bir durumdur.

### Rebuild'i ne tetikler

| İzlenir | İzlenmez |
| --- | --- |
| projenin herhangi bir yerindeki `.go` dosyaları | `_test.go` dosyaları |
| proje kökündeki `go.mod` ve `go.sum` | template'ler, static dosyalar ve `DevWatch` dizinleri (rebuild olmadan yeniden yüklenir) |
| aşağıda anlatılan ortam dosyası | gizli dizinler (`.git`, `.cache`, …) |
| | `bin`, `dist`, `node_modules`, `testdata`, `vendor` |

Atlanan dizinler, döngünün kendi kendini beslemesini engeller. Çalışan programın
yazdığı hiçbir şey (cache'i ya da bir export) rebuild tetikleyemez.

Rebuild şöyle ilerler:

- **Polling yapar.** Her 300 ms'de bir dosyaların değişiklik zamanlarını ve
  boyutlarını karşılaştırır. Dosya sistemi bildirimleri için bir kütüphane
  kullanmaz ve her platformda aynı şekilde davranır.
- **Art arda gelen yazmalar tek bir rebuild'dir.** Bir değişiklikten sonra,
  dosyalar bir polling aralığı boyunca değişmeyene kadar bekler. Böylece birkaç
  dosyayı yeniden yazan bir formatter yalnızca bir build'e yol açar.
- **Önce yeni build yapılır.** `go build`, binary'yi proje dışındaki geçici bir
  dizine yazar. Eski process ancak yeni binary derlendikten sonra durdurulur ve
  yenisi başlatılır. Eski process, Ctrl-C'de olduğu gibi elindeki request'leri
  bitirir. 10 saniye içinde çıkmazsa öldürülür.
- **Derlenmeyen bir değişiklikte son sağlam build çalışmaya devam eder.**
  Derleyicinin hatası ekranda görünür.
- **Kendiliğinden çıkan bir program** döngü içinde yeniden başlatılmaz. Örneğin
  açılışta panic olabilir ya da bir page'in template'i eksik olabilir. Program bir
  sonraki değişiklikte yeniden başlatılır.

### Tarayıcıdaki hatalar

v0.15.0'dan beri çalışmayan bir program, reddedilen bir bağlantı değil, bir
page'dir:

- **Program çıktığında** (register sırasında bulunamayan bir template, açılışta bir
  panic) **ya da ilk build başarısız olduğunda** her page, programın ya da
  derleyicinin yazdıklarını gösteren bir 503'tür. Açık olan bir page yenilenerek
  bu page'e geçer. Bir değişiklik programı geri getirdiğinde de yeniden yenilenir.
- **Program başlarken yapılan bir request onu bekler.** Program henüz dinlemediği
  için başarısız olmaz.
- **Kendisine söylenen yerde hiç dinlemeyen bir program**, yani `HOST` ve `PORT`'u
  yok sayan bir `main.go`, 10 saniye sonra, kendisine verilen adresle birlikte
  page'de belirtilir.

Program çalışırken başarısız olan bir render'da ise programın kendi development error
page'i görünür. Bu page hatanın sebebiyle başlar. Bkz.
[Hatalar](/docs/errors#rendering).

### Ortam dosyaları

`collage dev`, bulunduğunuz dizindeki `.env.development` dosyasının
değişkenlerini programın ortamına ekler. `.env.development` yoksa `.env`
dosyasını kullanır. Her zaman tek bir dosya okunur, asla ikisi birden okunmaz:
`.env.development`, `.env`'in üzerine merge edilmez, onun yerine geçer.

```sh
# .env.development
PORT=6060
HOST=localhost
export COLLAGE_CSRF_KEY="0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
```

- Dosya `KEY=value` satırlarından, boş satırlardan ve `#` ile başlayan
  satırlardan oluşur. Satırın başında `export ` olabilir. Bir değer, eşleşen tek ya
  da çift tırnak içine de alınabilir. Tırnaklar kaldırılır ve içlerindeki hiçbir
  şey yorumlanmaz.
- Tırnaksız bir değer, boşluktan sonra gelen ilk `#` işaretinde biter. Yani
  `PORT=6060 # dev` satırının değeri `6060`'tır.
- Key'ler harf, rakam ve alt çizgiden oluşur ve rakamla başlayamaz.
- **Hatalı bir satır programın başlamasını engeller.** `collage dev` hatayı dosya
  adı ve satır numarasıyla (`.env.development:3`) bildirir. Satır düzeltilene
  kadar hiçbir şeyi başlatmaz ya da yeniden başlatmaz. Zaten çalışan bir program,
  başlatıldığı değerlerle çalışmaya devam eder. `collage dev` izlemeyi sürdürür;
  düzeltilmiş dosyayı kaydettiğinizde kaldığı yerden devam eder. Hatalı satırı
  atlamak, yazdığınız bir ayarın programa hiç ulaşmaması demek olurdu.
- **Shell'de zaten ayarlı olan bir değişken dosyadakine göre önceliklidir.** Bu
  yüzden `PORT=4000 collage dev` yine çalışır.
- Dosyada ne yazarsa yazsın `COLLAGE_DEV=1` her zaman ayarlanır. Programın
  dinleyeceği `HOST` ve `PORT` da her zaman ayarlanır. Dosyadaki `HOST` ve `PORT`,
  `collage dev`'in kendisinin dinlediği yerdir ve `collage dev` başlarken bir kez
  okunur. `COLLAGE_DEV_HOST` da dosyadan ya da shell'den okunur ve program onu
  `collage dev`'in kendi `HOST`'u eklenmiş olarak alır.
- `BASE_URL`, shell de dosya da onu ayarlamadığında `collage dev`'in kendi adresidir:
  `http://localhost:6060` (v0.68.0'dan beri). Mutlak URL yazan bir plugin (sitemap,
  canonical link) origin olmadan başlamayı reddeder ve development'ta doğru origin
  budur. Deploy edildiğinde onu kendiniz ayarlayın.
- Dosyanın olmaması hata değildir. Bir dosya okunduğunda adı stderr'e yazdırılır.
- Dosya her yeniden başlatmada tekrar okunur. Dosya izlendiği için onu
  düzenlediğinizde program yeni değerlerle yeniden başlar.

`main` fonksiyonu `collage.Main`'i çağıran bir program da bu dosyaları, scaffold
edilen program gibi, aynı kurallarla okur (v0.68.0'dan beri). `go run .`,
`collage export` ve derlenmiş binary, `newApp` çalışmadan önce çalışma dizinindeki
`.env`'i yükler; `COLLAGE_DEV=1` altında önce `.env.development`'a bakılır. Böylece
`os.Getenv("BASE_URL")` onu `source .env` olmadan görür. Development dışında okunan
dosyanın adı stderr'e yazdırılır. Değişkenlerini kendisi ayarlayan bir sunucuda
değişkenler önceliklidir ve dosyanın olmaması hata değildir. Hatalı bir satır ise
programı dosya ve satırıyla birlikte durdurur. `collage.Main`'i çağırmayan kendi
yazdığınız bir `main` hiçbir dosya okumaz.

## collage build

```sh
collage build [-o path] [-os name] [-arch name] [-i]
```

Bulunduğunuz dizindeki projeyi deploy edeceğiniz binary'ye derler. Çalıştırdığı
komut şudur:

```sh
CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build -trimpath -ldflags="-s -w" -o bin/<name> .
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<name>` | Binary'nin yazılacağı yer |
| `-os name` | bu makineninki | Hedef işletim sistemi (`GOOS`) |
| `-arch name` | bu makineninki | Hedef mimari (`GOARCH`) |
| `-i` | kapalı | Binary'nin yanına hangi ek dosyaların yazılacağını sorar |

`<name>`, `go.mod`'daki module path'in son parçasıdır. Dizinde `go.mod` yoksa ya da
`go.mod` bir module tanımlamıyorsa komut hata verir. `-os windows` verildiğinde,
path zaten `.exe` ile bitmiyorsa sonuna `.exe` eklenir. Positional argümanlar
kullanım hatasıdır.

Bu ayarların nedenleri:

- **CGO kapalıdır**, çünkü ne collage ne de standart kütüphane C'ye ihtiyaç duyar.
  Static bir binary, içinde başka hiçbir şey olmayan bir image'a konabilir.
- **`-trimpath`** kullanılır, böylece binary onu derleyen makinedeki path'leri
  içermez.
- **`-s -w`** debug tablolarını çıkarır. Boyutun büyük kısmı bu tablolardır.
- **Varsayılan hedef üzerinde çalıştığınız makinedir** (v0.32.0'dan beri; öncesinde
  linux/amd64). Böylece binary derlendiği yerde çalışır. Sunucu çoğu zaman başka bir
  platformdur: Mac'te derlenen bir binary Linux container'ında çalışmaz. O yüzden
  sunucu için hedefi belirtin: `collage build -os linux -arch amd64`.
- **Binary `dist/` yerine `bin/` dizinine yazılır.** `dist/`, `collage export`'un
  yazdığı yerdir ve `export -clean` onu boşaltır.

İş bittiğinde binary'nin path'ini, platformunu, boyutunu ve build süresini
yazdırır.

### -i ile ek dosyalar

`-i` verildiğinde build'den önce iki soru sorar ve cevapları standart input'tan
okur. `y` ya da `yes` evet anlamına gelir, diğer her şey hayır sayılır:

- **Dockerfile yazılsın mı?** İki aşamalı bir image'dır. İlk aşama bir `golang`
  build aşamasıdır ve CLI'ı derleyen Go'nun major ve minor sürümüne sabitlenir.
  İkinci aşama `gcr.io/distroless/static-debian12` tabanlıdır ve yalnızca binary'yi
  içerir. Bu aşamada `HOST=0.0.0.0` ve `PORT=8080` ayarlanır, 8080 portu expose
  edilir ve `COLLAGE_CSRF_KEY` yorum satırı olarak bulunur.
- **systemd unit yazılsın mı?** `/usr/local/bin/<name>` binary'sini `/srv/<name>`
  dizininden çalıştıran bir `<name>.service` dosyasıdır. İçinde `HOST=127.0.0.1`,
  `PORT=8080`, `Restart=on-failure` ve `TimeoutStopSec=30` vardır. Kurmadan önce
  kendi ihtiyacınıza göre düzenlemeniz gerekir.

İki dosya da binary'nin yanına yazılır: `bin/Dockerfile` ve `bin/<name>.service`.
Bunun nedeni, bu dosyaların üretilmiş olmasıdır; proje kökü ise insanların
yazdığı dosyalar içindir. Bu yerleşimin tek bedeli farklı bir komuttur ve rapor o
komutu yazdırır:

```sh
docker build -f bin/Dockerfile .
```

**Var olan bir dosyanın üzerine asla yazılmaz.** Dosya zaten varsa ilgili soru
sorulmaz. Bunlar projelerin düzenlediği dosyalardır. Bu dosyaları varsayılan
hâliyle değiştiren bir build komutu, birinin emeğini sessizce silmiş olurdu.

`-i` olmadan yalnızca binary yazılır. Ayrıntılar için
[Deployment](/docs/deployment) sayfasına bakın.

## collage export

```sh
collage export [-out dir] [-clean]
```

Bulunduğunuz dizindeki projeyi bir static host için static dosyalara render eder.
HTML olabilen her page için HTML üretir, mount edilmiş bütün asset'leri de çıktıya
ekler. Çalıştırdığı komut şudur:

```sh
go run . -collage-build -out <dir>          # plus -clean when you passed it
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-out dir` | `dist` | Projenin render edildiği dizin |
| `-clean` | kapalı | Build'den önce dizinin mevcut içeriğini siler |

Programın çıktısı, yani build raporu, yeniden biçimlendirilmeden olduğu gibi
aktarılır. Positional argümanlar kullanım hatasıdır.

Strateji tanımlamayan bir page, render ettiği hiçbir şeyin data handler'ı yoksa
export edilir. Handler'ı olan bir page ise `Static()` ya da `Incremental(ttl)`
belirtiyorsa export edilir. Path'inde `{param}` olan bir page, `WithStaticParams`'ın
listelediği her değer için bir kez export edilir; `WithStaticParams` yoksa atlanır.
Neyin export edildiği, neyin hangi nedenle atlandığı ve çıktı dizini için yapılan
güvenlik kontrolleri [Static export](/docs/static-export) sayfasında anlatılır.
Form'ları ya da her request'te üretilen page'leri olan bir site için ihtiyacınız
olan komut `collage build`'dir.

## collage serve

```sh
collage serve [-dir dir] [-host name] [-port n]
```

Bir static export'u, bir static host'un sunacağı şekilde sunar. Böylece
burada gördüğünüz, deploy ettikten sonra göreceğinizle aynıdır. Yalnızca dosya
sunar, projenizi çalıştırmaz.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-dir dir` | `dist` | Sunulacak dizin |
| `-host name` | `localhost` | Dinlenecek interface |
| `-port n` | `4000` | Dinlenecek port |

Port 6060 değil 4000'dir, böylece `collage dev` ile yan yana çalışabilir. İkisini
karşılaştırmak istediğiniz an da tam olarak budur.

Sıradan bir dosya sunucusu gibi değil, bir static host gibi davranır:

- Uzantısı olmayan bir path'e `<path>/index.html` ile cevap verilir. Export'un
  yazdığı yapı da budur.
- Dizinler asla listelenmez.
- Hiçbir dosyaya karşılık gelmeyen bir path'e, export'un kendi `404.html`
  dosyası ve 404 status koduyla cevap verilir. Bu dosya yoksa sade bir 404 döner.
- Yalnızca `GET` ve `HEAD` request'lerine cevap verilir. Diğer her şey 405 alır.
- Bir dotfile (`.env`, `.git/config`) hiçbir zaman sunulmaz (v0.34.0'dan beri).
  Uzantısı bir tip belirtmeyen bir dosyanın tipi, mount'ta olduğu gibi içeriğinden
  belirlenir, ama hiçbir zaman HTML ya da XML olmaz. Her dosya `nosniff` ile gider
  (v0.34.2'den beri).
- Her response `Cache-Control: no-store` header'ıyla gönderilir. Böylece yeniden
  export edip sayfayı yenilediğinizde eski çıktıyı değil yenisini görürsünüz.

Dizin yoksa ya da içinde hiç dosya yoksa komut hata verir ve önce
`collage export` çalıştırmanızı söyler. Path varsa ama dizin değilse de hata
verir ve tam olarak bunu söyler.

## collage routes

```sh
collage routes
```

Projenin collage üzerinden cevap verdiği her URL'yi, method ve path başına bir
satır olarak, path'e göre sıralı listeler (v0.69.0'dan beri):

```text
METHOD  PATH          KIND    NAME    NOTES
GET     /             page    home    static
GET     /admin/users  page    users   dynamic, guard adminGuard
POST    /admin/users  action  rename  on page users, guard adminGuard
POST    /api/ping     action  ping    own URL, no guard
GET     /static/*     mount           12 files

Not listed: handlers mounted with app.Handle, and plugins' own endpoints.
```

Bir page, nasıl cache'lendiğini ve hangi guard'lara sorulduğunu söyler. Bir
action, bir page üzerinde mi (o zaman page'in guard'larına sorulur), yoksa
guard'ı olmayan kendi URL'sinde mi cevap verdiğini söyler. Bir page'in fragment
URL'leri de karşılaştıkları guard'larla listelenir. Sitenin birden fazla locale'i
varsa bir `LOCALE` sütunu çıkar. Komut, [`collage inspect`](#collage-inspect)'in
yazdırdığını okur; yani proje build edilir ve `newApp` çalışır. Collage üzerinden
register edilmeyenler listede yoktur: `app.Handle` ile mount edilen handler'lar ve
`/healthz` gibi plugin'lerin kendi endpoint'leri. Flag almaz.

## collage inspect

```sh
collage inspect
```

Bulunduğunuz dizindeki projenin nelerden oluştuğunu JSON olarak yazdırır
(v0.27.0'dan beri). Çıktıda pattern'leri, parametreleri, `layouts`'u (en dıştaki
başta) ve `guards`'ıyla her page, template'i ve slot'larıyla her fragment (inline
bir fragment `"inline": true` ile ve template path'i olmadan gelir), document'lar,
action'lar, template fonksiyonları, plugin'ler, locale'ler ve mount'ların sunduğu
dosyalar bulunur.
Hiçbir flag almaz; positional argümanlar kullanım hatasıdır. Çalıştırdığı komut
şudur:

```sh
go run . collage-inspect
```

`collage-inspect`, `DispatchCommands`'ın bir plugin'e iletmeden kendisinin cevapladığı
bir komuttur: uygulamanın tarifini girintili JSON olarak yazdırır. Sözleşme bu
JSON'dur; arkasındaki Go tipleri (`App.Inspect`, `collage.Inspection`,
`collage.InspectCommand`) v0.60.0'da public API'den çıktı ve çıktı byte byte
öncekiyle aynıdır.

Adı prefix'lidir, böylece hiçbir plugin'in komut adını elinden almaz. Scaffold
edilen `main.go`, flag'lerinden sonraki kelimeyi `DispatchCommands`'a verir (bkz.
[Plugin komutları](#plugin-commands)). Bu yüzden collage v0.27.0 veya daha yeni bir
sürümle çalışan, scaffold edilmiş bir projenin başka hiçbir şeye ihtiyacı yoktur.
Argümanlarını bu şekilde dispatch etmeyen bir programın ise cevap verecek bir şeyi
yoktur.

Bir editörün completion'ı da bunu okur. VS Code için Collage Snippets & Highlighter
extension'ı ([Editör desteği](/docs/installation#editor-support)),
`{{pageURL "…"}}` içindeki page adlarını, `{{slot "…"}}` içindeki slot'ları ve
`{{asset "…"}}` içindeki dosyaları buradan önerir.

v0.49.0'dan beri her fragment, template'inin `.` olarak gördüğü Go tipini de
taşır. Root da bu tiplerin ulaştığı tiplerin bir tablosu olan `types`'ı taşır.
Böylece bir editör `{{.`'yı tamamlayabilir ve bir alan adını register
sırasında olduğu gibi kontrol edebilir (bkz.
[Template'ler nasıl kontrol edilir](/docs/data-handlers#how-templates-are-checked)):

```json
"fragments": [
  {"name": "post-body", "template": "pages/post.html", "handler": true, "dataType": "*blog.Post"},
  {"name": "layout", "template": "layouts/master.html", "dataType": "nil"},
  {"name": "legacy", "template": "pages/legacy.html", "handler": true, "dataType": null, "typeCheck": false}
],
"types": {
  "blog.Post": {
    "kind": "struct",
    "fields": [
      {"name": "Base", "type": "blog.Base", "embedded": true},
      {"name": "ID", "type": "int"},
      {"name": "Title", "type": "string"},
      {"name": "Author", "type": "*blog.User"},
      {"name": "Comments", "type": "blog.Comments"},
      {"name": "Related", "type": "blog.Tags"}
    ],
    "methods": [{"name": "URL", "args": 0, "returns": "string"}]
  },
  "blog.Base": {"kind": "struct", "fields": [{"name": "ID", "type": "int"}]},
  "blog.Comments": {"kind": "slice", "elem": "blog.Comment"},
  "blog.Tags": {"kind": "map", "key": "string", "elem": "[]blog.Post"},
  "blog.User": {"kind": "struct", "fields": [{"name": "Name", "type": "string"}]}
}
```

- **`dataType`** her zaman yazılır: Go tipi; verisi olmayan ya da
  `collage.Effect` kullanan bir fragment için `"nil"`; tip bir interface olduğu
  ve bu yüzden page render edilene kadar bilinmediği durumda (`any` dönecek
  şekilde tanımlanmış bir handler) `null`.
- **`typeCheck: false`** yalnızca `WithoutTypeCheck()` ile oluşturulmuş bir
  fragment'te görünür.
- **`types`**, herhangi bir fragment'in veri tipinden alanlar, eleman ve key
  tipleri ve method sonuçları üzerinden ulaşılabilen her isimli tipi tutar. Her
  tip Go adıyla key'lenir ve export edilmiş alanlarını (promote edilenler dahil)
  ve kendisinin ve pointer'ının bir şey dönen export edilmiş method'larını
  içerir.
  - Export edilmiş bir embedded alan da bir alandır: template ona `{{.Base}}`
    diye ulaşır. Bu alan `"embedded": true` taşır, promote ettiği alanlar da her
    zamanki gibi yanında listelenir (v0.51.1'den beri).
  - İsimli bir pointer, slice, array, map ya da chan, tuttuğu tipi `elem` ile
    verir. Map ayrıca `key` de taşır (v0.51.1'den beri).
  - Bir tip bir kez listelenir, diğer her yerde adıyla anılır. Böylece
    özyinelemeli bir tip sonsuza gitmez. İsimsiz bir struct, Go yazımıyla
    listelenir. Diğer isimsiz bileşik tipler bir tip string'inin içinde yazılır
    (`[]blog.Comment`).
  - `time.Time` ve `template.HTML` gibi standart kütüphane tipleri
    ayrıntılandırılmaz, yalnızca adıyla anılır.
  - Hiçbir fragment'in bilinen bir veri tipi yoksa `types` yazılmaz.
  - İki farklı tipin paylaştığı bir key (`models` adlı iki paketteki
    `models.User`), alan ve method içermeden
    `{"kind": "struct", "ambiguous": true}` olarak yazılır ve editör onu
    bilinmeyen bir tip sayar (v0.51.2'den beri).

Bunlar eklemedir: çıktının `version`'ı hâlâ `1`'dir.

## collage check

```sh
collage check [-json]
```

Hiçbir şeyi render etmeden her template'in link'lerini kontrol eder (v0.40.0'dan
beri). Adla kurulan bir link (`{{pageURL "post" "slug" .Slug}}`,
`{{pageURLIn "en" "about"}}`, `{{actionURL "logout"}}`, `{{fragmentURL "home" "clock"}}`,
`{{localeURL "en"}}`) template render edildiğinde ve yalnızca ona ulaşan page'de
hata verir. `check` hepsini tek seferde bulur:

```text
$ collage check
error [unknown-route] inline template of fragment "hello":4:16: {{pageURL "featurs"}}: collage: no page or document by that name: "featurs"; did you mean "features"?
```

| Kural | Anlamı |
| --- | --- |
| `unknown-route` | Bu adda bir page, document, action ya da fragment path'i yoktur; register edilmiş en yakın ad önerilir |
| `route-params` | Parametreler route'un pattern'ini doldurmaz: biri eksiktir, biri için placeholder yoktur ya da ad-değer çiftleri hâlinde değildir |
| `unreachable-locale` | Hiçbir URL'nin taşıyamayacağı bir locale: ne `Locale.Default`'tur ne de `Locale.Supported` içindedir |
| `no-path-in-locale` | Route'un, `pageURLIn` ya da `fragmentURLIn`'in belirttiği locale'de bir path'i yoktur |

Kontrolü framework'ün kendi URL builder'larıyla yapar. Bu yüzden bildirdiği şey,
tam olarak bir render'ın hata vereceği şeydir. Yalnızca string literal olarak
yazılmış adlar kontrol edilir: bir alandan gelen ad (`{{pageURL .Name}}`) ancak
template render edildiğinde bilinir. Adı literal olmayan bir parametre de
parametrelerin kontrol edilmemesine yol açar. Kendine ait bir locale'i olmayan bir
link, route'u herhangi bir locale'de kurulabiliyorsa geçer, çünkü bir render
varsayılan locale'e geri döner.

Bir şey bulduğunda `1` ile çıkar. Bu yüzden CI'da `collage build`'den önce yer
alabilir. `-json`, bulguları bir editör için `{level, rule, message}` dizisi olarak
yazdırır. Çalıştırdığı komut şudur:

```sh
go run . collage-check
```

`collage.DispatchCommands` bu komuta uygulamayı başlatıp `App.Check`'i çağırarak
cevap verir. Yani açılışta bir veritabanı açan bir program onu burada da açar. Bunun
için projenin collage'ının v0.40.0 ya da daha yeni olması gerekir; daha eski bir
sürüm `unknown command: "collage-check"` cevabını verir. Bir test de aynı fonksiyonu
çağırabilir:

```go
if findings := app.Check(); len(findings) > 0 {
	t.Errorf("broken links: %v", findings)
}
```

## collage plugin

*v0.58.0'dan beri.* Projeye plugin ekler, plugin çıkarır ya da bir plugin
repository'sini kontrol eder:

```sh
collage plugin add <plugin>... [-dir dir]
collage plugin del <plugin>... [-dir dir]
collage plugin check [-release] [-run] [dir]
```

```sh
collage plugin add elagoht/consent            # github.com/elagoht/collage-consent
collage plugin add elagoht/sitemap@v0.3.1     # go get'in kabul ettiği bir sürüm
collage plugin add gitlab.com/kisi/plugin     # bir module path, yazıldığı gibi kullanılır
collage plugin del elagoht/consent
collage plugin add elagoht/session elagoht/flash   # birden fazla, sırayla (v0.59.0'dan beri)
collage plugin add @prod                      # bir paket
collage plugin del @live elagoht/session
collage plugin check -release -run            # bir plugin'in kendi repository'sinde
```

`owner/name`, `github.com/<owner>/collage-<name>` demektir. İlk parçasında nokta
olan bir path ise module path'tir ve yazıldığı gibi kullanılır. `add` ve `del`,
bulunduğunuz dizindeki projede ya da `-dir`'in gösterdiği projede çalışır. Bu dizinde
`go.mod` ve bir `package main` bulunmalıdır.

**`add`** modülü `go get` ile indirir. `@version` verilmezse `@upgrade` ister: `go.mod`'da
henüz olmayan bir modül için en son release'i, zaten olan bir modül için ise onu hiçbir
zaman geriye almaz. Sonra modülün bir plugin olup olmadığını kontrol eder. Sözleşme
[plugin yazarlarının sayfasındadır](/docs/plugin-authors):

- modülün kökünde, `name`'i `owner/name` olan bir `collage.json`. Kısa ad verdiyseniz
  `name`, yazdığınız adla aynı olmalıdır;
- plugin'in package'ı modülün kökünde olmalı, `package main` olmamalı;
- hiçbir şey ayarlanmadan çağrılabilen bir `New`: parametresiz, variadic ya da sıfır
  değeri yazılabilen tek parametreli. Bu, package'ın kendi export edilmiş bir struct'ı
  için `pkg.Options{}`, pointer, interface, slice, map ya da func için `nil` demektir;
- bir release tag'i; pseudo-version olmaz.

Önce manifest'in `requires` dediği plugin'leri ekler. Sonra `plugins.go`'ya import'u
ve çağrıyı yazar: manifest'inin `order`'ı daha büyük olan ilk girdinin önüne.
`plugins-config.json`'a da boş bir bölüm ekler. En son `go mod tidy` çalıştırır:

```text
$ collage plugin add elagoht/consent
added elagoht/consent v0.2.0
  plugins.go           consent.New(consent.Options{})
  plugins-config.json  "elagoht/consent": {}

elagoht/consent
  The banner needs its categories and wording: set `"categories"` (at least one) and `"text"` (the default locale's wording, labelling every category; other locales under `"locales"`) in this plugin's section of `plugins-config.json`. Without categories the application does not start.
  Docs: https://github.com/Elagoht/collage-consent
```

Ardından plugin'in sizden hâlâ beklediklerini manifest'inden yazdırır
(v0.68.0'dan beri): `setup` notunu, eklediği template fonksiyonlarını (her birini
açıklamasının ilk cümlesiyle) ve dokümantasyonunun adresini. Birden fazla plugin
eklendiğinde (bir paket, `new --with`) bunlar en sonda, "What each plugin needs
from you" başlığı altında bir arada gelir. Boş bölüm varsayılanlar demektir.
Plugin'in zorunlu tuttuğu bir ayar eksikse bunu uygulama başlarken plugin'in
kendisi bildirir.

- **Import, package'ın adını taşır.** Bu ad `plugins.go`'nun bir import'unda ya da
  package'da herhangi bir yerde tanımlı bir adla çakışırsa `<owner><Name>` olur,
  gerekirse sonuna sayı eklenir.
- **`plugins.go`'da zaten olan bir plugin ikinci kez eklenmez.** Yalnızca sürümü
  değişir ve komut bunu söyler.
- **Var olan bölüm korunur.** İki dosya da yerinde düzenlenir, bu yüzden argümanlarınız,
  yorumlarınız, sıralamanız ve biçimlendirmeniz olduğu gibi kalır.
- **Yarım iş kalmaz.** Ctrl-C dahil her hata, `go.mod`, `go.sum`, `plugins.go` ve
  `plugins-config.json` dosyalarını eski hâline döndürür. Önceden olmayan bir dosya
  yeniden silinir.

`plugins.go`'dan önce oluşturulmuş bir proje, bu dosyayı ilk `add`'inde alır. Komut da
`collage.New`'a bir kez verilmesi gereken `Plugins: plugins(),` satırını yazdırır.
Yeni projelerde ikisi de hazır gelir.

**`del`**, plugin'i `plugins.go`'nun girdileri arasında, manifest'lerindeki ada ya da
module path'e göre, büyük küçük harfe bakmadan bulur. Girdilerini, import'unu ve
bölümünü kaldırır, bölümde ayar varsa onu yazdırır ve `go mod tidy` çalıştırır.
Şu durumlarda reddeder:

- başka bir plugin'in manifest'i onu `requires` ediyorsa;
- başka bir girdinin `collage.json`'u okunamıyorsa (modül indirilemiyor ya da dosya
  bozuk). O plugin'in bunu isteyip istemediğini bilemez. Mesaj modülü ve nedeni belirtir;
- ona bir module path verilmişse ve plugin'in kendi `collage.json`'u okunamıyorsa,
  çünkü bölümünün adı bu durumda bilinmez.

Başka bir girdinin `collage.json`'undan yalnızca `name`, `requires` ve `order` okunur.
Bu yüzden daha yeni bir manifest şemasından gelen bir anahtar hiçbir şeyi engellemez.
Elle bağlanmış olup bir plugin modülünün kök package'ı olmayan girdi (kendi
modülünüzün bir package'ı ya da bir bağımlılığın alt package'ı) ve `collage.json`'u
hiç olmayan modül hiçbir şey istemez, hiçbir şeyi engellemez. Module cache'te olmayan
bir modül (yeni bir clone'dan sonra olduğu gibi) manifest'i okunmak için indirilir.
`del` de `add` gibi, bir adım başarısız olursa her dosyayı geri koyar.

### Birden fazla plugin

*v0.59.0'dan beri.* `add` ve `del` birden fazla plugin alır ve onları verilen
sırayla işler. Her biri ayrı bir adımdır: başarısız olan raporlanır ve geri alınır,
öncesindekiler ve sonrasındakiler yerinde kalır. Komut bu durumda `1` ile çıkar.
Okunamayan bir ad, iki farklı sürümle verilen bir plugin gibi, hiçbir şey
değişmeden komutu durdurur. İki kez verilen bir plugin bir kez işlenir.

Birden fazla plugin verildiğinde sonda bir özet gelir; `add`, setup notlarını her
plugin'in ardından değil, en sonda bir arada yazdırır:

```text
$ collage plugin add elagoht/feed elagoht/nope elagoht/toc
added elagoht/feed v0.3.0
  ...
collage: plugin add: elagoht/nope: ...

added: elagoht/feed, elagoht/toc
failed: elagoht/nope

To finish setting up:

elagoht/feed
  Feeds are Go values: give them to `feed.New(...)` in `plugins.go`, ...
```

Sırayı `del` sizin yerinize düşünür: listedeki başka bir plugin'in ihtiyaç duyduğu
plugin, o plugin'den sonra silinir. Bu yüzden `del elagoht/live elagoht/websocket`
çalışır.

### Paketler

*v0.59.0'dan beri.* Paket, tek bir ad altında toplanmış plugin'lerdir ve `@` ile
yazılır: `collage plugin add @prod`, `collage new site --with @blog`. Plugin'leri
aşağıdaki sırayla eklenir ve her biri `plugins.go`'da manifest'inin istediği yere
gider. `del @ad`, paketin projede olan plugin'lerini siler, olmayanları atlar.
`collage help plugin` da paketleri listeler.

| Paket | Ne için | Plugin'ler |
| --- | --- | --- |
| `@prod` | canlıdaki bir sitenin istedikleri | health, secure, compress, minimizer, accesslog, ratelimit |
| `@site` | herkese açık bir sitenin arama ve paylaşım temelleri | robots, sitemap, meta, jsonld, favicon |
| `@blog` | `@site` ve Markdown ile yazılan yazılar | `@site`'ın plugin'leri, markdown, feed, toc, highlight, ogimage |
| `@forms` | formlar ve cookie'de tutulan durum | session, flash, validate, honeypot |
| `@live` | sunucunun güncel tuttuğu fragment'lar | live, websocket |
| `@ops` | metrikler ve hata raporları | prometheus, errortrack |

Her plugin `elagoht/<ad>`'dır. Bazıları uygulama başlamadan önce bir ayar ister:
sitemap ve meta sitenin origin'ini (scaffold edilmiş projede `BASE_URL`), favicon bir
kaynak görseli, markdown ve ogimage `plugins.go`'da verilen bir dosya sistemini,
errortrack development dışında bir DSN'i. Ne yapılacağını setup notları söyler;
yapılana kadar uygulama, eksik ayarın adını vererek başlamayı reddeder.

**`check`**, bir plugin repository'sini (`dir`, varsayılan `.`) `add`'in uyguladığı
sözleşmeye göre kontrol eder. Kontrolün kendisi yalnızca repository'yi okur:

- `-release` ayrıca `HEAD`'de bir release tag'i ister, bunun için git gerekir;
- `-run`, repository'de `go run` ile geçici bir program derler. Module cache'te
  olmayanı indirir. Bu program `New`'u `add`'in yazacağı biçimde, bir `collage.Plugin`
  olarak çağırır ve `Name()`'in manifest'teki `name` olduğunu kontrol eder. Build
  hatasında, `New`'daki bir panic'te ya da farklı bir adda başarısız olur. Plugin'i
  configure etmez, başlatmaz; yani boş bir bölümde eksik kalan bir ayar onun işi
  değildir. `go.mod`, `go.sum` ve `go.work.sum` eski hâline döner, program silinir.

Bunlar `collage` binary'sinin komutlarıdır. Bir plugin'in kendi alt komutları ise
başka bir şeydir: bkz. [Plugin komutları](#plugin-commands).

## collage version

```sh
collage version
```

`collage version <version>` yazdırır. Sürüm, binary'nin build bilgisinden okunur.
Bu yüzden `go install ...@v0.9.0` ile kurulan binary `0.9.0` bildirir. Bir git
checkout'unda derlenen binary ise bir pseudo-version bildirir. Örneğin v0.11.0
tag'inden sonraki bir commit için bu `0.11.1-0.<timestamp>-<commit>` olur.
Yalnızca hiç sürüm bilgisi olmayan bir binary `devel` bildirir. Bu, bir
repository dışında ya da `-buildvcs=false` ile derlenmiş bir binary'dir.

## collage help

```sh
collage help            # every command
collage help export     # one command's usage
```

Argümansız çalıştırıldığında bütün komutları listeler ve `0` ile çıkar. Bir komut
adı verildiğinde o komutun kullanım bilgisini yazdırır. Bilinmeyen bir ad
verildiğinde `collage: unknown command: "nope"` yazdırır ve `2` ile çıkar.

## Örnek: yeni bir projeden ilk deploy'a

Bu bölüm, yukarıdaki komutları bir sitenin onlarla karşılaştığı sırayla izler.
Proje, bir laptop'ta yazılan ve başka platformdaki bir sunucuya deploy edilen
`myblog` adlı bir sitedir.
Çıktıdaki saatler, boyutlar, hash'ler ve cache dizinleri sizin makinenizde
farklı olur.

**1. Scaffold edin.**

```sh
collage new myblog
cd myblog
go mod tidy
```

`new`, [minimal projeyi](#collage-new) yazar ve sonraki adımları yazdırır:

```text
Scaffolded "myblog" in myblog

Next steps:
  cd myblog
  go mod tidy
  collage dev
```

Scaffold edilen `go.mod` yalnızca module adını ve Go sürümünü içerir;
collage'ı `go mod tidy` ekler.

**2. Çalıştırın.**

```sh
collage dev
```

```text
01:13:17 • collage dev: serving http://localhost:6060
01:13:18 • collage: dev mode is rendering templates from disk, not the embedded copy  root=templates
01:13:18 • collage: dev mode, using an in-memory cache instead of the configured disk cache
01:13:18 • collage: listening  addr=localhost:6060
```

`http://localhost:6060` adresinde `Hello from myblog` görünür. Çalışır hâlde
bırakın: sonraki adım Go dosyaları ekler ve `collage dev`, dosyalar oluştuğunda
projeyi yeniden build eder.

**3. Bir page ekleyin ve link'leri kontrol edin.** İkinci bir terminalde:

```sh
collage add page about
```

```text
wrote  fragments/pages/about.go
wrote  pages/about.go
edited routes.go
locale en, from the default; main.go sets no Locale.Default literal, --locale changes it
page "about" at /about
```

`collage dev` yeniden build eder ve `/about` cevap verir. Ardından:

```sh
collage check
```

```text
01:14:05 • collage: caching to disk  dir=.cache/6ba57471adaf48bf  version=19be6c1c06fea3dcff022e9dab873f1f
collage: check: nothing found
```

İlk satır programın kendisine aittir, çünkü `check` programı başlatır. İkincisi
kontrolün sonucudur ve çıkış kodu `0`'dır; CI'da bu, `collage build`'den önceki adımdır.

**4. Sunucu için build edin.** Laptop sunucu değildir, bu yüzden sunucunun
platformunu belirtin ve iki ek dosyayı `-i` ile isteyin:

```sh
collage build -i -os linux -arch amd64
```

```text
Building myblog for linux/amd64.

  Write a Dockerfile? [y/N]: y
  Write a systemd unit? [y/N]: y


✓ bin/myblog
    linux/amd64 · 10.3 MB
    wrote bin/Dockerfile    docker build -f bin/Dockerfile .
    wrote bin/myblog.service

10.3 MB · linux/amd64 · 6.7s
```

Scaffold'un `.gitignore`'u `bin/myblog`'u adıyla yok sayar, yanındaki iki dosyayı
ise yok saymaz: onlar sizin düzenleyip commit edeceğiniz dosyalardır. Komutu
yeniden çalıştırdığınızda iki soru da sorulmaz, çünkü dosyalar zaten vardır.

**5. Sunucuya koyun.** İki dosya, projeyi çalıştırmanın iki yoludur ve yalnızca
biri `bin/myblog`'u kullanır:

- **Servis unit'i**, derlediğiniz binary'yi çalıştırır. `bin/myblog`'u sunucuda
  `/usr/local/bin/myblog` konumuna kopyalayın. Unit'in `User` ve
  `WorkingDirectory` değerlerini düzenleyin (dosya bunları bilemez) ve unit'i
  ilk yorum satırında belirtilen yere kurun. Önüne konacak bir reverse proxy
  için `127.0.0.1:8080` üzerinde dinler.
- **Image dosyası** kendi binary'sini derler: ilk aşaması projeyi kaynaktan aynı
  flag'lerle derler, ikinci aşaması da sonucu taşır. Raporun söylediği gibi proje
  kökünden build edin. `0.0.0.0:8080` üzerinde dinler.

`8080` bu dosyaların kendi ayarıdır, `PORT=8080` olarak yazılır. `HOST` ya da
`PORT` verilmeden çalıştırılan binary, scaffold'un varsayılanı olan
`localhost:6060` üzerinde dinler; bkz. [Deployment](/docs/deployment#environment).
İki dosyada da yorum satırı olarak bir `COLLAGE_CSRF_KEY` bulunur: onu en az 32
rastgele byte'a ayarlayın. Aksi hâlde her process için bir key üretilir ve bir
restart'tan önce gönderilen her form, restart'tan sonra reddedilir.

Plugin'leri `-i`'den önce ekleyin. Image dosyası `plugins-config.json`'ı
yalnızca dosya yazıldığı sırada projede varsa kopyalar ve `-i` dosyayı hiçbir
zaman yeniden yazmaz. Sonradan eklenen bir plugin için o `COPY` satırını elle
eklemeniz gerekir.

**Ya da export edin.** Dosya sunan bir host için `collage export -clean`, bunun
yerine `dist/`'i render eder ve raporu yeni page'i dışarıda bırakır:

```text
✓ 3 files written
    dist/index.html
    dist/static/app.css
    dist/static/app.f106e88ebb47f51d.css

▲ 1 skipped
    about  page uses the dynamic render strategy, which cannot be built statically

3 written · 1 skipped · 0 failed · 1.5ms
```

`collage add`, `about`'a bir `collage.Load` handler'ı verdi ama strateji vermedi,
bu yüzden page dynamic'tir. `pages/about.go` içindeki builder'ına `.Static()`
ekleyin; bir sonraki export `dist/about/index.html`'i de yazar. Ardından
`collage serve`, `dist/`'i `localhost:4000`'de bir static host'un sunacağı gibi
gösterir.

## main.go ile sözleşme

`dev`, `export` ve `inspect`, `main.go`'nuzun yaptığı üç şeye dayanır. Scaffold
edilen `main.go`'nun çağırdığı `collage.Main` (v0.68.0'dan beri) üçünü de yapar.
Üçüncüsü, flag'lerden sonraki kelimeleri `collage.DispatchCommands`'a vermektir.
Plugin'lerinizin komutlarını çalıştıran da budur:

| Komut | Çalıştırdığı | `main.go`'nuzun yapması gereken |
| --- | --- | --- |
| `collage dev` | önce `go build -tags collage_dev`, sonra binary; `COLLAGE_DEV=1`, dinlenecek `HOST` ile `PORT` ve `COLLAGE_DEV_HOST` ayarlı olarak | `COLLAGE_DEV` `1` olduğunda development modunu açmak ve `HOST` ile `PORT` üzerinde dinlemek |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | `-collage-build`, `-out` ve `-clean` flag'lerini parse etmek; `-collage-build` verildiğinde sunmak yerine `<dir>` dizinine render etmek |
| `collage inspect` | `go run . collage-inspect` | flag'lerden sonraki kelimeleri `collage.DispatchCommands`'a vermek |

`collage build`'in `main.go`'dan hiçbir beklentisi yoktur. Derleme, `go build`'in
kendisine hiçbir şey söylenmeden yaptığı bir iştir.

Scaffold edilen `main.go` bunların hepsini `collage.Main`'e bırakır:

```go
func main() {
	collage.Main(newApp)
}

func newApp(env collage.Env) (*collage.App, error) {
	pluginConfig, err := env.PluginConfig(files)
	if err != nil {
		return nil, err
	}

	app, err := collage.New(&collage.Config{
		DevMode: env.Dev,
		BaseURL: os.Getenv("BASE_URL"),
		Server:  collage.ServerConfig{Host: env.Host, Port: env.Port},
		Template: collage.TemplateConfig{
			FS:        files,
			Root:      "templates",
			Extension: ".html",
		},
		Cache: collage.CacheConfig{
			Enabled:    true,
			Type:       "disk",
			Dir:        env.CacheDir,
			DefaultTTL: 5 * time.Minute,
		},
		PluginConfig: pluginConfig,
		Plugins:      plugins(),
		Security: collage.SecurityConfig{
			CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
		},
	})
	if err != nil {
		return nil, err
	}

	if err := register(app); err != nil {
		return nil, err
	}

	return app, app.Mount("/static/", env.Static(files, "static"))
}
```

`collage.Main` [ortam dosyasını](#environment-files) yükler; `-collage-build`,
`-out`, `-clean` ve `-port` flag'lerini parse eder ve `newApp`'i bir `collage.Env`
ile çağırır:

| Alan | Değer |
| --- | --- |
| `Dev` | `collage dev`'in ayarladığı `COLLAGE_DEV=1` |
| `Host` | `HOST` ya da `localhost` |
| `Port` | `-port`, `PORT` ya da `6060` |
| `CacheDir` | `.cache`; bir test kendine ait bir dizin verir |

`env.PluginConfig(files)`, `plugins-config.json` çalışma dizininde varsa onu okur;
böylece bir deployment bir ayarı rebuild olmadan değiştirebilir. Yoksa binary'nin
embed ettiği kopyayı okur; böylece dosyası olmadan deploy edilen bir binary
ayarlarını korur. `env.Static(files, "static")`, static dizinini development'ta
diskten, diğer durumlarda embed edilmiş hâlinden sunar. Ardından `collage.Main` bir
plugin'in komutunu dispatch eder, `-collage-build` verildiğinde `-out`'a render
eder ya da hizmet verir. Bir build'den ya da bir komuttan sonra uygulamayı shutdown
eder.

v0.68.0'dan önce scaffold edilmiş bir proje olduğu gibi çalışmaya devam eder.
Geçmek için yeni `main.go`, `embed.go` ve `embed_dev.go`'yu `collage new`'dan
kopyalayın ve testlerde `newApp(collage.Env{CacheDir: t.TempDir()})` çağırın.
`collage.Main` olmadan yazılmış bir `main` sözleşmeyi şöyle yerine getirir;
v0.68.0'dan önce scaffold edilmiş bir projenin `main`'i bu biçimdedir (burada
yalnızca önemli kısmı bırakılmıştır):

```go
func main() {
	buildFlag := flag.Bool("collage-build", false, "render the app to static files instead of serving it")
	outFlag := flag.String("out", "dist", "output directory for -collage-build")
	cleanFlag := flag.Bool("clean", false, "remove -out's existing contents before building")
	portFlag := flag.Int("port", envInt("PORT", 6060), "port to listen on (env PORT)")
	flag.Parse()

	devMode := os.Getenv("COLLAGE_DEV") == "1"

	app, err := newApp(devMode, *portFlag)
	if err != nil {
		log.Fatal(err)
	}

	// A word after the flags is a plugin's command: go run . <command>
	if args := flag.Args(); len(args) > 0 {
		code, err := collage.DispatchCommands(context.Background(), app, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := shutdown(app); err != nil {
			fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
			code = max(code, 1)
		}
		os.Exit(code)
	}

	if *buildFlag {
		buildErr := staticBuild(app, *outFlag, *cleanFlag)
		if err := shutdown(app); err != nil {
			log.Printf("shutdown: %v", err)
			buildErr = errors.Join(buildErr, err)
		}
		if buildErr != nil {
			log.Fatalf("static build: %v", buildErr)
		}
		return
	}

	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// shutdown runs every plugin's Shutdown once a build or a command is done.
// Neither serves, so nothing else would. Only after a successful start: a start
// that failed has already shut down the plugins it started.
func shutdown(app *collage.App) error {
	if app.Start() != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return app.Shutdown(ctx)
}
```

Bir build ve bir komut hizmet vermez, bu yüzden `ListenAndServe` onlar için
uygulamayı hiçbir zaman shutdown etmez. `collage.Main`, ikisinden biri bittiğinde
`app.Shutdown`'ı kendisi çağırır; kendi yazdığınız bir `main` bunu yukarıdaki gibi
yapar. Böylece her plugin'in `Shutdown`'ı çalışır: bir plugin'in elinde tuttuğu
şey flush edilir ve [elagoht/jobs](/docs/plugins#elagohtjobs) gibi bir queue
plugin'ine verilen iş, drop edildiğinde en azından loglanır.

`main`'i `collage.Main` olmadan yazarsanız bunların hepsinin çalışmaya devam
ettiğinden emin olun. Aksi hâlde `collage dev`, `collage export`, `collage inspect` ve
plugin'lerinizin komutları projenizde işe yarar hiçbir şey yapmaz.

## Plugin komutları

Plugin eklemek ve çıkarmak [`collage plugin`](#collage-plugin) işidir. Bu bölüm ise plugin'in kendi alt komutlarını anlatır. Bir plugin, `Host.RegisterCommand` ile bir komut ekleyebilir. **`collage` CLI bu
komutu çalıştırmaz.** CLI uygulamanızı hiçbir zaman yüklemez. Bu yüzden ancak
plugin'leriniz başladıktan sonra var olan bir komuta CLI erişemez ve
`collage <plugin-command>` bilinmeyen bir komut olarak kalır.

Bu komutları kendi programınız `collage.DispatchCommands` ile dispatch eder.
Scaffold edilen `main.go`'nun çağırdığı `collage.Main` bunu zaten yapar. Kendi
yazdığınız bir `main` bunu şöyle yapar ([yukarıdaki](#the-contract-with-maingo)
blok): flag'ler parse edilip uygulama kurulduktan sonra geriye kalan bir kelime
komut olarak yorumlanır:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := shutdown(app); err != nil { // see above
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
		code = max(code, 1)
	}
	os.Exit(code) // runs no defers, so the shutdown comes first
}
```

`DispatchCommands` uygulamayı başlatır. Bunu yaparken her plugin'in `Init`'ini
çalıştırır; plugin'ler komutlarını `Init` içinde register eder. Ardından ilk argümanın
adını taşıyan komutu çalıştırır. Böylece bir plugin'in komutu, programınızın bir
alt komutu olarak ve varsa flag'lerden sonra çalışır:

```sh
go run . pages
./bin/myblog -port 4000 pages
```

Çıkış kodları CLI'ınkilerle aynı mantığı izler:

| Kod | Ne zaman |
| --- | --- |
| `0` | başarı |
| `1` | açılış hatası, çalışıp başarısız olan bir komut ya da `Run`'ı olmayan bir komut |
| `2` | nil bir app, argüman verilmemesi ya da hiçbir plugin'in register etmediği bir kelime (`ErrUnknownCommand`) |

Hiçbir komuta karşılık gelmeyen bir kelime, yanlışlıkla başlatılmış bir sunucuya
değil, kullanım hatasına yol açar.
Hiçbir plugin'in register etmediği bir kelimeye yine de cevap verilir:
[`collage inspect`](#collage-inspect) komutunun çalıştırdığı `collage-inspect`.
Hiçbir komut eşleşmediğinde sunucuyu başlatmayı tercih eden bir program,
`errors.Is(err, collage.ErrUnknownCommand)` ile kontrol edip çıkmak yerine devam
edebilir. Kendi kullanım metninizi yazdırmak isterseniz `app.Commands()`
plugin'lerin register ettiği komutları listeler.

Dispatch işlemi uygulamayı başlatır ve bu da register aşamasını kapatır. Sonradan
çağrılan `ListenAndServe` aynı başlatmayı yeniden kullanır. `DispatchCommands`
uygulamayı shutdown etmez, çünkü bir program ardından onunla hizmet vermeye devam
edebilir. Bir komut çalıştıktan sonra, yukarıdaki gibi `app.Shutdown`'ı kendiniz
çağırın. Böylece her plugin'in `Shutdown`'ı çalışır. `collage.Main` de bunu
yapar. Böyle bir komutun nasıl yazıldığı
[Plugin yazmak](/docs/writing-plugins#commands) sayfasında anlatılır.
