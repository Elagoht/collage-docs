---
description: Go'yu ve collage CLI'ını kurun, collage new ile bir proje iskeleti oluşturun ve collage dev altında çalıştırın.
---

# Kurulum

Bir collage projesi sıradan bir Go modülüdür. `collage` komut satırı aracı bu
modülün iskeletini oluşturur, siz üzerinde çalışırken onu çalıştırır ve yayına
alacağınız zaman derler; framework'ün kendisi ise projenin import ettiği bir
kütüphanedir. İkisi de Go dışında hiçbir şeye ihtiyaç duymaz.

## Go'yu kurun

collage **Go 1.26 veya daha yenisini** gerektirir. Go'yu [go.dev/dl](https://go.dev/dl)
adresinden ya da paket yöneticinizden kurun ve kontrol edin:

```sh
go version
```

## CLI'ı kurun

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

`go install` binary'yi `$(go env GOBIN)` dizinine, `GOBIN` ayarlı değilse
`$(go env GOPATH)/bin` dizinine koyar; bu dizin `PATH`'inizde olmalıdır. Olduğunu
kontrol edin:

```sh
collage version
```

Yazdırdığı sürüm build'in kendisinden okunur; yani `go install` hangi sürümü
çektiyse odur. `collage help` her komutu listeler, `collage help new` ise tek bir
komutu açıklar.

## Proje oluşturun

```sh
collage new mysite
```

Bu komut `./mysite` içine, modül yolu `mysite` olan, çalıştırılabilir bir proje
yazar ve sonraki adımları yazdırır:

```text
Scaffolded "mysite" in mysite

Next steps:
  cd mysite
  go mod tidy
  cp .env.example .env.development
  collage dev
```

`go mod tidy` framework'ü çeker. İskeletteki `go.mod` yalnızca modülü ve Go
sürümünü belirtir; dolayısıyla projenin hangi collage sürümüyle derlendiğini
kaydeden şey ilk tidy'dir.

Birkaç flag, projenin nereye ve nasıl yazılacağını değiştirir. Addan önce de sonra
da gelebilirler:

| Flag | Etkisi |
| --- | --- |
| `-minimal` | Tek bir layout, boş bir ana sayfa ve bir bulunamadı sayfası — demo yok |
| `-dir path` | İskeleti `./<name>` yerine `path` içine oluşturur |
| `-module path` | `go.mod`'daki modül yolu, örneğin `github.com/you/mysite`. Varsayılanı addır |
| `-force` | İskeleti boş olmayan bir dizine oluşturur |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite -minimal
collage new mysite -dir . -force
```

### Demolarla ya da demosuz

`-minimal` olmadan bir ana sayfa ve canlı demolardan oluşan bir `/features`
sayfası elde edersiniz: önbellekteki bir sayfayı etiketle geçersiz kılan bir API
action'ına istek gönderen bir düğme, kendi sayfasına gönderilen bir HTML formu,
kendi URL'sinde açılan bir saat fragment'i ve `/healthz` adresinde bir JSON
document'ı. Bunların her birinin çalıştığını görmenin en hızlı yolu budur; her
birinin arkasındaki kod da bir oturuşta okunacak kadar kısadır.

`-minimal` ile aynı `main.go`'yu, aynı dizin düzenini ve aynı türden testleri elde
edersiniz; içlerinde silinecek hiçbir şey olmadan. Gerçek bir siteye başlarken bunu
kullanın. [İlk sayfanız](/docs/your-first-page) minimal bir projeden başlar.

## İskeletin içinde ne var

Minimal bir proje şöyle görünür:

```text
mysite/
├── main.go                     configuration, the static mount, the CLI contract, plugin commands
├── routes.go                   every page, document and action — a new route goes here
├── main_test.go                tests that drive app.Handler() with no server
├── go.mod
├── .env.example                variables for collage dev, to copy
├── .gitignore
├── plugins-config.json         plugin settings, keyed by plugin name
├── README.md
├── pages/
│   ├── home.go                 the home page: layout, content, path
│   └── not-found.go            the site-wide 404 page
├── fragments/
│   └── layouts/main.go         the layout fragment every page shares, and the site's title
├── templates/
│   ├── layouts/default.html    the layout's HTML, with {{slot "content"}}
│   └── pages/
│       ├── home.html
│       └── 404.html
└── static/
    ├── app.css
    └── favicon.svg
```

Demo projesi buna `actions/`, `documents/`, `store/` ve `fragments/demo/`
dizinlerini, ayrıca bunların şablonlarını ve script'lerini ekler.

İçindeki birkaç şeyi değiştirmeden önce bilmeye değer.

**`main.go` CLI ile bir sözleşmeyi korur.** `collage dev` programı ortamında
`COLLAGE_DEV=1` ile, `collage export` ise `-collage-build -out <dir>` ile
çalıştırır. İskeletteki `main.go` ikisini de okur: değişken geliştirme modunu açar,
flag ise siteyi sunmak yerine dosyalara render eder. Flag'lerden sonra gelen bir
sözcük — `go run . <command>` — bir [plugin'in komutunu](/docs/plugins) çalıştırır.
`main.go`'yu yeniden yazarsanız bunların hepsini çalışır hâlde tutun; aksi hâlde bu
komutlar işe yarar hiçbir şey yapmaz. Ayrıntılar
[CLI başvurusunda](/docs/cli).

**`newApp`, `main`'den ayrıdır.** Uygulamanın tamamını — yapılandırmayı,
`routes.go`'daki route'ları, `/static/` mount'unu — kurar ve bir sunucu başlatmadan
döndürür. `main_test.go` aynı fonksiyonu çağırır ve `app.Handler()`'ı
`net/http/httptest` ile sürer; böylece testler sitenin ikinci bir kablolamasını
değil, gerçekten çalışan siteyi sınar. Bkz. [Test](/docs/testing).

**Şablonlar ve statik dosyalar gömülüdür.** `//go:embed all:templates` ve
`//go:embed all:static` onları binary'nin içine koyar; böylece binary herhangi bir
çalışma dizininden çalışır. Geliştirmede diskteki dizinler, var oldukları her
durumda önceliklidir; böylece düzenlediğiniz bir şablon bir sonraki istekte yine
görünür.

**Geliştirmede statik dosyalar `os.DirFS` ile değil, `os.OpenRoot` ile mount
edilir.** Bir `os.Root`, dizinin dışına çıkan bir sembolik bağlantıyı reddeder;
`os.DirFS` ise onu izler. Production'da gömülü kopya sunulur; `static/` dizini
ikinci bir yol parçası olmasın diye `fs.Sub` üzerinden. Bkz.
[Statik dosyalar](/docs/assets).

**Production'da render edilen sayfalar disk üzerinde, `.cache/` altında önbelleğe
alınır.** Geliştirmede sayfa önbelleği hiç okunmaz; böylece bir düzenleme hiçbir
zaman bayat bir sayfanın arkasında gizli kalmaz. Bkz. [Önbellek](/docs/caching).

## Ortam dosyası

İskelet `.env.example` ile gelir:

```sh
COLLAGE_CSRF_KEY=
PORT=3000
HOST=localhost
```

Onu git tarafından yok sayılan `.env.development` dosyasına kopyalayın:

```sh
cp .env.example .env.development
```

`collage dev`, `.env.development` içindeki değişkenleri programın ortamına ekler —
`.env.development` yoksa `.env` içindekileri. Tek bir dosya okur, asla ikisini
birden değil. Kurallar kısadır:

- Kabuğunuzda zaten ayarlı olan bir değişken önceliklidir; bu yüzden
  `PORT=4000 collage dev` çalışır. Dosyada ne yazarsa yazsın `COLLAGE_DEV=1` her
  zaman ayarlanır.
- Dosya `KEY=value` satırlarından, `#` yorumlarından ve boş satırlardan oluşur.
  Başta bir `export ` öneki ve değerin etrafında tırnaklar kullanılabilir.
- Hatalı bir satır atlanmaz; dosya adı ve satır numarasıyla bildirilir: atlanan bir
  satır, sizin yazdığınız ama programın hiç görmediği bir ayardır. Siz düzeltene
  kadar hiçbir şey başlatılmaz ya da yeniden başlatılmaz — zaten çalışan bir build
  sunmaya devam eder — ve `collage dev` izlemeyi sürdürür; düzeltmeyi kaydettiğinizde
  kaldığı yerden devam eder.
- Hiç dosya olmaması bir hata değildir. Bir dosya varsa adı stderr'e yazdırılır.

**Bu dosyaları yalnızca `collage dev` okur.** `collage build`, `collage export` ve
derlenmiş binary asla okumaz; production'da ortam, binary nerede çalışıyorsa
oradan gelir.

`COLLAGE_CSRF_KEY`, formların taşıdığı token'ları imzalar. Geliştirirken boş
bırakmanız sorun değildir — her süreç için bir anahtar üretilir. Form içeren
herhangi bir şeyi yayına almadan önce bir anahtar ayarlayın; aksi hâlde yeniden
başlatmadan önce gönderilen her form, yeniden başlatmadan sonra reddedilir. İçinde
form olan önbellekteki bir sayfa da anahtara bağlıdır: yeni bir anahtarla yeniden
başlatmadan sonra o sayfa, içinde eski anahtarın token'ı ile sunulmak yerine
baştan render edilir — bkz. [Önbellek](/docs/caching#the-namespace). Bir anahtarı
şöyle üretin:

```sh
openssl rand -hex 32
```

## Çalıştırın

```sh
collage dev
```

Site [http://localhost:3000](http://localhost:3000) adresindedir. Çalışırken üç şey
olur.

### Go değişiklikleri yeniden derlenir

`collage dev` projeyi `go build` ile derler ve binary'yi çalıştırır. Programın
neyden oluştuğunu — `.go` dosyalarını (testler hariç), `go.mod`, `go.sum` ve ortam
dosyasını — izler ve bir değişiklikte yeniden derler.

Önce yeni build yapılır. Eski süreç ancak yeni build derlendikten sonra, düzgün bir
şekilde durdurulur ve yenisi başlatılır. Derlenmeyen bir değişiklik son sağlam
build'i sunmaya devam ettirir ve derleyicinin hatasını terminale yazdırır; böylece
bir yazım hatası sizi `localhost:3000`'de hiçbir şey olmadan bırakmaz.

Art arda yapılan kayıtlar tek bir yeniden derlemedir. Gizli dizinler, `bin`,
`dist`, `node_modules`, `testdata` ve `vendor` hiçbir zaman izlenmez; böylece
çalışan programın yazdığı hiçbir şey kendi yeniden derlenmesini tetikleyemez.
Kendiliğinden çıkan bir program da — başlangıçta bir panic, zaten kullanımda olan
bir port — bir döngü içinde yeniden başlatılmaz; bir sonraki değişikliğiniz onu
yeniden başlatır.

### Şablonlar ve statik dosyalar diskten okunur

Geliştirme modunda her şablon, her render'dan önce diskten yeniden ayrıştırılır ve
statik dosyalar diskteki `static/` dizininden sunulur. İkisini düzenlemek de
yeniden derleme gerektirmez ve yeniden derleme olmaz.

### Tarayıcı kendini yeniler

Geliştirmede sunulan her sayfa, bir şablon ya da statik dosya değiştiğinde ve
program bir yeniden derlemeden geri döndüğünde sayfayı yenileyen küçük bir script
taşır. Bir dosyayı kaydedin ve tarayıcıya bakın; kurulacak bir şey yok.
Programınızın diskten kendisinin okuduğu içerik — Markdown, JSON — ne şablondur ne
de statik dosya; değişikliklerinin de sayfayı yenilemesi için dizinini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtin (v0.10.0'dan
itibaren). Script hiçbir zaman bir production sayfasına eklenmez; bir form
gönderiminin yanıtına da eklenmez, çünkü yenilemek formu yeniden gönderirdi.

### Hatalar sayfada görünür

Geliştirmede hata veren bir fragment sessizce kaybolmaz. Sayfa, üzerinde fragment'i
ve hatasını — şablon için dosya ve satırla birlikte — belirten bir panelle sunulur;
bir yedek onun yerini tutmuş olsa bile. Sayfanın tamamı hata verdiğinde yerleşik
hata sayfası, hatanın başladığı fragment'i adlandırır ve bir panic'in yığınıyla
birlikte hata zincirinin tamamını gösterir — sizin kendi hata sayfanızın üstüne de
aynı panel, neyin yerine durduğunu söyleyerek eklenir.

Bunların hiçbiri geliştirme dışında yoktur. Production'daki bir hata sayfası tek bir
genel cümle söyler; çünkü hata mesajları host adları, dosya yolları ve kimlik
bilgileri taşır ve bir hata sayfası, bunları bir yabancıya teslim etme ihtimali en
yüksek yanıttır. Bkz. [Hatalar](/docs/errors).

## Sırada ne var

[İlk sayfanız](/docs/your-first-page), minimal bir projede sıfırdan veriyle bir
sayfa kurar. Yayına almaya hazır olduğunuzda `collage build` bir binary üretir,
`collage export` ise statik dosyalar yazar — bkz. [Yayına alma](/docs/deployment) ve
[Statik dışa aktarma](/docs/static-export).
