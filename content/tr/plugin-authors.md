---
description: collage plugin add'in repository'yi kurması için gerekenler (ad, manifest, ayar istemeyen bir New, release tag'i) ve collage plugin check'in bunu doğrulaması.
reference: Plugin, Host, FindPlugin, Configurer, PluginConfig
---

# Plugin yazmak: repository sözleşmesi

[Plugin yazmak](/docs/writing-plugins) Go tarafını anlatır: `Plugin` interface'i ve
hook'ları. Bu sayfa ise onu çevreleyen repository ile ilgilidir.

Aşağıdaki kurallara uyan bir plugin'i herkes tek komutla kurabilir, bir başka
komutla da kaldırabilir:

```sh
collage plugin add you/stamp      # github.com/you/collage-stamp
collage plugin del you/stamp
```

`collage plugin add`, bir kuralı çiğneyen modülü reddeder, hangi kuralın
çiğnendiğini söyler ve bu sayfaya yönlendirir.
[`collage plugin check`](#check-it-before-you-tag) aynı kuralları, başkası
çalıştırmadan önce sizin repository'nizde çalıştırır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 520" role="img" aria-labelledby="pa-contract-title">
<title id="pa-contract-title">Bir plugin'in repository'sinden kurulu bir plugin'e. github.com/you/collage-stamp repository'sinde module path'i yazan go.mod, plugin'e you/stamp adını veren collage.json, New'u Options alan kök package stamp ve HEAD'de bir v0.1.0 release tag'i bulunur. collage plugin check collage.json'ı, kök package'ı ve New'un biçimini okur; -release ile HEAD'de bir release tag'i ister, -run ile New'u çağırır ve Name()'i manifest'teki adla karşılaştırır. Ardından bir kullanıcının projesinde collage plugin add you/stamp modülü go get ile çeker, aynı kuralları kontrol eder, you/stamp adını ve pseudo-version olmayan bir sürüm ister ve gerektirdiği plugin'leri önce ekler. plugins.go'ya order'ın gösterdiği yere stamp.New(stamp.Options{}) çağrısını, plugins-config.json'a boş bir you/stamp bölümünü yazar ve varsa setup notunu yazdırır.</title>
<defs><marker id="pa-contract-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="8" width="312" height="144" rx="8" />
<text x="14" y="25">plugin'in repository'si</text>
<rect class="d-folder" x="14" y="34" width="292" height="24" rx="4" /><text class="d-small" x="160" y="50" text-anchor="middle">go.mod · module github.com/you/collage-stamp</text>
<rect class="d-folder" x="14" y="62" width="292" height="24" rx="4" /><text class="d-small" x="160" y="78" text-anchor="middle">collage.json · "name": "you/stamp"</text>
<rect class="d-folder" x="14" y="90" width="292" height="24" rx="4" /><text class="d-small" x="160" y="106" text-anchor="middle">stamp.go · package stamp · func New(Options)</text>
<rect class="d-folder" x="14" y="118" width="292" height="24" rx="4" /><text class="d-small" x="160" y="134" text-anchor="middle">HEAD'de git tag v0.1.0, push edilmiş</text>
<line class="d-flow" x1="160" y1="152" x2="160" y2="166" marker-end="url(#pa-contract-arrow)" />
<rect class="d-cmd" x="46" y="170" width="228" height="24" rx="12" /><text class="d-cmd-text" x="160" y="186" text-anchor="middle">collage plugin check -release -run</text>
<line class="d-flow" x1="160" y1="194" x2="160" y2="206" marker-end="url(#pa-contract-arrow)" />
<rect class="d-soft-a" x="14" y="210" width="292" height="58" rx="6" /><text class="d-small" x="160" y="226" text-anchor="middle">collage.json, kök package, New'un biçimi</text><text class="d-small" x="160" y="241" text-anchor="middle">-release: HEAD'de bir release tag'i (git)</text><text class="d-small" x="160" y="256" text-anchor="middle">-run: New'u çağırır; Name() = name olmalı</text>
<line class="d-rule" x1="6" y1="284" x2="314" y2="284" />
<text class="d-small" x="160" y="300" text-anchor="middle">sonra, bir kullanıcının projesinde</text>
<rect class="d-cmd" x="60" y="310" width="200" height="24" rx="12" /><text class="d-cmd-text" x="160" y="326" text-anchor="middle">collage plugin add you/stamp</text>
<line class="d-flow" x1="160" y1="334" x2="160" y2="346" marker-end="url(#pa-contract-arrow)" />
<rect class="d-soft-a" x="14" y="350" width="292" height="58" rx="6" /><text class="d-small" x="160" y="366" text-anchor="middle">go get @upgrade, sonra aynı kurallar</text><text class="d-small" x="160" y="381" text-anchor="middle">ad you/stamp olmalı; pseudo-version olmaz</text><text class="d-small" x="160" y="396" text-anchor="middle">gerektirdiği plugin'ler önce eklenir</text>
<line class="d-flow" x1="160" y1="408" x2="160" y2="420" marker-end="url(#pa-contract-arrow)" />
<rect class="d-folder" x="14" y="424" width="292" height="24" rx="4" /><text class="d-small" x="160" y="440" text-anchor="middle">plugins.go · stamp.New(stamp.Options{}) (order)</text>
<rect class="d-folder" x="14" y="452" width="292" height="24" rx="4" /><text class="d-small" x="160" y="468" text-anchor="middle">plugins-config.json · "you/stamp": {}</text>
<line class="d-flow" x1="160" y1="476" x2="160" y2="488" marker-end="url(#pa-contract-arrow)" />
<rect class="d-pill" x="60" y="492" width="200" height="24" rx="12" /><text class="d-small" x="160" y="508" text-anchor="middle">varsa setup notu yazdırılır</text>
</svg>
<figcaption>Bir repository'den kurulu bir plugin'e. Repository'de module path'i <code>github.com/you/collage-stamp</code> olan bir <code>go.mod</code>, plugin'e <code>you/stamp</code> adını veren bir <code>collage.json</code>, ayar gerektirmeyen bir <code>New</code> içeren kök package ve bir release tag'i bulunur. <code>collage plugin check</code> manifest'i, package'ı ve <code>New</code>'un biçimini okur; <code>-release</code> ayrıca <code>HEAD</code>'de bir release tag'i ister, <code>-run</code> ise <code>New</code>'u çağırıp <code>Name()</code> değerini manifest'le karşılaştırır. Bir kullanıcının <code>collage plugin add you/stamp</code> komutu modülü çeker, aynı kuralları uygular, manifest'teki adın yazılanla aynı, sürümün de bir release olmasını ister, gerektirdiği plugin'leri önce ekler, sonra çağrıyı <code>plugins.go</code>'da <code>order</code>'ın gösterdiği yere, boş bir bölümü de <code>plugins-config.json</code>'a yazar.</figcaption>
</figure>

## Kurallar

### Repository'nin adı collage-name olur

`you/stamp` gibi kısa bir referans `github.com/you/collage-stamp` demektir.
Repository'ye `collage-<name>` adını verirseniz bu şekilde eklenebilir.

Başka bir yerde barındırılan ya da farklı adlandırılmış bir plugin module path'iyle
eklenir (`collage plugin add gitlab.com/you/stamper`). Diğer kurallar yine
geçerlidir.

### Kökte bir collage.json bulunur

Manifest, modülün kök dizininde durur. İçindeki `name` `owner/name` biçimindedir ve
her kısım küçük harf, rakam ve tireden oluşur:

```json
{
  "name": "you/stamp",
  "description": "Adds a build stamp to every page.",
  "repository": "https://github.com/you/collage-stamp"
}
```

Bu ad aynı zamanda plugin'inizin `Name()` metodunun döndürdüğü değer ve
`plugins-config.json`'daki bölümünün key'idir. Kısa bir referansta manifest'teki
ad, yazılanla aynı olmalıdır: `collage plugin add you/stamp` için
`"name": "you/stamp"` gerekir.

Manifest katı biçimde okunur: şemanın tanımadığı bir key reddedilir. Şema,
[editör extension'ı](/docs/installation#editor-support) ile birlikte gelir. Extension
key'leri tamamlar ve siz yazarken dosyayı kontrol eder.

### Plugin modülün köküdür

`New`'u içeren package, modülün kök package'ıdır; yani import path'i module path'in
kendisidir. `package main` olamaz, çünkü onu hiçbir şey import edemez. Kodunuzun
geri kalanı için alt package'lar kullanabilirsiniz.

### New ayar gerektirmez

`add`, seçeneklerinizin ne olduğunu bilmeden kullanıcının `plugins.go` dosyasına
`New` çağrısını yazar. Bu yüzden `New`, hiçbir şey ayarlanmadan çağrılabilen bir
fonksiyon olmalıdır. Kabul edilen biçimler:

| `New`'un aldığı | `add`'in yazdığı |
| --- | --- |
| hiçbir şey | `stamp.New()` |
| yalnızca variadic bir parametre | `stamp.New()` |
| package'ınızın tek bir struct tipi | `stamp.New(stamp.Options{})` |
| tek bir pointer, interface, slice, map ya da func | `stamp.New(nil)` |

`New` tam olarak bir değer döndürür: bir `collage.Plugin`. Reddedilenler:

- iki ya da daha fazla parametre;
- `string` gibi temel bir tip;
- başka bir package'ın struct tipi (sıfır değerini kullanıcının dosyası her zaman
  yazamaz);
- başka hiçbir package'ın adlandıramadığı, export edilmemiş bir tip.

Çağrı başarılı olmalıdır. Sıfır değerin ne anlama geldiğine siz karar verirsiniz:
varsayılanlarıyla çalışan bir plugin olmalıdır. Makul bir varsayılanı olmayan bir
ayar ise ayrı bir durumdur ve [aşağıda](#a-required-setting-fails-in-init)
anlatılır.

### Bir release tag'i vardır

`add`'in çözdüğü sürüm, `v0.1.0` gibi bir semver tag'idir. Hiç tag'i olmayan bir
repository'nin yalnızca pseudo-version'ı (`v0.0.0-2026…-abcdef`) vardır ve bu
reddedilir: sabitlenecek kararlı bir şey olmaz.

## İsteğe bağlı manifest alanları

Üç alan, plugin'inizin nereye ve nasıl yerleşeceğini `add`'e anlatır.

**`requires`**, projede ayrıca bulunması gereken plugin adlarının listesidir. `add`
eksik olanların her birini önce ekler. `del` ise bir plugin'i, başka bir plugin
onu gerektirdiği sürece kaldırmayı reddeder. Websocket plugin'i
`["elagoht/live"]` gerektirir.

**`order`**, girdinin `plugins.go`'da nereye gideceğini söyleyen bir tam sayıdır
(yoksa 0). Küçük olan daha erken gelir, `Config.Plugins`'te daha erken olan da daha
dıştaki middleware'dir.

`add`, yeni girdiyi mevcut girdiler arasında order'ı daha büyük olan ilkinin önüne
koyar ve zaten var olan hiçbir girdiyi yerinden oynatmaz. Kendine yer isteyen
yayımlanmış plugin'ler:

| Plugin | Order | Neden |
| --- | --- | --- |
| elagoht/health | −30 | bir request'i reddeden ya da cevaplayan her şeyden önce |
| elagoht/fail2ban | −25 | health'in hemen ardından |
| elagoht/compress | −20 | response body'lerini yeniden yazan her plugin'den önce |
| elagoht/websocket | −1 | elagoht/live'dan önce configure edilir |
| elagoht/devtoolbar | 100 | en sonda, böylece diğer plugin'lerin finding'lerini görür |

Diğer bütün plugin'lerin order'ı 0'dır. Konum önemli değilse alanı yazmayın.

**`setup`**, `add` başarılı olduktan sonra yazdırılan kısa bir nottur. Yalnızca Go
kodunun verebileceği şeyleri anlatır. Feed plugin'i, feed'lerin `plugins.go`'daki
`feed.New(...)`'a verildiğini söyler.

```json
{
  "name": "you/stamp",
  "requires": ["elagoht/live"],
  "order": -5,
  "setup": "Pass your stamps to stamp.New(...) in plugins.go."
}
```

## Zorunlu bir ayar Init'te başarısız olur

`add`, `plugins-config.json`'a `"you/stamp": {}` bölümünü yazar: yani varsayılanları.
Plugin'iniz bir ayar olmadan çalışamıyorsa (bir API anahtarı, bir site URL'si gibi)
boş bölümde o ayar eksiktir.

Bu bir sözleşme ihlali değildir, çünkü `New` başarılı oldu. `Init`'te, kullanıcının
ne yapacağını anlayacağı bir mesajla reddedin:

```go
func (p *Plugin) Init(ctx context.Context, host collage.Host) error {
	if p.cfg.APIKey == "" {
		return errors.New(`you/stamp: "apiKey" is required; set it in the "you/stamp" section of plugins-config.json`)
	}
	return nil
}
```

Kullanıcı plugin'i ekler, uygulamayı başlatır, bu mesajı okur ve bölümü doldurur.
Sizden başka bir şey gerekmez.

## Başka bir plugin'e ihtiyaç duyan plugin

Plugin'iniz başka bir plugin üzerinden çalışıyorsa, onu kullanıcıdan istemek yerine
tipiyle bulun. `collage.FindPlugin[T](host)`, uygulamanın plugin'leri arasında `T`
tipindeki plugin'i döndürür (birden fazlaysa ilkini). Onu `Configure`'dan çağırın ve
bulduğunu saklayın: v0.78.0'dan beri `Init`'e verilen `Host` plugin'leri listelemez,
bu yüzden `Init`'ten çağrılan `FindPlugin` false döner.

```go
live, ok := collage.FindPlugin[*live.Plugin](host)
if !ok {
	return errors.New("you/stamp needs elagoht/live in the plugin list")
}
```

`FindPlugin` listedeki her plugin'i, o plugin henüz configure edilmiş olsun
olmasın görür. Sonra diğer plugin'in adını `requires`'a yazın ki
`add` onu da getirsin. Websocket plugin'i tam olarak bunu yapar; `New(nil)`'in onun
için çalışmasının nedeni de budur.

## Tag koymadan önce kontrol edin

Repository'nizin içinde çalıştırın:

```sh
collage plugin check              # manifest ve New
collage plugin check -release     # ayrıca: HEAD'de bir release tag'i (git ister)
collage plugin check -run         # ayrıca: çağrıyı derle ve çalıştır
collage plugin check -release -run ./path/to/repo
```

Flag verilmezse repository'yi okur ve manifest'i, adı, package'ı ve `New`'un
biçimini kontrol eder. `-release`, `HEAD`'de bir release tag'i ister.

`-run` bir adım daha ileri gider:

- `add`'in yazacağı çağrının aynısını yapan küçük, geçici bir program derler, o
  çağrıyı çalıştırır ve plugin'in `Name()` değerini manifest'in `name`'i ile
  karşılaştırır.
- Build hatasında, `New`'daki bir panic'te ya da farklı bir adda başarısız olur.
- Bir uygulamayı başlatmaz, `Configure`'u ya da `Init`'i çağırmaz, hiçbir config
  okumaz. Bu yüzden boş bir bölümde eksik kalan bir ayarı raporlamaz.
- Programın dosyaları sonra silinir, `go.mod` ve `go.sum` eski hâline döner.

`check` ile `add` kuralları paylaşır ama her adımı değil. `check`'in adı
karşılaştıracağı kısa bir referansı yoktur. Ayrıca `add` çektiği sürüme bakarken
`check` sizin `HEAD`'inizdeki tag'lere bakar:

| Kural | `collage plugin check` | `collage plugin add` |
| --- | --- | --- |
| `collage.json` ve `name`'i | evet | evet; kısa bir referansta ad, yazılanla aynı olmalıdır |
| Kök package ve `New`'un biçimi | evet | evet |
| `New` çalışır ve `Name()` eşleşir | `-run` ile | hayır |
| Bir release | `-release` ile: `HEAD`'de bir semver tag'i | çekilen sürüm pseudo-version değildir |

Yayımlanmış plugin'ler her release'te `collage plugin check -run` çalıştırır. Siz de
CI'ınızda aynısını yapın. Geçen bir plugin'i herkes tek satırla ekleyebilir.

## Örnek: okuma süresi plugin'i yayımlamak

Bir blog, her yazının başlığının altında "3 min read" göstermek istiyor. Bu bölüm o
özelliği `you/readtime` adlı bir plugin olarak ele alır ve boş bir repository'den
birinin `plugins.go` dosyasındaki bir satıra kadar götürür.

### Repository

Repository'nin adı `collage-readtime`'dır; bu sayede `you/readtime` kısa referansı
onu bulur. Her şey modülün kökünde durur:

```text
collage-readtime/
├── collage.json
├── go.mod          module github.com/you/collage-readtime
├── go.sum
└── readtime.go     package readtime
```

Manifest plugin'e adını verir. Bu ad, `Name()`'in döndürdüğü değer ve bölümünün
key'idir:

```json
{
  "name": "you/readtime",
  "description": "Adds a readTime template function: minutes to read a text."
}
```

Package, export edilmiş tek bir options struct'ı alır; bu yüzden `add`
`readtime.New(readtime.Options{})` yazacaktır. Sıfır `Options` çalışan bir
plugin'dir: 0 okuma hızı varsayılan demektir, yani dakikada 200 kelime.

```go
// Package readtime adds a readTime template function: the minutes it takes to
// read a text.
package readtime

import (
	"context"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"
)

// Options is the plugin's configuration. The zero value is the defaults.
type Options struct {
	// WordsPerMinute is the reading speed; 200 when zero.
	WordsPerMinute int `json:"wordsPerMinute"`
}

// Plugin is the readtime plugin. Construct it with New.
type Plugin struct{ opts Options }

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

var (
	_ collage.Plugin     = (*Plugin)(nil)
	_ collage.Configurer = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return "you/readtime" }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure reads the section and adds the template function.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts)
	if err != nil {
		return err
	}
	if opts.WordsPerMinute <= 0 {
		opts.WordsPerMinute = 200
	}
	p.opts = opts
	return host.AddTemplateFunc("readTime", func(text string) int {
		words := len(strings.Fields(text))
		return (words + p.opts.WordsPerMinute - 1) / p.opts.WordsPerMinute
	})
}

func (p *Plugin) Init(context.Context, collage.Host) error { return nil }
func (p *Plugin) Shutdown(context.Context) error           { return nil }
```

### Kontrol, tag, yeniden kontrol

Dosyalar commit edildikten sonra sözleşmeyi kontrol edin:

```text
$ collage plugin check
you/readtime: meets the contract
```

`New` okuma hızını bir `int` olarak alsaydı, kontrol onu kuralı ve bu sayfanın
adresiyle birlikte reddederdi:

```text
$ collage plugin check
collage: plugin check: not a collage plugin: New takes a int; it may take none, or one parameter whose zero value works (a struct, pointer, interface, slice, map or func); see https://collage.furkanbaytekin.dev/en/docs/plugin-authors/
```

`-release`, bulunduğunuz commit'te bir tag ister; bu yüzden tag koyana kadar
başarısız olur:

```text
$ collage plugin check -release
you/readtime: meets the contract
collage: plugin check: no release tag at HEAD; add rule 5 with git tag vX.Y.Z
```

Commit'e tag koyun, sonra `New` da dahil bütün kontrolleri çalıştırın:

```text
$ git tag v0.1.0
$ collage plugin check -release -run
you/readtime: meets the contract
you/readtime: v0.1.0 at HEAD is a release
you/readtime: New runs, and Name() matches collage.json
```

Tag'i `git push origin v0.1.0` ile push edin. Push edilene kadar `go get` release'i
göremez ve `add`'in çekeceği bir şey olmaz.

### Bir bloga kurmak

Blogun projesinde tek bir komut modülü çeker, yeniden kontrol eder ve iki dosyayı
düzenler:

```text
$ collage plugin add you/readtime
added you/readtime v0.1.0
  plugins.go           readtime.New(readtime.Options{})
  plugins-config.json  "you/readtime": {}
```

`plugins.go`'ya import ve girdi eklenir:

```go
import (
	"github.com/Elagoht/collage/pkg/collage"

	readtime "github.com/you/collage-readtime"
)

func plugins() []collage.Plugin {
	return []collage.Plugin{
		readtime.New(readtime.Options{}),
	}
}
```

Boş bölüm varsayılanlar demektir. Okurları daha yavaş olan bir blog bölümü doldurur,
yazı template'i de fonksiyonu çağırır:

```json
{
  "you/readtime": { "wordsPerMinute": 150 }
}
```

```html
<p>{{readTime .Body}} min read</p>
```

Buradaki `.Body` yazının metnini temsil eder; yazının data handler'ı ona hangi adı
veriyorsa odur. `collage plugin del you/readtime` girdiyi, import'u ve bölümü
yeniden kaldırır.
