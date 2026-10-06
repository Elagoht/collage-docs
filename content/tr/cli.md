---
description: collage CLI'ın bütün komutları (new, add, dev, build, export, serve, inspect, check, version ve help), flag'leri ve her birinin tam olarak neyi çalıştırdığı.
reference: DispatchCommands, Command, ErrUnknownCommand, InspectCommand, Inspection, FragmentBuilder.WithoutTypeCheck, App, Registrable
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
| `add` | Bulunduğunuz projeye bir page, fragment, action ya da document yazar (v0.40.0'dan beri) |
| `dev` | Bulunduğunuz dizindeki projeyi development modunda çalıştırır |
| `build` | Bulunduğunuz dizindeki projeyi deploy edeceğiniz binary'ye derler |
| `export` | Bulunduğunuz dizindeki projeyi static dosyalara render eder |
| `serve` | Bir static export'u, bir static host'un sunacağı şekilde sunar |
| `inspect` | Bulunduğunuz dizindeki projenin nelerden oluştuğunu JSON olarak yazdırır |
| `check` | Bulunduğunuz dizindeki projenin template'lerini render etmeden kontrol eder (v0.40.0'dan beri) |
| `version` | collage CLI'ın sürümünü yazdırır |
| `help` | Bir komutun yardımını gösterir ya da bütün komutları listeler |

`collage help <command>` o komutun kendi kullanım bilgisini yazdırır.
`collage <command> -h` de aynı şeyi yapar. Flag'ler Go'nun `flag` sözdizimini
kullanır: `-out dist` ile `-out=dist` aynıdır, `--out` da çalışır.

### Çıkış kodları

| Kod | Anlamı |
| --- | --- |
| `0` | Başarılı. `help`, `collage -h` (v0.11.0'dan beri) ve `collage <command> -h` de bu kodla çıkar |
| `1` | Komut doğru parse edildi ama işini yapamadı |
| `2` | Kullanım hatası: komut verilmedi, komut bilinmiyor, flag hatalı ya da beklenmeyen bir argüman var |

## collage new

```sh
collage new <name> [--template minimal|demo] [--dir path] [--module path] [--force]
```

`<name>` adında, çalıştırılmaya hazır yeni bir proje scaffold eder.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--template name` | `minimal` | Scaffold edilecek proje: `minimal` ya da `demo` (v0.14.2'den beri; v0.32.0'dan beri varsayılan `minimal`, öncesinde `demo`) |
| `-dir path` | `./<name>` | Projenin scaffold edileceği dizin |
| `-module path` | `<name>` | `go.mod`'a yazılan module path |
| `-force` | kapalı | Dizin boş olmasa da scaffold eder |

```sh
collage new myblog                                   # into ./myblog, module "myblog", one page
collage new myblog --template demo                   # the demos, their tests, a .env.example
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
```

Flag'leri tek ya da çift tireyle yazabilirsiniz. Flag'ler addan önce de sonra da
gelebilir. Tam olarak bir ad vermeniz gerekir. Hiç ad vermemek ya da birden fazla
ad vermek kullanım hatasıdır. Hedef dizin varsa ve boş değilse, `-force`
vermediğiniz sürece komut reddedilir. `-force` verdiğinizde ise scaffold'un
yazdığı dosyalar, dizindeki aynı adlı dosyaların yerine geçer.

**Her projede** şunlar bulunur: `go.mod`; config'i, static mount'u,
[aşağıda](#the-contract-with-maingo) anlatılan CLI sözleşmesini ve
[plugin komutlarının](#plugin-commands) dispatch'ini içeren bir `main.go`; bütün
route'ları register eden bir `routes.go`; sitenin adını `WithTitle` ile veren ve
hiç slot tanımlamayan bir layout (böylece bir page'in kendi başlığı sitenin
başlığının yerine geçer); template'ine proje adını `collage.Value` ile veren bir ana
sayfa (içinde hiçbir şey veri çekmediği için `Static()` demeden static'tir);
`static/`, bir `.gitignore` ve bir README.

**`--template demo`**, projeye canlı demolardan oluşan bir page ekler:
JSON dönen bir action, kendi page'ine post eden bir form, kendi URL'si olan bir
fragment ve bir JSON document. Demo projesi ayrıca bir not-found page'i,
[`collagetest`](/docs/testing) ile yazılmış, her biri için testleri,
`plugins-config.json`'ı, bir favicon'u ve bir `.env.example`'ı da ekler.

**Yerleşim alana göredir** (v0.40.0'dan beri), yani gerçek bir uygulamanın
yerleşimidir:

```text
pages/<area>/<name>.go            a page: its layouts, content, path, actions
fragments/layouts/main.go         Master(), the layout every page wraps itself in
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

İş bittiğinde sonraki adımları yazdırır. `.env.example` dosyası olan demo
projesinde, `collage dev`'den önce bir de `cp .env.example .env.development` satırı
yazdırılır:

```sh
cd myblog
go mod tidy
collage dev
```

## collage add

```sh
collage add <page|fragment|action|document> <[area/]name> [flags]
```

Scaffold'un yerleşimine uygun bir page, fragment, action ya da document yazar ve
onu `routes.go` içinde register eder (v0.40.0'dan beri):

```sh
collage add page blog/post          # pages/blog/post.go + fragments/pages/blog/post.go
collage add page blog/post --file   # its template in templates/pages/blog/post.html
collage add fragment blog/sidebar   # fragments/pages/blog/sidebar.go, for a slot
collage add action blog/comment     # actions/blog.go + actions/funcs/blog.go
collage add action blog/ping --path /api/ping
collage add document feed --path /feed.xml --type application/xml
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--file` | kapalı | Template'i inline değil, template kökü altındaki bir dosyada tutar |
| `--path pattern` | `/<area>/<name>` | URL. Path'i olmayan bir action, eklendiği page'de cevap verir |
| `--name name` | son parça | Page'in, action'ın ya da document'ın adı |
| `--locale code` | `Locale.Default` ya da `en` | Path'in ait olduğu locale |
| `--type type` | `text/plain; charset=utf-8` | Bir document'ın content type'ı |
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
- **Bir document**, `documents/<name>.go` dosyasıdır.
- **Bir fragment**, bir slot için page'i olmayan bir page içeriğidir. Komut, onu
  bağlayan `.WithSlotFragment(...)` satırını yazdırır.

Bir ad küçük harflerden, rakamlardan ve tirelerden oluşur. `--name` başka bir ad
vermedikçe komutun yazdığı her şey son parçanın adını alır: `blog/post`, `Post()`
ile kurulan `"post"` page'idir. Path'in locale'i, `main.go` onu nasıl yazıyorsa
`Locale.Default`'tur; orada bir literal değilse `en`'dir; ya da `--locale` ile
verilendir. Bir template dosyası `Template.Root` altına, `Extension`'ıyla birlikte
yazılır; bunlar da aynı şekilde okunur. Komut hangi locale'i kullandığını söyler.

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
cevap verir: `localhost` ya da onun altındaki bir ad, bir IP adresi veya
başlatıldığı `HOST`. Diğer her şeyi `403` ile reddeder. Başka bir sitedeki bir page,
kendi adının `127.0.0.1`'e çözümlenmesini sağlayabilir (DNS rebinding) ve böylece
`collage dev` ile same-origin olur. Bu durumda page'lerinizi, stack'leriyle birlikte
development hata sayfalarını ve programın çıktısını okuyabilirdi. Gönderdiği `Host`
ise kendi adıdır ve bunu değiştiremez. `collage dev`'e başka bir adla erişmek için
onu `HOST` bu ada ayarlanmış olarak başlatın.

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
ve kendisine verilen `HOST` ve `PORT` üzerinde dinler. Development modu template'leri
ve static dosyaları her request'te diskten okur. Bu yüzden onları düzenlediğinizde
rebuild gerekmez ve rebuild yapılmaz; bunun yerine tarayıcınızdaki page kendini
yeniler. Programınızın diskten kendisi okuduğu içerik de (örneğin Markdown), dizini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtildiğinde page'i
yeniler (v0.10.0'dan beri). Development modunun başka neleri değiştirdiğini
[Template'ler](/docs/templates#reloading-in-development) sayfasında bulabilirsiniz.

### Development build'leri hiçbir şey embed etmez

v0.46.0'dan beri `collage dev`, `go build -tags collage_dev` ile build eder.
Scaffold da `//go:embed` satırlarını `//go:build !collage_dev` ile başlayan
`embed.go` dosyasında tutar. Yanındaki `embed_dev.go` (`//go:build collage_dev`)
aynı iki değişkeni boş olarak tanımlar. Development modu `templates/` ve `static/`
dizinlerini zaten diskten okur, yani bir development build'inin embed edeceği bir
şey yoktur. `collage build`, `collage export` ve düz bir `go build` tag vermez ve
eskisi gibi embed eder.

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

Düzeltmek için iki `//go:embed` değişkenini `main.go`'dan (ve import'larındaki
`"embed"`'i) çıkarıp kendi dosyalarına taşıyın:

```go
// embed.go
//go:build !collage_dev

package main

import "embed"

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS
```

```go
// embed_dev.go
//go:build collage_dev

package main

import "embed"

var (
	templatesFS embed.FS
	staticFS    embed.FS
)
```

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
  okunur.
- Dosyanın olmaması hata değildir. Bir dosya okunduğunda adı stderr'e yazdırılır.
- Dosya her yeniden başlatmada tekrar okunur. Dosya izlendiği için onu
  düzenlediğinizde program yeni değerlerle yeniden başlar.

Bu dosyaları yalnızca `collage dev` okur. `collage build`, `collage export` ve
derlenmiş binary, ortam değişkenlerini çalıştıkları ortamdan alır.

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

`collage-inspect`, `collage.InspectCommand`'dır. `DispatchCommands` bu komutu bir
plugin'e iletmez, kendisi cevaplar: `App.Inspect()`'in döndüğü
`collage.Inspection`'ı girintili JSON olarak yazdırır. Adı prefix'lidir, böylece
hiçbir plugin'in komut adını elinden almaz. Scaffold edilen `main.go`, flag'lerinden
sonraki kelimeyi `DispatchCommands`'a verir (bkz.
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
  {"name": "layout", "template": "layouts/default.html", "dataType": "nil"},
  {"name": "legacy", "template": "pages/legacy.html", "handler": true, "dataType": null, "typeCheck": false}
],
"types": {
  "blog.Post": {
    "kind": "struct",
    "fields": [{"name": "Title", "type": "string"}, {"name": "Author", "type": "*blog.User"}],
    "methods": [{"name": "URL", "args": 0, "returns": "string"}]
  },
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
  içerir. Bir tip bir kez listelenir, diğer her yerde adıyla anılır. Böylece
  özyinelemeli bir tip sonsuza gitmez. İsimsiz bir struct, Go yazımıyla
  listelenir. Diğer isimsiz bileşik tipler bir tip string'inin içinde yazılır
  (`[]blog.Comment`). `time.Time` ve `template.HTML` gibi standart kütüphane
  tipleri ayrıntılandırılmaz, yalnızca adıyla anılır. Hiçbir fragment'in bilinen bir veri
  tipi yoksa `types` yazılmaz.

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

## main.go ile sözleşme

`dev`, `export` ve `inspect`, `main.go`'nuzun yaptığı üç şeye dayanır. Scaffold
edilen `main.go` üçünü de yapar. Üçüncüsü, flag'lerden sonraki kelimeleri
`collage.DispatchCommands`'a vermektir. Plugin'lerinizin komutlarını çalıştıran da
budur:

| Komut | Çalıştırdığı | `main.go`'nuzun yapması gereken |
| --- | --- | --- |
| `collage dev` | önce `go build -tags collage_dev`, sonra binary; `COLLAGE_DEV=1` ve dinlenecek `HOST` ile `PORT` ayarlı olarak | `COLLAGE_DEV` `1` olduğunda development modunu açmak ve `HOST` ile `PORT` üzerinde dinlemek |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | `-collage-build`, `-out` ve `-clean` flag'lerini parse etmek; `-collage-build` verildiğinde sunmak yerine `<dir>` dizinine render etmek |
| `collage inspect` | `go run . collage-inspect` | flag'lerden sonraki kelimeleri `collage.DispatchCommands`'a vermek |

`collage build`'in `main.go`'dan hiçbir beklentisi yoktur. Derleme, `go build`'in
kendisine hiçbir şey söylenmeden yaptığı bir iştir.

Scaffold edilen sürümün, yalnızca önemli kısmı bırakılmış hâli:

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
		os.Exit(code)
	}

	if *buildFlag {
		if err := staticBuild(app, *outFlag, *cleanFlag); err != nil {
			log.Fatalf("static build: %v", err)
		}
		return
	}

	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

`main.go`'yu yeniden yazarsanız bunların hepsinin çalışmaya devam ettiğinden emin
olun. Aksi hâlde `collage dev`, `collage export`, `collage inspect` ve
plugin'lerinizin komutları projenizde işe yarar hiçbir şey yapmaz.

## Plugin komutları

Bir plugin, `Host.RegisterCommand` ile bir komut ekleyebilir. **`collage` CLI bu
komutu çalıştırmaz.** CLI uygulamanızı hiçbir zaman yüklemez. Bu yüzden ancak
plugin'leriniz başladıktan sonra var olan bir komuta CLI erişemez ve
`collage <plugin-command>` bilinmeyen bir komut olarak kalır.

Bu komutları kendi programınız `collage.DispatchCommands` ile dispatch eder.
Scaffold edilen `main.go` da bunu yapar (v0.10.0'dan beri; daha önce scaffold
edilmiş bir proje [yukarıdaki](#the-contract-with-maingo) bloğu kopyalayabilir).
Flag'ler parse edilip uygulama kurulduktan sonra geriye kalan bir kelime komut
olarak yorumlanır:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
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

Çıkış kodları CLI'ınkilerle aynı mantığı izler. Başarı için `0` döner. Açılış
hatası, çalışıp başarısız olan bir komut ya da `Run`'ı olmayan bir komut için `1`
döner. Nil bir app, argüman verilmemesi ya da hiçbir plugin'in register etmediği bir
kelime (`ErrUnknownCommand`) için `2` döner. Hiçbir komuta karşılık gelmeyen bir
kelime, yanlışlıkla başlatılmış bir sunucuya değil, kullanım hatasına yol açar.
Hiçbir plugin'in register etmediği bir kelimeye yine de cevap verilir:
[`collage inspect`](#collage-inspect) komutunun çalıştırdığı `collage-inspect`.
Hiçbir komut eşleşmediğinde sunucuyu başlatmayı tercih eden bir program,
`errors.Is(err, collage.ErrUnknownCommand)` ile kontrol edip çıkmak yerine devam
edebilir. Kendi kullanım metninizi yazdırmak isterseniz `app.Commands()`
plugin'lerin register ettiği komutları listeler. Dispatch işlemi uygulamayı başlatır
ve bu da register aşamasını kapatır. Sonradan çağrılan `ListenAndServe` aynı başlatmayı
yeniden kullanır. Böyle bir komutun nasıl yazıldığı
[Plugin yazmak](/docs/writing-plugins#commands) sayfasında anlatılır.
