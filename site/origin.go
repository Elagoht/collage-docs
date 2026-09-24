package site

// Origin is where the site is published — the custom domain set on the GitHub Pages
// repository — and what canonical links and the sitemap are absolute against.
const Origin = "https://collage.furkanbaytekin.dev"

// Original is the language the documentation is written in, served without a
// prefix; Translations are the languages it is translated into, each served under
// its own, "/tr/docs/caching".
const Original = "en"

var Translations = []string{"tr"}

// Locales is every language the site is in, the original first.
func Locales() []string { return append([]string{Original}, Translations...) }
