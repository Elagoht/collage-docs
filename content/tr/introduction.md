---
description: collage nedir, hangi fikrin üzerine kurulu ve ne zaman doğru araçtır.
---

# Giriş

collage, sunucuda render edilen web siteleri için bir Go framework'üdür. Sayfaları,
fragment denen küçük parçaların etrafını saran bir layout olarak yazarsınız; her
fragment kendi verisini çeker ve kendi şablonunu render eder. collage sayfayı bir
araya getirir, önbelleğe alır ve içeriğiniz değiştiğinde neyi atması gerektiğini
tam olarak bilir.

Tek bir Go programıdır. Aynı binary sitenizi sunar, formlarınıza ve API
çağrılarınıza yanıt verir ve — isterseniz — sitenin tamamını statik dosyalar olarak
yazar. İstemci framework'ü yok, bundler yok, standart kütüphane dışında bağımlılık
yok.

## Fikir

Çoğu sayfa, birbirinden haberi olmayan parçalardan oluşur. Bir blog yazısında yazının
kendisi, bir yazar kartı, ilgili yazıların listesi ve bir gezinme çubuğu vardır. Her
parça farklı bir veriye ihtiyaç duyar, bu veriyi farklı bir yerden alır ve farklı
nedenlerle değişir.

collage'da bu parçaların her biri bir **fragment**'tir: bir şablon, o şablonun
ihtiyaç duyduğunu getiren bir fonksiyon ve başka fragment'lerin yerleştiği, adı
konmuş **slot**'lar. Bir **sayfa** ise içinde bir içerik fragment'i bulunan bir
layout fragment'i ve bir URL'dir.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(collage.DataHandler(loadAuthor)).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	WithSlot("author", true, false).
	WithSlotFragment("author", author).
	Build()

page := collage.NewPage("post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()
```

Sayfaları böyle kurmanın üç sonucu var.

- **Veri eşzamanlı çekilir, sayfa sırasıyla yazılır.** Bir fragment render
  edilmeden önce slot'larındaki her fragment'in data handler'ı aynı anda başlar.
  Birbirinden bağımsız parçalardan oluşan bir sayfa, dış servis çağrılarını birlikte
  yapar; HTML ise her seferinde aynı çıkar.
- **Sayfa neyden yapıldığını bilir.** Her data handler hangi içerik parçalarını
  kullandığını söyler — `post:hello-world`, `author:ada` — ve önbellekteki sayfa bu
  etiketleri taşır. Bir yazar değiştiğinde tek bir çağrı onu gösteren her sayfayı
  düşürür, başka hiçbir şeyi değil.
- **Bir parça, sayfayı da götürmeden hata verebilir.** Bir fragment zorunlu olabilir,
  bir yedeği olabilir ya da hiçbir şey render etmeyebilir. Bozuk bir kenar çubuğu,
  eksik bir kenar çubuğudur; 500 değil.

## Neler var

- URL başına önbelleğe alınan sayfalar ve üç strateji: bir kez render edilip siz
  geçersiz kılana kadar tutulan, bir TTL'den sonra yeniden render edilen ya da hiç
  önbelleğe alınmayan.
- Sayfalar arasında paylaşılan veri önbelleği: önbellek açıkken aynı yazarın otuz
  yazısı yazarı bir kez çeker.
- Kendi sayfalarına gönderilen formlar ve yerleşik istek sahteciliği koruması —
  önbellekteki bir sayfa da form taşıyabilir.
- Sayfa adlarından kurulan bağlantılar: bir sayfanın yolu değiştiğinde, sitenin her
  dilinde bağlantılar da onu izler.
- Başlıklar, meta etiketleri, stil dosyaları gibi head öğelerini, sayfada nerede
  durursa dursun, onlara ihtiyaç duyan fragment bildirir.
- Sitemap'ler, feed'ler ve `robots.txt` için HTML olmayan route'lar; geri kalan her
  şey için de kendi `http.Handler`'ınız.
- Siteyi herhangi bir statik barındırma hizmetine uygun dosyalara dönüştüren bir
  statik dışa aktarma. Okuduğunuz sayfalar onunla üretildi.
- Her Go değişikliğinde yeniden derleyen, her şablon değişikliğinde tarayıcıyı
  yenileyen bir geliştirme sunucusu. Render edilemeyen bir sayfa hatasını tarayıcıda
  gösterir; derlenmeyen Go kodu hatasını terminalde gösterir ve son sağlam build
  sunulmaya devam eder.

## Ne zaman doğru araç

collage, sayfaları çoğunlukla okunan sitelere uyar: bir blog, belgeler, bir tanıtım
sitesi, bir katalog, bir haber sitesi, bir uygulamanın herkese açık yüzü. İçeriğin
başka bir yerden — bir CMS'ten, bir veritabanından, bir API'den — geldiği ve
sayfaların bu kaynakların birkaçından aynı anda kurulduğu durumlarda en iyi hâlindedir.

Ekranları çoğunlukla etkileşimden oluşan bir uygulamaya ise pek uymaz: bir editör,
her saniye güncellenen bir pano, durumu tarayıcıda yaşayan herhangi bir şey. Bunlar
için bir istemci framework'ü kullanın; isterseniz etraflarındaki sayfalar için de
— [`app.Handle`](/docs/middleware-and-apis) üzerinden — collage'ı.

## Bilerek dışarıda bıraktıkları

collage bir veritabanı katmanı, bir oturum deposu, bir ORM, bir çeviri kataloğu ya
da bir JavaScript araç zinciriyle gelmez. Go yazıyorsunuz ve Go'nun bunların hepsine
zaten iyi yanıtları var; sizin yerinize birini seçen bir framework, etrafından
dolaşmanız gereken bir şey daha olurdu. collage'ın sahip olduğu kısım, sayfa render
etmeye özgü olan kısımdır: sayfaları bir araya getirmek, önbelleğe almak ve ne zaman
bayatladıklarını bilmek.

## Buradan nereye

[Kurulum](/docs/installation) bir dakikada çalışan bir proje verir.
[İlk sayfanız](/docs/your-first-page) sıfırdan bir sayfa kurar;
[Sayfalar ve layout'lar](/docs/pages-and-layouts) ise kavramların başladığı yerdir.
