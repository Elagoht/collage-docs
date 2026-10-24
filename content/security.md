---
description: What collage does against request forgery, open redirects, path tricks, cache leaks and oversized bodies — and what is left to you.
reference: SecurityConfig, ErrCSRFCrossOrigin, Vary
---

# Security

A bug in a framework is a bug in every site built on it, so collage does the
security work that every site would otherwise have to remember.

This page gathers what it does for you, in one place, with a link to where each
part is explained. What it does not do is at the end.

Every request meets collage's checks in the same order, and each one can answer
before what comes after it runs:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 490" role="img" aria-labelledby="sec-flow-title">
<title id="sec-flow-title">A request first gets the security headers, then its path is cleaned: a dirty path is redirected, an encoded slash is a 404 and a malformed escape a 400. A path mounted with plain app.Handle is then yours alone, and no middleware runs for it. Everything else meets the middleware, yours and the plugins'. A page, its actions and its fragment paths meet the page's guards before the cache; a guard's own answer is no-store. A page is served from the cache or rendered with a key that holds no cookie, and goes out private, no-cache if it is guarded. A fragment path is an action answering GET, never cached. An action meets the body limit, answering 413, then the origin and the token, answering 403 for an unsafe method, then your handler. In production any error page shows only the status.</title>
<defs><marker id="sec-flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">request</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="52" width="184" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">security headers</text>
<line class="d-flow" x1="207" y1="65" x2="218" y2="65" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="54" width="96" height="22" rx="11" /><text class="d-small" x="270" y="68" text-anchor="middle">every response</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="98" width="184" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">path cleaned</text>
<line class="d-flow" x1="207" y1="111" x2="218" y2="111" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="94" width="96" height="34" rx="6" /><text class="d-small" x="270" y="108" text-anchor="middle">redirect</text><text class="d-small" x="270" y="121" text-anchor="middle">404 · 400</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="23" y="144" width="184" height="36" rx="6" /><text class="d-on-c" x="115" y="160" text-anchor="middle">middleware</text><text class="d-on-c d-small" x="115" y="173" text-anchor="middle">yours and plugins'</text>
<path class="d-flow" d="M115 132H270V140" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="222" y="144" width="96" height="36" rx="6" /><text class="d-on-c" x="270" y="160" text-anchor="middle">app.Handle</text><text class="d-on-c d-small" x="270" y="173" text-anchor="middle">all yours</text>
<line class="d-flow" x1="115" y1="180" x2="115" y2="196" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="200" width="184" height="36" rx="6" /><text x="115" y="216" text-anchor="middle">guards, before cache</text><text class="d-small" x="115" y="229" text-anchor="middle">page, actions, fragment paths</text>
<line class="d-flow" x1="207" y1="218" x2="218" y2="218" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="201" width="96" height="34" rx="6" /><text class="d-small" x="270" y="215" text-anchor="middle">guard's answer</text><text class="d-small" x="270" y="228" text-anchor="middle">no-store</text>
<path class="d-flow" d="M95 236C95 248 80 246 80 258" marker-end="url(#sec-flow-arrow)" />
<path class="d-flow" d="M135 236C135 248 240 246 240 258" marker-end="url(#sec-flow-arrow)" />
<text class="d-small" x="50" y="254" text-anchor="middle">page</text>
<text class="d-small" x="272" y="254" text-anchor="middle">action</text>
<rect class="d-frame" x="4" y="262" width="152" height="36" rx="6" /><text x="80" y="278" text-anchor="middle">cache or render</text><text class="d-small" x="80" y="291" text-anchor="middle">key holds no cookie</text>
<line class="d-flow" x1="80" y1="298" x2="80" y2="310" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="4" y="314" width="152" height="36" rx="6" /><text x="80" y="330" text-anchor="middle">if guarded</text><text class="d-small" x="80" y="343" text-anchor="middle">private, no-cache</text>
<rect class="d-frame" x="164" y="262" width="152" height="36" rx="6" /><text x="240" y="278" text-anchor="middle">body limit</text><text class="d-small" x="240" y="291" text-anchor="middle">413 · too large</text>
<line class="d-flow" x1="240" y1="298" x2="240" y2="310" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="164" y="314" width="152" height="36" rx="6" /><text x="240" y="330" text-anchor="middle">origin, then token</text><text class="d-small" x="240" y="343" text-anchor="middle">403 · unsafe methods</text>
<line class="d-flow" x1="240" y1="350" x2="240" y2="362" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="164" y="366" width="152" height="26" rx="6" /><text class="d-on-c" x="240" y="383" text-anchor="middle">your handler</text>
<line class="d-rule" x1="23" y1="408" x2="297" y2="408" />
<rect class="d-frame" x="23" y="418" width="274" height="36" rx="6" /><text x="160" y="434" text-anchor="middle">any error, in production</text><text class="d-small" x="160" y="447" text-anchor="middle">status only: no text, no stack</text>
<rect class="d-frame" x="96" y="468" width="12" height="12" rx="2" /><text class="d-small" x="114" y="478">collage</text>
<rect class="d-c" x="180" y="468" width="12" height="12" rx="2" /><text class="d-small" x="198" y="478">yours</text>
</svg>
<figcaption>collage's own steps are outlined, and what is yours is filled. Every request gets the security headers and a clean path before any middleware runs. A path mounted with plain app.Handle is then yours alone: no middleware, no guard, no forgery check, no body limit and no cache. Everything else meets the middleware, yours and the plugins'. A page, its actions and its fragment paths meet the page's guards before the cache. An action then meets the body limit (the body is bounded before any middleware runs; an action whose body is too large is refused here), then the origin and the token, before your handler; a fragment path is an action answering GET, and is never cached. In production, an error page shows the status and nothing else.</figcaption>
</figure>

## Forms cannot be forged

Every action with an unsafe method checks a signed token before its handler runs,
and a page with `{{csrfToken}}` is still cached: each reader is sent their own
token.

The token is not checked alone. A submission the browser marks as coming from
another origin — `Sec-Fetch-Site`, or `Origin` against `Host` — is refused
whatever it carries, sibling subdomains included, so a planted cookie is no way in.
An origin meant to post here is named in `Security.CSRFTrustedOrigins`. See
[Forms and actions](/docs/forms-and-actions#forgery-protection).

Set `Security.CSRFKey` before deploying: a generated key differs in every process.

A token carries the time it was issued, signed, and is refused once it is older
than `Security.CSRFTokenTTL` — twelve hours by default, so a leaked token cannot
be replayed forever.

A reader with a long-open form is sent a fresh token rather than refused; set
`CSRFTokenTTL` negative to keep a token valid for as long as its signature.

Since v0.72.0 the token in a form is masked anew in every response: XORed with a
fresh random pad and prefixed `~`. A compressed response that reflects what a
request sent is shorter when the reflection matches part of the token, and an
attacker who can make a reader's browser send requests could read the token from
those lengths (BREACH). Masked, it is different bytes every time. The check takes
the mask off, from the form and the header, and still accepts an unmasked token.

## Redirects stay on the site

No redirect collage writes can send a reader to another site:

- The redirect that sends `/a/./b` or `/a//b` to its clean spelling writes the path
  escaped, so `/./%5Cevil.com` goes to `/%5Cevil.com`, not to `/\evil.com`, which a
  browser reads as `//evil.com`.
- Locale and trailing-slash redirects never begin with `//`.
- A registered redirect to `//host` or `/\host` is refused at registration, and a
  value it captured is escaped for where it lands — in the path, or as one query
  value, so `/login?next={slug}` cannot be handed a second `next`.

An action's `Location` and a guard's are yours: collage writes them as you set
them, so do not build one from the request without checking it. Some must leave
the site (a sign-in provider's page), which is why collage does not check them.

For a `next` taken from the URL, collage offers its own check, since v0.44.0:

```go
loc := collage.SafeRedirect(rc.Request.URL.Query().Get("next"), "/")
```

`SafeRedirect` returns `next` when it is a path on this site — it starts with
`/`, not with `//`, holds no backslash and no control character, and still starts
with a single `/` once cleaned — and the fallback otherwise (`/` when the fallback
is not one either).

The cleaning matters because `http.Redirect` cleans a rooted path:
`/./\evil.example` would leave as `/\evil.example`, which a browser reads as
`//evil.example`. An absolute URL is never accepted, even one on this site's own
origin.

## A path means one thing

Middleware reads the decoded `r.URL.Path`; the router reads the escaped one. Where
the two could disagree, the request reaches nothing:

- A path with dot segments or doubled slashes is redirected to its clean spelling
  before middleware runs, so a check on `/admin/` never meets `/x/../admin`.
- A path holding an encoded slash, `%2F`, is a 404. `%2F` is not a separator, but
  a middleware reads the decoded path, where `/public%2Fsecret` is `/public/secret`:
  two segments to the middleware, one to the router.
- A malformed escape is a 400, not a 500.

See [Paths are cleaned first](/docs/middleware-and-apis#paths-are-cleaned-first).

## Private pages stay private

A [guard](/docs/pages-and-layouts#private-pages-guards) runs before the cache is
read, for the page's renders, the actions on its URL and its fragment paths.

A guarded page goes out `private, no-cache`, and a guard's own answer `no-store`, so
a CDN never hands one reader's page or refusal to the next. A static export does
not write guarded pages at all.

## A cached page is nobody's in particular

A render stored in the cache is served to every reader whose request has its key,
so it is handed only what the key holds:

- the path and the host;
- the query parameters in the key;
- the headers declared with `collage.Vary`.

It is never handed a reader's cookies, credentials or address. The host is in the
key, so a request naming another host cannot poison the copy everyone else gets.

A degraded render, an error page and a page in development are never publicly
cacheable. See [What a shared render sees](/docs/caching#what-a-shared-render-sees).

## Bodies are bounded

A request body is limited — 4 MiB unless you say otherwise — before your middleware
runs, at the limit of the action it routes to; since v0.71.0 also when there is no
middleware, so a page's or an action's guard reads it bounded. A body that says it is
larger is refused before the handler runs.

Multipart files spilled to disk are removed when the request has been answered, whoever
parsed the form, and a
cookie the forgery check never signed is refused before the body is read at all.
See [Request bodies are bounded](/docs/forms-and-actions#request-bodies-are-bounded).

## Files are what they say they are

An [asset mount](/docs/assets) types a file by its extension. One whose name says
nothing is typed from its content, so an extensionless image upload is still an
image.

It is never typed into a type that runs: one that looks like HTML or XML is
`text/plain`, and with `nosniff` the browser keeps to that, so an upload holding
markup is not served as a page.

A mount never lists a directory, and never serves or exports a dotfile — `.env`,
`.git/` — except `.well-known`. `collage serve` does the same.

## Errors say nothing in production

In production, collage's own error page shows the status and nothing else: no
error text, no stack, no path.

The development overlay and the reload channel exist only in development, and
`collage dev` answers only a `Host` naming this machine, so a page on another site
cannot read them through DNS rebinding.

A request path in a log line is quoted when it holds a control character, so it
cannot forge a line or rewrite your terminal.

## Security headers

Every response carries `X-Content-Type-Options: nosniff` and
`X-Frame-Options: SAMEORIGIN` by default, so a site has MIME-sniff and clickjacking
protection on its forms without adding anything.

`Security.FrameOptions` sets the value — `"-"` sends none, anything else is sent
verbatim — and `Security.DisableNoSniff` turns nosniff off (`NoSniff *bool` before
v0.60.0).

A full `Content-Security-Policy`, HSTS and the rest are what the
[elagoht/secure](/docs/plugins#elagohtsecure) plugin adds, and its headers override
these; a production page carries no framework script, so a strict policy needs no
nonce.

## What is left to you

- **A Go toolchain with the security fixes.** A program carries the standard library
  it was built with. Since v0.72.0 collage requires Go 1.26.10, so with the default
  `GOTOOLCHAIN=auto` the go command fetches it for a build that would have used an
  older Go. Check a binary with `go version -m`. Go 1.26.10 rejects the compromised
  eMudhra intermediate certificates (CVE-2026-111943).
- **TLS**, at a proxy that passes `X-Forwarded-Proto` and the browser's `Host` on.
  See [Deployment](/docs/deployment#tls-behind-a-proxy).
- **Rate limits** — [elagoht/ratelimit](/docs/plugins#elagohtratelimit).
- **A handler mounted with `app.Handle`** is your own: no forgery check, no body
  limit, no cache, and since v0.67.0 no middleware — no plugin's authentication, rate
  limit, ban or access log covers it unless it is mounted `collage.WithMiddleware()`.
  See [No middleware, unless you ask](/docs/middleware-and-apis#no-middleware-unless-you-ask).
- **What a data handler reads.** On a cacheable page, a value middleware put in the
  request context is hidden from the shared render, so one reader's cannot reach
  another's page; declare it with `collage.Vary` and read it with `collage.Varied`.

To report a vulnerability, use GitHub's private reporting on the
[collage repository](https://github.com/Elagoht/collage/security).

## Example: a team's task list, before it goes live

The team task list from
[Middleware and your own API](/docs/middleware-and-apis#example-a-teams-task-list)
and [Forms and actions](/docs/forms-and-actions#example-a-team-task-list) is about
to go live:

- every member signs in, and has a role: an admin may change any task of the team,
  a member only the tasks assigned to them;
- the team's list is a page at `/teams/{team}/tasks`, with an "add a task" form;
- a JSON API for the team's command-line tool is mounted with `app.Handle`,
  `WithMiddleware()`.

Each risk below is covered by collage, by a plugin the application adds, or by
the application itself:

| Risk | collage, built in | Plugin | Still the app's job |
| --- | --- | --- | --- |
| Another site posts the "add a task" form | Origin, then token, on every action with an unsafe method | — | `{{csrfToken}}` in the form; `Security.CSRFKey` set |
| Who is signed in | — | elagoht/session: a signed cookie, `HttpOnly`, `SameSite=Lax`, encrypted with `Encrypt` | The login action; `Regenerate` when a reader signs in |
| A signed-out reader on the list | Guards run before the cache, for the page, its actions and its fragment paths | `session.RequireUser("/login")` as the guard | Putting the guarded layout on every private page |
| A member changing another member's task, or another team's | — | — | The role and team check, where the data is: a guard such as `requireRole`, and the action and the API |
| `next` after sign-in sends the reader off the site | `collage.SafeRedirect` | `RequireUser` puts `next` in the URL | Passing `next` through `SafeRedirect` |
| `/x/../admin` slipping past a path check | Path cleaned before middleware | — | — |
| One member's page served to another | A guarded page is `private, no-cache`; the cache key holds no cookie | elagoht/session: a request with a session is rendered fresh, `private, no-store` | — |
| A huge upload | 4 MiB per action by default; `WithMaxBodyBytes` | — | The API's own bound: `http.MaxBytesReader` |
| Password guessing on `/login` | — | elagoht/ratelimit: a `/login` rule | `Server.TrustedProxies` behind a proxy; buckets are per process |
| Bots filling a form | — | elagoht/honeypot: `{{honeypot}}`, checked after the forgery check, before the handler | Pairing it with elagoht/ratelimit |
| Injected scripts, framing, plain HTTP | `nosniff`; `X-Frame-Options: SAMEORIGIN` | elagoht/secure: a CSP with a nonce, HSTS | The policy itself; TLS at the proxy |
| An error showing a database address | The production error page shows the status alone | — | The error text the API writes itself |
| A stolen session cookie | — | elagoht/session: a session cannot be revoked, it lives in the cookie | A short `maxAge` or `idleTimeout` |

The plugins go in `Config.Plugins`: elagoht/secure and elagoht/honeypot add
template functions (`{{cspNonce}}`, `{{honeypot}}`). A plugin's key is read from an
environment variable — `SESSION_KEY`, `SECURE_KEY`, `HONEYPOT_KEY`, or another its
`KeyEnv` names — never from `plugins-config.json`:

```go
app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{
		// The proxy that ends TLS, so ClientIP and IsHTTPS believe it.
		TrustedProxies: []string{"10.0.1.10"},
	},
	Security: collage.SecurityConfig{
		CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	},
	Plugins: []collage.Plugin{
		session.New(session.Options{Encrypt: true}),
		secure.New(secure.Options{
			CSP: "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
		}),
		ratelimit.New(ratelimit.Options{Rules: []ratelimit.Rule{
			{PathPrefix: "/login", Rate: 0.1, Burst: 5},
			{Rate: 0.5, Burst: 10},
		}}),
		honeypot.New(honeypot.Options{}),
	},
})
```

The second rule has no path and no methods, so it counts every method but `GET`,
`HEAD` and `OPTIONS` everywhere outside its skip list — the API's `PATCH` too, since the API is mounted
`WithMiddleware()`. A plain `app.Handle` mount runs no middleware, so neither this
rate limit nor any other plugin's middleware would cover it.

The roles stay the application's. The guard that lets only admins through is the
`requireRole` of
[Example: a blog with a private area](/docs/pages-and-layouts#example-a-blog-with-a-private-area),
and the API checks the team and the assignee itself, as
[The API, mounted as it is](/docs/middleware-and-apis#the-api-mounted-as-it-is)
shows: collage says nothing about who may change what.
