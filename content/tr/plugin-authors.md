---
description: collage plugin add'in repository'yi kurması için gerekenler (ad, manifest, ayar istemeyen bir New, release tag'i) ve collage plugin check'in bunu doğrulaması.
reference: Plugin, Host, FindPlugin, Configurer, PluginConfig
---

# Plugin yazmak: repository sözleşmesi

[Plugin yazmak](/docs/writing-plugins) Go tarafını anlatır: `Plugin` interface'i ve
hook'ları. Bu sayfa ise onu çevreleyen repository ile ilgilidir. Aşağıdaki
kurallara uyan bir plugin'i herkes tek komutla kurabilir, bir başka komutla da
kaldırabilir:

```sh
collage plugin add you/stamp      # github.com/you/collage-stamp
collage plugin del you/stamp
```

`collage plugin add`, bir kuralı çiğneyen modülü reddeder, hangi kuralın
çiğnendiğini söyler ve bu sayfaya yönlendirir.
[`collage plugin check`](#check-it-before-you-tag) aynı kuralları, başkası
çalıştırmadan önce sizin repository'nizde çalıştırır.

## Kurallar

### Repository'nin adı collage-name olur

`you/stamp` gibi kısa bir referans `github.com/you/collage-stamp` demektir.
Repository'ye `collage-<name>` adını verirseniz bu şekilde eklenebilir. Başka bir
yerde barındırılan ya da farklı adlandırılmış bir plugin module path'iyle eklenir
(`collage plugin add gitlab.com/you/stamper`). Diğer kurallar yine geçerlidir.

### Kökte bir collage.json bulunur

Manifest, modülün kök dizininde durur. İçindeki `name` `owner/name` biçimindedir ve
her parça küçük harf, rakam ve tireden oluşur:

```json
{
  "name": "you/stamp",
  "description": "Adds a build stamp to every page.",
  "repository": "https://github.com/you/collage-stamp"
}
```

Bu ad, plugin'inizin `Name()` değeri ve `plugins-config.json`'daki bölümünün
anahtarıdır. Kısa bir referansta manifest'teki ad, yazılanla aynı olmalıdır:
`collage plugin add you/stamp` için `"name": "you/stamp"` gerekir.

Manifest katı biçimde okunur: şemanın tanımadığı bir anahtar reddedilir. Şema,
[editör extension'ı](/docs/installation#editor-support) ile birlikte gelir. Extension
anahtarları tamamlar ve siz yazarken dosyayı kontrol eder.

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

Tam olarak bir değer döndürür: bir `collage.Plugin`. Reddedilenler: iki ya da daha
fazla parametre, `string` gibi temel bir tip, başka bir package'ın struct tipi
(sıfır değerini kullanıcının dosyası her zaman yazamaz) ve başka hiçbir package'ın
adlandıramadığı, export edilmemiş bir tip.

Çağrı başarılı olmalıdır. Sıfır değerin ne anlama geldiğine siz karar verirsiniz:
varsayılanlarıyla çalışan bir plugin olmalıdır. Makul bir varsayılanı olmayan bir
ayar ise ayrı bir durumdur, [aşağıda](#a-required-setting-fails-in-init)
anlatılıyor.

### Bir release tag'i vardır

`add`'in çözdüğü sürüm, `v0.1.0` gibi bir semver tag'idir. Hiç tag'i olmayan bir
repository'nin yalnızca pseudo-version'ı (`v0.0.0-2026…-abcdef`) vardır ve bu
reddedilir: sabitlenecek kararlı bir şey olmaz.

## İsteğe bağlı manifest alanları

Üç alan, plugin'inizin nereye ve nasıl yerleşeceğini `add`'e anlatır.

**`requires`**, projede ayrıca bulunması gereken plugin adlarının listesidir. `add`
eksik olanların her birini önce ekler. `del` ise başka bir plugin gerektirirken
onu kaldırmayı reddeder. Websocket plugin'i `["elagoht/live"]` ister.

**`order`**, girdinin `plugins.go`'da nereye gideceğini söyleyen bir tam sayıdır
(yoksa 0). Küçük olan daha erken gelir, `Config.Plugins`'te daha erken olan da daha
dıştaki middleware'dir. `add`, yeni girdiyi order'ı daha büyük olan ilk girdinin
önüne koyar ve zaten var olan hiçbir girdiyi yerinden oynatmaz. Kendine yer isteyen
yayımlanmış plugin'ler:

| Plugin | Order | Neden |
| --- | --- | --- |
| elagoht/health | −30 | bir request'i reddeden ya da cevaplayan her şeyden önce |
| elagoht/fail2ban | −25 | health'in hemen ardından |
| elagoht/compress | −20 | response body'lerini yeniden yazan her plugin'den önce |
| elagoht/websocket | −1 | elagoht/live'dan önce configure edilir |
| elagoht/devtoolbar | 100 | en sonda, diğer plugin'lerin finding'lerini görür |

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

## Zorunlu bir ayar Init'te reddedilir

`add`, `plugins-config.json`'a `"you/stamp": {}` bölümünü yazar: yani varsayılanları.
Plugin'iniz bir ayar olmadan çalışamıyorsa (bir API anahtarı, bir site URL'si gibi)
boş bölümde o ayar eksiktir. Bu bir sözleşme ihlali değildir, çünkü `New` başarılı
oldu. `Init`'te, kullanıcının ne yapacağını anlayacağı bir mesajla reddedin:

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
tipindeki plugin'i döndürür (birden fazlaysa ilkini). `Configure`'dan da `Init`'ten
de çağrılabilir:

```go
live, ok := collage.FindPlugin[*live.Plugin](host)
if !ok {
	return errors.New("you/stamp needs elagoht/live in the plugin list")
}
```

Listedeki her plugin'i, o plugin henüz configure ya da initialise edilmiş olsun
olmasın görür. Sonra diğer plugin'in adını `requires`'a yazın ki `add` onu da
getirsin. Websocket plugin'i tam olarak bunu yapar, `New(nil)` onun için bu yüzden
çalışır.

## Etiketlemeden önce kontrol edin

Repository'nizin içinde çalıştırın:

```sh
collage plugin check              # manifest ve New
collage plugin check -release     # ayrıca: HEAD'de bir release tag'i (git ister)
collage plugin check -run         # ayrıca: çağrıyı derle ve çalıştır
collage plugin check -release -run ./path/to/repo
```

Flag'siz hâli repository'yi okur ve manifest'i, adı, package'ı ve `New`'un biçimini
kontrol eder. `-release`, `HEAD`'de bir release tag'i ister.

`-run` bir adım daha ileri gider. `add`'in yazacağı çağrının aynısını yapan küçük,
geçici bir program derler, o çağrıyı çalıştırır ve plugin'in `Name()` değerini
manifest'in `name`'i ile karşılaştırır. Build hatasında, `New`'daki bir panic'te ya
da farklı bir adda başarısız olur. Bir uygulamayı başlatmaz, `Configure`'u ya da
`Init`'i çağırmaz, hiçbir config okumaz. Bu yüzden boş bir bölümde eksik kalan bir
ayarı raporlamaz. Programın dosyaları sonra silinir, `go.mod` ve `go.sum` eski
hâline döner.

Yayımlanmış plugin'ler her release'te `collage plugin check -run` çalıştırır. Siz de
CI'ınızda aynısını yapın. Geçen bir plugin'i herkes tek satırla ekleyebilir.
