// Package ui holds the words of the site that are not documentation: the header,
// the page furniture, the home page. The documentation itself is content, in
// content/ and its translations; these are the few strings around it.
package ui

// Text is every string the templates show, in one language.
type Text struct {
	// Name is the language's name in itself, for the language switcher.
	Name string

	Docs              string
	GoReference       string
	SearchPlaceholder string
	SearchLabel       string
	SearchResults     string
	// SearchEmpty and SearchFailed are read by static/search.js, which puts the
	// query for {query} and the error for {error}.
	SearchEmpty  string
	SearchFailed string
	LanguageNav  string
	Footer       string

	Contents      string
	Documentation string
	OnThisPage    string
	Pages         string
	Previous      string
	Next          string

	HomeTitle       string
	HomeDescription string
	HomeHeadline    string
	HomeTagline     string
	ReadGuide       string
	ViewOnGitHub    string
	WhatYouGet      string
	Points          []Point
	TheGuide        string

	NotFoundTitle     string
	NotFoundHeadline  string
	NotFoundBefore    string
	NotFoundGuideLink string
	NotFoundAfter     string
}

// Point is one of the home page's four claims.
type Point struct {
	Title string
	Text  string
}

var texts = map[string]Text{
	"en": {
		Name:              "English",
		Docs:              "Docs",
		GoReference:       "Go reference",
		SearchPlaceholder: "Search the docs",
		SearchLabel:       "Search the documentation",
		SearchResults:     "Results",
		SearchEmpty:       "Nothing matches “{query}”.",
		SearchFailed:      "Search is unavailable ({error}).",
		LanguageNav:       "Language",
		Footer:            "collage is MIT-licensed. These pages are a collage site, exported to static files.",

		Contents:      "Contents",
		Documentation: "Documentation",
		OnThisPage:    "On this page",
		Pages:         "Pages",
		Previous:      "Previous",
		Next:          "Next",

		HomeTitle:       "collage — server-rendered pages from cached fragments, in Go",
		HomeDescription: "A Go framework for server-rendered pages composed from cached fragments. No dependencies.",
		HomeHeadline:    "Pages composed from cached fragments.",
		HomeTagline:     "collage is a Go framework for server-rendered sites. A page is a layout around fragments; each fragment fetches its own data, and the page is cached and invalidated by what it was built from. No client framework, no JavaScript build step, no dependencies.",
		ReadGuide:       "Read the guide",
		ViewOnGitHub:    "View on GitHub",
		WhatYouGet:      "What you get",
		Points: []Point{
			{"Fragments, not components", "A template, a data handler and the slots it exposes. Siblings fetch concurrently; the page is written in order."},
			{"Cached by what it depends on", "Pages and data carry tags. Change an author and one call drops the author and every page that shows them."},
			{"Forms without JavaScript", "Actions answer POST with forgery protection built in, and a cached page can still carry a form."},
			{"Served or exported", "One program is a server and a static site generator. This site is the second."},
		},
		TheGuide: "The guide",

		NotFoundTitle:     "Not found — collage",
		NotFoundHeadline:  "There is nothing here.",
		NotFoundBefore:    "The page may have moved. The ",
		NotFoundGuideLink: "guide",
		NotFoundAfter:     " starts at the beginning.",
	},
	"tr": {
		Name:              "Türkçe",
		Docs:              "Belgeler",
		GoReference:       "Go referansı",
		SearchPlaceholder: "Belgelerde ara",
		SearchLabel:       "Belgelerde ara",
		SearchResults:     "Sonuçlar",
		SearchEmpty:       "“{query}” ile eşleşen bir şey yok.",
		SearchFailed:      "Arama şu an kullanılamıyor ({error}).",
		LanguageNav:       "Dil",
		Footer:            "collage MIT lisanslıdır. Bu sayfalar, statik dosyalara dışa aktarılmış bir collage sitesidir.",

		Contents:      "İçindekiler",
		Documentation: "Belgeler",
		OnThisPage:    "Bu sayfada",
		Pages:         "Sayfalar",
		Previous:      "Önceki",
		Next:          "Sonraki",

		HomeTitle:       "collage — önbelleğe alınmış fragment'lerden sunucuda render edilen sayfalar, Go ile",
		HomeDescription: "Önbelleğe alınmış fragment'lerden oluşan, sunucuda render edilen sayfalar için bir Go framework'ü. Bağımlılığı yok.",
		HomeHeadline:    "Önbelleğe alınmış fragment'lerden kurulan sayfalar.",
		HomeTagline:     "collage, sunucuda render edilen siteler için bir Go framework'ü. Bir sayfa, fragment'lerin etrafını saran bir layout'tur; her fragment kendi verisini çeker, sayfa da neyden kurulduysa ona göre önbelleğe alınır ve geçersiz kılınır. İstemci framework'ü yok, JavaScript build adımı yok, bağımlılık yok.",
		ReadGuide:       "Rehberi oku",
		ViewOnGitHub:    "GitHub'da gör",
		WhatYouGet:      "Neler var",
		Points: []Point{
			{"Component değil, fragment", "Bir şablon, bir data handler ve dışarı açtığı slot'lar. Kardeş fragment'ler verilerini aynı anda çeker; sayfa yine sırasıyla yazılır."},
			{"Neye bağlıysa ona göre önbellekte", "Sayfalar ve veriler etiket taşır. Bir yazarı değiştirin; tek bir çağrı hem yazarı hem onu gösteren her sayfayı önbellekten düşürür."},
			{"JavaScript'siz formlar", "Action'lar POST isteklerini yerleşik sahtecilik korumasıyla karşılar; önbellekteki bir sayfa da form taşıyabilir."},
			{"Sunulur ya da dışa aktarılır", "Tek bir program hem sunucu hem statik site üreticisidir. Bu site ikincisi."},
		},
		TheGuide: "Rehber",

		NotFoundTitle:     "Bulunamadı — collage",
		NotFoundHeadline:  "Burada bir şey yok.",
		NotFoundBefore:    "Sayfa taşınmış olabilir. ",
		NotFoundGuideLink: "Rehber",
		NotFoundAfter:     " en baştan başlıyor.",
	},
}

// For returns the strings for locale, or English for a locale the site has no
// strings for.
func For(locale string) Text {
	if text, ok := texts[locale]; ok {
		return text
	}
	return texts["en"]
}
