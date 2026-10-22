---
description: Go'yu ve collage CLI'ını kurun, collage new ile bir proje scaffold edin ve onu collage dev ile çalıştırın.
---

# Kurulum

Bir collage projesi sıradan bir Go modülüdür. `collage` komut satırı aracı bu
projeyi scaffold eder, siz üzerinde çalışırken çalıştırır ve yayına alırken build
eder. Framework'ün kendisi ise projenin import ettiği bir kütüphanedir. İkisi de Go
dışında hiçbir şeye ihtiyaç duymaz.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 364" role="img" aria-labelledby="inst-flow-title">
<title id="inst-flow-title">Sıfırdan çalışan bir projeye böyle gelinir. go install, collage CLI'ını PATH'inize koyar. collage new mysite varsayılan olarak minimal projeyi yazar: tek bir page'i saran bir layout ve bir stylesheet. --template demo ile demoları, testlerini ve bir .env.example yazar. go mod tidy, go.mod'a collage'ı ekler ve go.sum'ı yazar. Ardından collage dev localhost:6060'ta serve eder ve her request'i, kendine ait loopback bir adreste çalışan programa iletir. Kaydedilen bir Go dosyası yeniden build edilip yeniden başlatılır, kaydedilen bir template ya da static dosya page'i yeniler.</title>
<defs><marker id="inst-flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="50" y="8" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">go install …/collage@latest</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="48" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="64" text-anchor="middle">collage new mysite</text>
<line class="d-flow" x1="160" y1="72" x2="83" y2="88" marker-end="url(#inst-flow-arrow)" />
<line class="d-flow" x1="160" y1="72" x2="237" y2="88" marker-end="url(#inst-flow-arrow)" />
<rect class="d-a" x="10" y="90" width="145" height="26" rx="4" /><text class="d-on-a" x="82" y="107" text-anchor="middle">minimal</text>
<rect class="d-b" x="165" y="90" width="145" height="26" rx="4" /><text class="d-on-b" x="237" y="107" text-anchor="middle">--template demo</text>
<text class="d-small" x="82" y="130" text-anchor="middle">varsayılan: layout,</text>
<text class="d-small" x="82" y="142" text-anchor="middle">page, stylesheet</text>
<text class="d-small" x="237" y="130" text-anchor="middle">demolar ve testleri,</text>
<text class="d-small" x="237" y="142" text-anchor="middle">bir .env.example</text>
<line class="d-flow" x1="82" y1="148" x2="140" y2="164" marker-end="url(#inst-flow-arrow)" />
<line class="d-flow" x1="237" y1="148" x2="180" y2="164" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="166" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="182" text-anchor="middle">go mod tidy</text>
<text class="d-small" x="160" y="204" text-anchor="middle">go.mod'da collage, go.sum</text>
<line class="d-flow" x1="160" y1="208" x2="160" y2="222" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="224" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="240" text-anchor="middle">collage dev</text>
<line class="d-flow" x1="160" y1="248" x2="160" y2="262" marker-end="url(#inst-flow-arrow)" />
<rect class="d-frame" x="10" y="264" width="300" height="92" rx="8" />
<rect class="d-pill" x="80" y="272" width="160" height="22" rx="11" /><text x="160" y="287" text-anchor="middle">localhost:6060</text>
<text class="d-small" x="22" y="312">Go dosyası kaydedildi: rebuild, restart</text>
<text class="d-small" x="22" y="328">template ya da static dosya: yenileme</text>
<text class="d-small" x="22" y="344">program: kendine ait loopback adres</text>
</svg>
<figcaption>Sıfırdan çalışan bir projeye böyle gelinir. <code>go install</code> CLI'ı <code>PATH</code>'inize koyar. <code>collage new</code>, siz <code>--template demo</code> istemedikçe minimal projeyi yazar. Demo template'i buna demoları, testlerini ve bir <code>.env.example</code> ekler. <code>go mod tidy</code> framework'ü çeker ve sürümünü <code>go.mod</code> ile <code>go.sum</code>'a kaydeder; <code>--with</code> bunu sizin yerinize yapar. <code>collage dev</code> varsayılan olarak 6060 portunda dinler ve her request'i, kendine ait loopback bir adreste çalıştırdığı programınıza iletir. Kaydedilen bir Go dosyası yeniden build edilip yeniden başlatılır, kaydedilen bir template ya da static dosya ise yalnızca page'i yeniler.</figcaption>
</figure>

## Go'yu kurun

collage **Go 1.26 veya daha yeni bir sürüm** ister. Go'yu [go.dev/dl](https://go.dev/dl)
adresinden ya da paket yöneticinizden kurun ve kontrol edin:

```sh
go version
```

## CLI'ı kurun

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

`go install` binary'yi `$(go env GOBIN)` dizinine koyar. `GOBIN` ayarlı değilse
`$(go env GOPATH)/bin` dizinini kullanır. Bu dizin `PATH`'inizde olmalıdır. Olup
olmadığını kontrol edin:

```sh
collage version
```

Yazdırılan sürüm build'in kendisinden okunur, yani `go install` hangi sürümü
çektiyse o görünür. `collage help` bütün komutları listeler, `collage help new` ise
tek bir komutu açıklar.

## Proje oluşturun

```sh
collage new mysite
```

Bu komut `./mysite` içine, module path'i `mysite` olan, çalışmaya hazır bir proje
yazar ve sonraki adımları ekrana basar:

```text
Scaffolded "mysite" in mysite

Next steps:
  cd mysite
  go mod tidy
  collage dev
```

Demo projesinde (aşağıda) bir de `.env.example` vardır, bu yüzden onun adımlarında
`collage dev`'den önce `cp .env.example .env.development` satırı da yer alır.

`go mod tidy` framework'ü çeker. Scaffold edilen `go.mod` yalnızca modülün adını
ve Go sürümünü içerir. Bu yüzden projenin hangi collage sürümüyle build edildiğini
kaydeden şey ilk `tidy` çalıştırmasıdır.

Birkaç flag, projenin nereye ve nasıl yazılacağını değiştirir. Flag'leri tek ya da
çift tireyle yazabilirsiniz. Proje adından önce de gelebilirler, sonra da:

| Flag | Etkisi |
| --- | --- |
| `--template demo` | Demoları, testlerini ve bir `.env.example` oluşturur. v0.32.0'dan beri varsayılan `--template minimal`'dir: tek bir page'i saran bir layout ve bir stylesheet |
| `-dir path` | Projeyi `./<name>` yerine `path` içine scaffold eder |
| `-module path` | `go.mod`'daki module path'i belirler, örneğin `github.com/you/mysite`. Varsayılan değer proje adıdır |
| `-force` | Boş olmayan bir dizine de scaffold eder |
| `--with plugins` | Proje yazıldıktan sonra plugin ekler: `@prod` gibi bundle'lar ve `elagoht/session` gibi plugin'ler, virgülle ayrılmış (v0.59.0'dan beri). Plugin'ler çekildiği için `go mod tidy` artık bir sonraki adım değildir |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite --template demo
collage new mysite -dir . -force
collage new mysite --with @site,@prod
```

### Demolarla ya da demolar olmadan

`--template demo`, bir ana sayfa ve canlı demolarla dolu bir `/features` page'i
oluşturur. Bu demolar şunlardır:

- bir API action'ına post edip cache'lenmiş bir page'i tag ile invalidate eden bir
  buton;
- kendi page'ine post edilen bir HTML form;
- kendi URL'sinde açılabilen bir saat fragment'i;
- `/healthz` adresindeki bir JSON document.

Bunların her birini çalışırken görmenin en hızlı yolu budur. Her birinin
arkasındaki kod da tek oturuşta okunacak kadar kısadır.

Varsayılan template olan `minimal` aynı `main.go`'yu ve aynı dizin yapısını
oluşturur. İçinde yalnızca "hello" diyen tek bir page'i saran bir layout ve dark
mode'u olan bir stylesheet bulunur. Silmeniz gereken hiçbir şey yoktur. Gerçek bir
siteye başlarken bunu kullanın. [İlk page'iniz](/docs/your-first-page) rehberi de
minimal bir projeden başlar.

## Scaffold'un içinde neler var

Minimal bir proje şöyle görünür:

```text
mysite/
├── main.go                     newApp, the configuration; main hands it to collage.Main
├── embed.go                    files: templates/, static/ and plugins-config.json, embedded
├── embed_dev.go                files as the project directory, for collage dev's builds
├── routes.go                   every page, document and action — a new route goes here
├── plugins.go                  the plugin list; collage plugin add and del edit it
├── plugins-config.json         plugin settings, keyed by plugin name
├── .env.example                the environment variables, to copy to .env
├── go.mod
├── .gitignore
├── README.md
├── pages/
│   └── home.go                 the home page: its layout, content and path
├── fragments/
│   ├── layouts/main.go         Master(), the layout every page shares, and the site's title
│   └── pages/home.go           the home page's content and its data
├── templates/
│   ├── layouts/master.html     the layout's HTML, with {{slot "content"}}
│   └── pages/home.html         <h1>Hello from {{.Name}}</h1>
└── static/
    └── app.css                 the background and text colour, light and dark
```

Kod, alanlara (area) göre düzenlenir. Bir page `pages/<area>/<name>.go`
dosyasındadır, içeriği ise `pages/` yapısını aynen yansıtan
`fragments/pages/<area>/<name>.go` dosyasında, onun karşılığında durur. Proje
büyüdükçe geri kalan her şey demo projesinin koyduğu yere gider:

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

`collage add page blog/post` bir page'i oraya yazar ve `routes.go`'da register
eder. `collage add` aynısını bir fragment, bir action ya da bir document için de
yapar; bkz. [CLI](/docs/cli#collage-add).

Demo projesi bunlara ek olarak şunları içerir:

- testleri `app.Handler()`'ı [`collagetest`](/docs/testing) ile ve sunucu
  başlatmadan çalıştıran `main_test.go`;
- bir not-found page'i;
- bir favicon;
- `demo` ile `errors` alanlarındaki page'ler. Bu page'lerle birlikte action'ları,
  bir document, `data/count/` içinde bir sayaç ve bunların template'leri ile
  script'leri de gelir.

Scaffold'daki birkaç şeyi, değiştirmeden önce bilmekte fayda var.

**`main.go` CLI ile bir sözleşmeye uyar.** `collage dev` programı, environment'ında
`COLLAGE_DEV=1` ve dinleyeceği `HOST` ile `PORT` tanımlı olarak çalıştırır.
`collage export` ise programı `-collage-build -out <dir>` ile çalıştırır.

Scaffold edilen `main.go` yaklaşık elli satırdır, çünkü `main` yalnızca
`collage.Main(newApp)`'i çağırır (v0.68.0'dan beri) ve bunların hepsine
`collage.Main` uyar. `newApp`'e bir `collage.Env` verir: değişkenin belirlediği
`Dev`, dinlenecek `Host` ile `Port` ve `CacheDir`. Ardından kendisine söylenen
adreste serve eder, flag verildiğinde ise siteyi serve etmek yerine dosyalara
render eder.

Flag'lerden sonra gelen bir kelime, yani `go run . <command>`, bir
[plugin'in komutunu](/docs/plugins) çalıştırır. Bu komutu da `collage.Main`
dispatch eder. `collage.Main`'i kendi yazdığınız bir `main` ile değiştirirseniz
bunların hepsinin çalışmaya devam etmesini sağlayın. Aksi hâlde bu komutlar işe
yarar hiçbir şey yapmaz. Ayrıntılar [CLI referansında](/docs/cli#the-contract-with-maingo) yer
alır.

**`plugins.go` plugin listesidir.** Bu dosyadaki `plugins()` fonksiyonu
plugin'leri sırayla döndürür. `main.go` onu bir kez, `Plugins: plugins()` olarak
verir. Liste başlangıçta boştur. `collage plugin add elagoht/sitemap` listeye bir
girdi yazar, `collage plugin del` ise bir girdiyi kaldırır. Bkz.
[Plugin kullanmak](/docs/plugins#registering-a-plugin).

**`newApp`, `main`'den ayrıdır.** `newApp` bütün uygulamayı kurar: config'i,
`routes.go`'daki route'ları ve `/static/` mount'unu. Sonra bir sunucu başlatmadan
uygulamayı döner. Demo projesindeki `main_test.go` aynı fonksiyonu çağırır ve
`app.Handler()`'ı `collagetest` ile çalıştırır. Böylece testler sitenin ayrıca
kurulmuş ikinci bir kopyasını değil, gerçekten çalışan siteyi test eder.
[Test yazmak](/docs/testing) sayfasına bakın.

**Template'ler, static dosyalar ve plugin ayarları embed edilir.** `embed.go`
içindeki tek bir `//go:embed all:templates all:static plugins-config.json` satırı
bunları binary'nin içine tek bir dosya sistemi, yani `files` olarak embed eder. Bu
sayede binary herhangi bir çalışma dizininden çalışabilir. `collage dev` ise
`-tags collage_dev` ile build eder: bu tag `embed.go` yerine `embed_dev.go`'yu alır
ve hiçbir şey embed edilmez. Orada `files` diskteki proje dizinidir. Böylece
düzenlediğiniz bir template bir sonraki request'te yine görünür, `files` verilen
bir plugin düzenlemekte olduğunuz kopyayı okur ve Go build cache'i her kaydetmede
onların bir kopyasını daha biriktirmez (v0.46.0'dan beri;
[collage dev](/docs/cli#development-builds-embed-nothing) sayfasına bakın).

`newApp`, `Env` üzerinden `files`'tan iki şey okur. `env.PluginConfig(files)`,
`plugins-config.json` dosyasıdır: çalışma dizininde bu dosya varsa o okunur, böylece
bir deployment bir ayarı yeniden build etmeden değiştirebilir; yoksa binary'deki
kopya okunur. `env.Static(files, "static")` ise `/static/`'in serve ettiği şeydir.

**Development'ta static dosyalar `os.DirFS` ile değil, `os.OpenRoot` ile mount
edilir.** Bunu `env.Static` yapar. Bir `os.Root`, dizinin dışına çıkan bir symlink'i
reddeder. `os.DirFS` ise onu takip eder. Production'da embed edilmiş kopya serve
edilir. Bu kopya `fs.Sub` üzerinden sunulur, böylece `static/` dizini URL'de ikinci
bir path segment olarak görünmez. [Static asset'ler](/docs/assets) sayfasına bakın.

**Production'da render edilen page'ler diskte, `.cache/` altında cache'lenir.**
Development'ta page cache hiç okunmaz. Böylece yaptığınız bir değişiklik hiçbir
zaman stale bir page'in arkasında kalmaz. [Caching](/docs/caching) sayfasına bakın.

## Environment dosyası

Her proje bir `.env.example` dosyasıyla gelir (minimal proje de v0.68.0'dan beri).
Yorumları çıkarıldığında dosya şöyledir:

```sh
COLLAGE_CSRF_KEY=
PORT=6060
HOST=localhost
BASE_URL=http://localhost:6060
```

Bu dosyayı git'in yok saydığı `.env.development` dosyasına kopyalayın:

```sh
cp .env.example .env.development
```

`collage dev`, `.env.development` içindeki değişkenleri programın environment'ına
ekler. `.env.development` yoksa `.env` içindekileri ekler. Her zaman tek bir dosya
okur, ikisini birden okumaz. Kurallar kısadır:

- Shell'inizde zaten tanımlı bir değişken önceliklidir. Bu yüzden
  `PORT=4000 collage dev` beklendiği gibi çalışır. Dosyada ne yazarsa yazsın
  `COLLAGE_DEV=1` her zaman tanımlanır.
- `HOST` ve `PORT`, `collage dev`'in kendisinin dinlediği adrestir ve başlarken bir
  kez okunur. Programa ise loopback bir adreste kendine ait bir `HOST` ve `PORT`
  verilir. Ayrıntılar için [aşağıya](#run-it) bakın.
- Dosya `KEY=value` satırlarından, `#` ile başlayan yorumlardan ve boş satırlardan
  oluşur. Satır başında `export ` öneki ve değerin etrafında tırnak kullanabilirsiniz.
- Hatalı bir satır atlanmaz, dosya adı ve satır numarasıyla birlikte raporlanır.
  Çünkü atlanan bir satır, sizin yazdığınız ama programın hiç görmediği bir
  ayar demektir. Siz satırı düzeltene kadar hiçbir şey başlatılmaz ya da yeniden
  başlatılmaz. Zaten çalışan bir build serve etmeye devam eder. `collage dev` de
  izlemeyi sürdürür, yani düzeltmeyi kaydettiğinizde süreç kaldığı yerden devam
  eder.
- Hiç dosya olmaması bir hata değildir. Dosya varsa adı stderr'e yazılır.
- Shell'iniz de dosya da `BASE_URL`'i ayarlamıyorsa değeri `collage dev`'in kendi
  adresi olan `http://localhost:6060`'tır (v0.68.0'dan beri). Mutlak URL yazan bir
  plugin, örneğin bir sitemap ya da bir canonical link, origin olmadan başlamayı
  reddeder. Development'ta doğru origin de budur.

**Program da bu dosyaları okur** (v0.68.0'dan beri). Scaffold edilen `main`'in
çağırdığı `collage.Main`, `newApp` çalışmadan önce çalışma dizinindeki `.env`'i aynı
kurallarla yükler. `COLLAGE_DEV=1` altında önce `.env.development`'a bakar. Böylece
`os.Getenv("BASE_URL")` dosyayı `go run .`'ta, `collage export`'ta ve build edilmiş
binary'de görür, `source .env` gerekmez. Değişkenlerini kendisi ayarlayan bir
sunucuda o değişkenler önceliklidir ve dosyanın olmaması bir hata değildir.

`BASE_URL`, okuyucuların siteye ulaştığı adrestir. Bir sitemap, bir canonical link
ya da bir feed bu adresi bir path'in önüne yazar. `PORT`'u değiştirirseniz bunu da
değiştirin. Production'da sitenin kendi adresidir, örneğin `https://example.com`.

`COLLAGE_CSRF_KEY`, form'ların taşıdığı token'ları imzalar. Development sırasında
boş bırakabilirsiniz. Bu durumda her process için yeni bir key üretilir. Form
içeren herhangi bir şeyi deploy etmeden önce bir key tanımlayın. Aksi hâlde restart
öncesinde gönderilmiş her form, restart'tan sonra reddedilir.

İçinde form olan cache'lenmiş bir page de key'e bağlıdır. Yeni bir key ile restart
ettikten sonra o page, içinde eski key'in token'ı ile serve edilmez, baştan render
edilir. [Caching](/docs/caching#the-namespace) sayfasına bakın. Bir key'i şöyle
üretebilirsiniz:

```sh
openssl rand -hex 32
```

## Çalıştırın

```sh
collage dev
```

Site [http://localhost:6060](http://localhost:6060) adresinde açılır (v0.32.0'dan
önce port 3000'di). Bu adres `collage dev`'in kendisidir: `HOST` ve `PORT`'u
programınızın okuyacağı şekilde okuyup orada dinler. Her request'i, loopback bir
adreste çalıştırdığı programa iletir. Program başlarken gelen bir request,
programı bekler.

`collage dev` çalışırken dört şey olur.

### Go değişikliklerinde yeniden build alınır

`collage dev` projeyi `go build` ile build eder ve binary'yi çalıştırır. Programı
oluşturan dosyaları izler: `.go` dosyalarını (testler hariç), `go.mod`, `go.sum`
ve environment dosyasını. Bunlardan biri değiştiğinde projeyi yeniden build eder.

Önce yeni build alınır. Eski process ancak yeni build başarıyla derlendikten sonra
düzgünce durdurulur ve yenisi başlatılır. Derlenmeyen bir değişiklikte son sağlam
build serve etmeye devam eder ve compiler'ın hatası terminale yazılır. Böylece bir
yazım hatası yüzünden `localhost:6060` hiçbir zaman boş kalmaz.

Art arda yapılan kayıtlar tek bir rebuild tetikler. Gizli dizinler, `bin`, `dist`,
`node_modules`, `testdata` ve `vendor` hiçbir zaman izlenmez. Böylece çalışan
programın yazdığı hiçbir dosya kendi rebuild'ini tetikleyemez. Kendiliğinden
kapanan bir program da döngü halinde yeniden başlatılmaz. Başlangıçta oluşan bir
panic ya da template'i eksik olan bir page buna örnektir. Program bir sonraki
değişikliğinizde yeniden başlar.

### Template'ler ve static dosyalar diskten okunur

Development mode'da her template, her render'dan önce diskten yeniden parse edilir.
Static dosyalar da diskteki `static/` dizininden serve edilir. İkisini düzenlemek
de rebuild gerektirmez ve rebuild yapılmaz.

### Tarayıcı kendini yeniler

Development'ta serve edilen her page küçük bir script taşır. Bu script bir template
ya da static dosya değiştiğinde ve program bir rebuild'den sonra yeniden ayağa
kalktığında page'i yeniler. Bir dosyayı kaydedip tarayıcıya bakmanız yeterlidir,
kurmanız gereken bir şey yoktur.

Programınızın diskten kendisinin okuduğu içerik, örneğin Markdown ya da JSON, ne
template ne de static dosyadır. Bu içerikteki değişikliklerin de page'i
yenilemesini istiyorsanız dizinini [`Config.DevWatch`](/docs/configuration#devwatch)
içinde belirtin (v0.10.0'dan beri).

Script hiçbir zaman production'daki bir page'e eklenmez. Bir form gönderiminin
response'una da eklenmez, çünkü page'i yenilemek form'u yeniden gönderirdi.

Script sunucuya açık bir stream tutar. Bir tarayıcı HTTP/1.1 üzerinden bir
origin'e, tüm sekmeleri toplamında en fazla altı bağlantı açar. Açık her
development sekmesinde bir stream olunca, yan yana açık altı page altı bağlantının
hepsini tutardı ve başka hiçbir şey yüklenmezdi: ne yedinci sekme yüklenirdi ne
de bu page'lerin kendi fragment'leri.

v0.20.0'dan beri her development sekmesi tek bir shared worker üzerinden dinler:
`/_collage/reload-worker.js`. Bu worker, kaç page açık olursa olsun hepsi için tek
bir stream tutar. Shared worker olmayan yerlerde her sekme kendi stream'ini tutar
ve v0.18.1'den beri olduğu gibi gizliyken onu kapatır, yeniden görüldüğünde tekrar
bağlanır. Gizliyken bir şey değiştiyse page o anda yenilenir.

Hiç kapanmayan bir stream, yüklenmesi hiç bitmeyen bir page demektir. Bir ekran
görüntüsü aracı ya da bir end-to-end test de tam olarak bunun bitmesini bekler.

Playwright, Puppeteer ya da Selenium'un yönettiği bir tarayıcı
`navigator.webdriver`'ı set eder ve script orada bağlanmaz (v0.18.0'dan beri).
Bunu set etmeyen araçlar için, yani headless Chrome'un `--screenshot` ve
`--dump-dom` seçenekleri için, URL'ye `?collage-reload=0` ekleyin. Page script
olmadan serve edilir.

### Hatalar page'in üzerinde görünür

Development'ta başarısız olan bir fragment sessizce ortadan kaybolmaz. Page,
üzerinde bir panel ile serve edilir. Bu panel fragment'in adını ve hatasını
gösterir, template hatalarında dosya ve satırı da belirtir. Hatayı bir fallback
kapatmış olsa bile panel yine görünür. Page'in tamamı başarısız olduğunda,
built-in error page önce nedeni gösterir: başarısız olan çağrının template'ini,
satırını, sütununu ve çağrının ne döndüğünü. Ardından hatanın başladığı fragment'i
söyler ve bir panic'in stack'i dahil bütün hata zincirini gösterir. Kendi
yazdığınız bir error page'in üstünde de aynı panel çıkar ve hangi page'in yerine
gösterildiğini belirtir.

Cevap verecek bir program olmadığında (program başlarken kapandıysa ya da ilk
build derlenmediyse) page, reddedilen bir bağlantı yerine programın ya da
derleyicinin yazdırdıklarını gösteren bir 503 olur. Açık olan bir page yenilenerek
bu page'e geçer ve bir değişiklik programı geri getirdiğinde yeniden yenilenir.
Başladıktan 10 saniye sonra hâlâ kendisine verilen adreste dinlemeyen bir program
da bu page'de, kendisine verilen adresle birlikte belirtilir.

Bunların hiçbiri development dışında yoktur. Production'daki bir error page tek bir
genel cümle gösterir. Çünkü hata mesajları hostname'ler, dosya yolları ve
credential'lar içerir. Bunları bir yabancının eline verme ihtimali en yüksek
response da tam olarak error page'dir. [Hatalar](/docs/errors) sayfasına bakın.

## Örnek: bir ekip arkadaşı için proje kurmak

Başka birinin de üzerinde çalışacağı bir projeye başlıyorsunuz: ekibin iş listesi
uygulaması. Projeyi bir kez scaffold edip commit'lersiniz. Sonra her biriniz onu
kendi makinesinde, kendi ayarlarıyla çalıştırır.

Kendi makinenizde şunları bir kez çalıştırın:

```sh
collage new team --template demo -module github.com/you/team
cd team
go mod tidy
```

Scaffold edilen `.gitignore`, paylaşılanla paylaşılmayanın arasındaki çizgiyi
zaten çeker. Yorumları çıkarıldığında kuralları şunlardır:

```text
/bin/team
/bin/team.exe
/dist/
/.cache/
*.exe
/.env
/.env.*
!/.env.example
```

| Yol | Commit'lenir mi | Neden |
| --- | --- | --- |
| `go.mod`, `go.sum` | evet | Herkesin build aldığı collage sürümünü sabitler |
| `.env.example` | evet | Kopyalanacak değişkenleri tutar, içinde hiçbir secret yoktur |
| `collage new`'in yazdığı diğer her şey | evet | Projenin kendisi |
| `.env.development`, `.env` | hayır | Herkesin kendi değişkenleri |
| `/bin/team` | hayır | `collage build`'in yazdığı dosya |
| `/dist/` | hayır | `collage export`'un yazdıkları |
| `/.cache/` | hayır | Restart'lar arasında saklanan render edilmiş page'ler |

Ekip arkadaşınızın da yukarıda anlatıldığı gibi Go'ya ve CLI'a ihtiyacı vardır.
Repository'yi clone'ladıktan sonra demo projesinin `README.md`'sinde listelenen
adımları çalıştırır:

```sh
cd team
go mod tidy
cp .env.example .env.development
collage dev
```

`go.mod` zaten bir collage sürümünü gerektirdiği için bu `go mod tidy` başka bir
sürüm seçmez, o sürümü çeker. Minimal bir projede `.env.example` yoktur:
[environment dosyası](#the-environment-file) bölümündeki gibi bir tane yazın ve
commit'leyin. `!/.env.example` satırı onu yok sayılan dosyaların dışında tutar.

**Herkesin `.env.development` dosyası kendine aittir.** Orada `COLLAGE_CSRF_KEY`
yukarıda anlatıldığı gibi boş kalabilir ya da o kişinin kendi key'ini tutabilir.
Deploy edilen sitenin kullandığı key, production environment'ının
secret'larına girer. Herkesin okuyabildiği `.env.example`'a asla girmez.

**Portu değiştirmek.** 6060 portu dolu olan biri, kendi `.env.development`
dosyasında başka bir port belirler. Demo projesinde `BASE_URL`'i de onunla
birlikte değiştirin. Bu değişken `collage dev`'in serve ettiği adrestir. Bir
sitemap, canonical bir link ya da bir feed de path'lerin önüne bu adresi koyar.

```sh
PORT=6161
HOST=localhost
BASE_URL=http://localhost:6161
```

`collage dev` başlarken adresi ve okuduğu dosyayı yazar:

```text
01:09:37 - collage dev: serving http://localhost:6161
01:09:37 - collage dev: loaded .env.development
```

Bu adres bir kez, başlarken belirlenir. Dosyaya yeni bir `PORT` kaydetmek
programı yeniden build eder, ama `collage dev` siz onu durdurup yeniden
başlatana kadar eski adreste dinlemeye devam eder. Tek bir çalıştırma için
`PORT=4000 collage dev` dosyaya dokunmadan aynı işi görür.

## Editör desteği

VS Code için Collage Snippets & Highlighter extension'ı vardır. Şunları ekler:

- **Snippet'ler**: Go için (`cpage`, `cfragd`, `caction`, `cplugin` …) ve
  template'ler için (`clayout`, `cslot`, `cform`, `cformv` …). Birçok Go snippet'i
  (`cpagef`, `cactionf`, `cvalid`, `credirect`, `cdataf`, `cmeta` …), dosyada
  `package` satırı yoksa onu da ekler: yanındaki dosyaların package'ı, `go.mod`'un
  yanında `main`, değilse dizinin adı. Kullandıkları import'ları da ekler;
  collage-validate ve collage-flash import'ları da buna dahildir.
- **Highlighting**: HTML içindeki `{{ … }}` renklendirilir. Template'ler HTML
  dosyası olarak kalır, bu yüzden Emmet, tag tamamlama ve biçimlendirme çalışmaya
  devam eder. Go'da
  [inline bir fragment](/docs/fragments-and-slots#inline-templates)'in template'i
  de HTML olarak renklendirilir; argümanlar tek satırda ya da birkaç satırda
  olabilir. `collage.InlineHTML` olarak tanımlanmış bir sabit de öyle.
- **Go içindeki HTML, HTML olarak düzenlenir** (0.6.0'dan beri): inline bir
  template'te tag ve attribute tamamlama, HTML snippet'leriniz, Emmet, hover,
  katlama, eşleşen tag ve yazarken kapanan tag'ler. `{{ … }}` içinde de aşağıdaki
  her şey çalışır.
- **Completion ve hover**: collage'ın, yayımlanmış plugin'lerinin ve Go'nun
  builtin'leri olan her template fonksiyonu için, imzası ve dokümantasyonuyla.
- **Projenizdeki adlar**: `{{pageURL "…"}}` içinde page'ler ve parametreleri,
  `{{fragmentURL "…" "…"}}` içinde fragment'ler, `{{slot "…"}}` içinde slot'lar ve
  `{{asset "…"}}` içinde mount edilmiş dosyalar.
- **Diagnostics ve go to definition**: var olmayan bir page, fragment ya da dosya
  için bir uyarı verir. F12 ile bir addan onu tanımlayan Go koduna gidersiniz.
- **`plugins-config.json` doğrulaması**: yayımlanmış her plugin ve
  [`collage.json`](/docs/writing-plugins#editor-support-collagejson) içeren her
  plugin için completion, açıklamalar ve yanlış yazılmış bir key için uyarı.

Extension, projenizdeki adları projeyi `collage dev` gibi derleyip
`collage-inspect` ile çalıştırarak öğrenir;
[`collage inspect`](/docs/cli#collage-inspect) de aynısını yapar. Bunu
`go.mod`'u collage'ı gerektiren her dizinde yapar ve bir Go dosyası her
kaydedildiğinde tekrarlar. collage v0.27.0 veya daha yeni bir sürümle çalışan bir
proje, hiçbir şey eklemeden bu komuta cevap verir.

Extension henüz VS Code Marketplace'te değildir. `.vsix` dosyasını
[son release'inden](https://github.com/Elagoht/collage-snippets-highlighter/releases/latest)
indirin ve kurun:

```sh
code --install-extension collage-snippets-highlighter-<version>.vsix
```

| Ayar | Varsayılan | Ne yapar |
| --- | --- | --- |
| `collage.completions` | `auto` | Completion ve hover'ın nerede sunulacağı: `auto`, yalnızca `go.mod`'u collage'ı gerektiren bir dizinde; `always`; ya da `never` |
| `collage.inspect` | `true` | Projeyi öğrenmek için projenin `collage-inspect` ile çalıştırılıp çalıştırılmayacağı. Bu komut uygulamayı sunmadan kurar. Bu yüzden kurulurken veritabanına bağlanan bir program burada da bağlanır |
| `collage.diagnostics` | `warning` | Var olmayan bir adın nasıl raporlanacağı: `off`, `information`, `warning` ya da `error` |
| `collage.goCommand` | `go` | Çalıştırılacak go komutu |

Snippet'ler ve highlighting her zaman açıktır. Extension'ın
[README'si](https://github.com/Elagoht/collage-snippets-highlighter#readme) bütün
snippet'leri ve renklendirilecek scope'ları listeler.

## Sırada ne var

[İlk page'iniz](/docs/your-first-page), minimal bir projede sıfırdan, veri kullanan
bir page oluşturur. Yayına almaya hazır olduğunuzda `collage build` bir binary
üretir, `collage export` ise static dosyalar yazar. [Deployment](/docs/deployment)
ve [Static export](/docs/static-export) sayfalarına bakın.
