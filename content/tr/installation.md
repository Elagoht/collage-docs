---
description: Go'yu ve collage CLI'ını kurun, collage new ile bir proje scaffold edin ve onu collage dev ile çalıştırın.
---

# Kurulum

Bir collage projesi sıradan bir Go modülüdür. `collage` komut satırı aracı bu
projeyi scaffold eder, siz üzerinde çalışırken çalıştırır ve yayına alırken build
eder. Framework'ün kendisi ise projenin import ettiği bir kütüphanedir. İkisi de Go
dışında hiçbir şeye ihtiyaç duymaz.

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
  cp .env.example .env.development
  collage dev
```

Minimal bir projede (aşağıda) `.env.example` yoktur, bu yüzden onun adımlarında `cp`
satırı yer almaz. `go mod tidy` framework'ü indirir. Scaffold edilen `go.mod`
yalnızca modülün adını ve Go sürümünü içerir. Bu yüzden projenin hangi collage
sürümüyle build edildiğini kaydeden şey ilk `tidy` çalıştırmasıdır.

Birkaç flag, projenin nereye ve nasıl yazılacağını değiştirir. Flag'leri tek ya da
çift tireyle yazabilirsiniz. Proje adından önce de gelebilirler, sonra da:

| Flag | Etkisi |
| --- | --- |
| `--template minimal` | Tek bir page'i saran bir layout ve bir stylesheet oluşturur, demo içermez. Varsayılan değer `--template demo`'dur |
| `-dir path` | Projeyi `./<name>` yerine `path` içine scaffold eder |
| `-module path` | `go.mod`'daki module path'i belirler, örneğin `github.com/you/mysite`. Varsayılan değer proje adıdır |
| `-force` | Boş olmayan bir dizine de scaffold eder |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite --template minimal
collage new mysite -dir . -force
```

### Demolarla ya da demolar olmadan

Varsayılan template olan `demo`, bir ana sayfa ve canlı demolarla dolu bir
`/features` page'i oluşturur. Bu demolar şunlardır: bir API action'ına post
edip cache'lenmiş bir page'i tag ile invalidate eden bir buton, kendi page'ine
post edilen bir HTML form, kendi URL'sinde açılabilen bir saat fragment'i ve
`/healthz` adresindeki bir JSON document. Bunların her birini çalışırken görmenin
en hızlı yolu budur. Her birinin arkasındaki kod da tek oturuşta okunacak kadar
kısadır.

`--template minimal` aynı `main.go`'yu ve aynı dizin yapısını oluşturur. İçinde
yalnızca "hello" diyen tek bir page'i saran bir layout ve dark mode'u olan bir
stylesheet bulunur. Silmeniz gereken hiçbir şey yoktur. Gerçek bir siteye
başlarken bunu kullanın. [İlk page'iniz](/docs/your-first-page) rehberi de minimal
bir projeden başlar.

## Scaffold'un içinde neler var

Minimal bir proje şöyle görünür:

```text
mysite/
├── main.go                     configuration, the static mount, the CLI contract, plugin commands
├── routes.go                   every page, document and action — a new route goes here
├── go.mod
├── .gitignore
├── README.md
├── pages/
│   └── home.go                 the home page: layout, content, path, and its data
├── fragments/
│   └── layouts/main.go         the layout fragment every page shares, and the site's title
├── templates/
│   ├── layouts/default.html    the layout's HTML, with {{slot "content"}}
│   └── pages/home.html         <h1>Hello from {{.Name}}</h1>
└── static/
    └── app.css                 the background and text colour, light and dark
```

Demo projesi bunlara ek olarak şunları içerir: testleri `app.Handler()`'ı sunucu
başlatmadan çalıştıran `main_test.go`, bir not-found page'i, `collage dev` için
kopyalanacak değişkenleri tutan `.env.example`, plugin ayarlarını plugin adına göre
tutan `plugins-config.json`, bir favicon ve template'leri ile script'leriyle
birlikte `actions/`, `documents/`, `store/` ve `fragments/demo/` dizinleri.

Bu dosyalardan birkaçını değiştirmeden önce bazı şeyleri bilmekte fayda var.

**`main.go` CLI ile bir sözleşmeye uyar.** `collage dev` programı, environment'ında
`COLLAGE_DEV=1` ve dinleyeceği `HOST` ile `PORT` tanımlı olarak çalıştırır.
`collage export` ise programı `-collage-build -out <dir>` ile çalıştırır. Scaffold
edilen `main.go` bunların hepsine uyar. Değişken development mode'u açar, sunucu
kendisine söylenen adreste dinler, flag ise siteyi serve etmek yerine dosyalara
render eder. Flag'lerden sonra gelen bir kelime, yani `go run . <command>`, bir
[plugin'in komutunu](/docs/plugins) çalıştırır. `main.go`'yu yeniden yazarsanız
bunların hepsinin çalışmaya devam etmesini sağlayın. Aksi hâlde bu komutlar işe yarar
hiçbir şey yapmaz. Ayrıntılar [CLI referansında](/docs/cli) yer alır.

**`newApp`, `main`'den ayrıdır.** `newApp` bütün uygulamayı kurar: config'i,
`routes.go`'daki route'ları ve `/static/` mount'unu. Sonra bir sunucu başlatmadan
uygulamayı döner. Demo projesindeki `main_test.go` aynı fonksiyonu çağırır ve
`app.Handler()`'ı `net/http/httptest` ile çalıştırır. Böylece testler sitenin ayrıca
kurulmuş ikinci bir kopyasını değil, gerçekten çalışan siteyi test eder.
[Test yazmak](/docs/testing) sayfasına bakın.

**Template'ler ve static dosyalar embed edilir.** `//go:embed all:templates` ve
`//go:embed all:static` bunları binary'nin içine gömer. Bu sayede binary herhangi
bir çalışma dizininden çalışabilir. Development'ta diskteki dizinler varsa her
zaman onlar kullanılır. Böylece düzenlediğiniz bir template bir sonraki request'te
yine görünür.

**Development'ta static dosyalar `os.DirFS` ile değil, `os.OpenRoot` ile mount
edilir.** Bir `os.Root`, dizinin dışına çıkan bir symlink'i reddeder. `os.DirFS`
ise onu takip eder. Production'da embed edilmiş kopya serve edilir. Bu kopya
`fs.Sub` üzerinden sunulur, böylece `static/` dizini URL'de ikinci bir path segment
olarak görünmez. [Static asset'ler](/docs/assets) sayfasına bakın.

**Production'da render edilen page'ler diskte, `.cache/` altında cache'lenir.**
Development'ta page cache hiç okunmaz. Böylece yaptığınız bir değişiklik hiçbir
zaman stale bir page'in arkasında kalmaz. [Caching](/docs/caching) sayfasına bakın.

## Environment dosyası

Scaffold, bir `.env.example` dosyasıyla gelir:

```sh
COLLAGE_CSRF_KEY=
PORT=3000
HOST=localhost
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

**Bu dosyaları yalnızca `collage dev` okur.** `collage build`, `collage export` ve
build edilmiş binary bu dosyaları hiçbir zaman okumaz. Production'da environment,
binary'nin çalıştığı ortamdan gelir.

`COLLAGE_CSRF_KEY`, form'ların taşıdığı token'ları imzalar. Development sırasında
boş bırakabilirsiniz. Bu durumda her process için yeni bir key üretilir. Form
içeren herhangi bir şeyi deploy etmeden önce bir key tanımlayın. Aksi hâlde restart
öncesinde gönderilmiş her form, restart'tan sonra reddedilir. İçinde form olan
cache'lenmiş bir page de key'e bağlıdır. Yeni bir key ile restart ettikten sonra o
page, içinde eski key'in token'ı ile serve edilmez, baştan render edilir.
[Caching](/docs/caching#the-namespace) sayfasına bakın. Bir key'i şöyle
üretebilirsiniz:

```sh
openssl rand -hex 32
```

## Çalıştırın

```sh
collage dev
```

Site [http://localhost:3000](http://localhost:3000) adresinde açılır. Bu adres
`collage dev`'in kendisidir: `HOST` ve `PORT`'u programınızın okuyacağı şekilde
okuyup orada dinler. Her request'i, loopback bir adreste çalıştırdığı programa
iletir. Program başlarken gelen bir request, programı bekler. `collage dev`
çalışırken dört şey olur.

### Go değişikliklerinde yeniden build alınır

`collage dev` projeyi `go build` ile build eder ve binary'yi çalıştırır. Programın
parçası olan dosyaları izler: `.go` dosyalarını (testler hariç), `go.mod`, `go.sum`
ve environment dosyasını. Bunlardan biri değiştiğinde projeyi yeniden build eder.

Önce yeni build alınır. Eski process ancak yeni build başarıyla derlendikten sonra
düzgünce durdurulur ve yenisi başlatılır. Derlenmeyen bir değişiklikte son sağlam
build serve etmeye devam eder ve compiler'ın hatası terminale yazılır. Böylece bir
yazım hatası yüzünden `localhost:3000` hiçbir zaman boş kalmaz.

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
kurmanız gereken bir şey yoktur. Programınızın diskten kendisinin okuduğu içerik,
örneğin Markdown ya da JSON, ne template ne de static dosyadır. Bu içerikteki
değişikliklerin de page'i yenilemesini istiyorsanız dizinini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtin (v0.10.0'dan
itibaren). Script hiçbir zaman production'daki bir page'e eklenmez. Bir form
gönderiminin response'una da eklenmez, çünkü sayfayı yenilemek formu yeniden
gönderirdi.

### Hatalar page'in üzerinde görünür

Development'ta hata veren bir fragment sessizce ortadan kaybolmaz. Page, üzerinde
bir panel ile serve edilir. Bu panel fragment'in adını ve hatasını gösterir,
template hatalarında dosya ve satırı da belirtir. Hatayı bir fallback kapatmış olsa
bile panel yine görünür. Page'in tamamı hata verdiğinde, built-in error page önce
nedeni gösterir: başarısız olan çağrının template'ini, satırını, sütununu ve
çağrının ne döndüğünü. Ardından hatanın başladığı fragment'i söyler ve bir
panic'in stack'i dahil bütün error zincirini gösterir. Kendi yazdığınız bir error
page'in üstünde de aynı panel çıkar ve hangi page'in yerine gösterildiğini
belirtir.

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

## Sırada ne var

[İlk page'iniz](/docs/your-first-page), minimal bir projede sıfırdan, veri kullanan
bir page oluşturur. Yayına almaya hazır olduğunuzda `collage build` bir binary
üretir, `collage export` ise static dosyalar yazar. [Deployment](/docs/deployment)
ve [Static export](/docs/static-export) sayfalarına bakın.
