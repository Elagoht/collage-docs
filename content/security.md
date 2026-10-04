---
description: What collage does against request forgery, open redirects, path tricks, cache leaks and oversized bodies — and what is left to you.
reference: SecurityConfig, ErrCSRFCrossOrigin, Vary
---

# Security

A bug in a framework is a bug in every site built on it, so collage does the
security work that every site would otherwise have to remember. This page gathers
what it does for you, in one place, with a link to where each part is explained.
What it does not do is at the end.

## Forms cannot be forged

Every action with an unsafe method checks a signed token before its handler runs,
and a page with `{{csrfToken}}` is still cached: each reader is sent their own
token. The token is not checked alone. A submission the browser marks as coming
from another origin — `Sec-Fetch-Site`, or `Origin` against `Host` — is refused
whatever it carries, sibling subdomains included, so a planted cookie is no way in.
An origin meant to post here is named in `Security.CSRFTrustedOrigins`. See
[Forms and actions](/docs/forms-and-actions#forgery-protection).

Set `Security.CSRFKey` before deploying: a generated key differs in every process.

A token carries the time it was issued, signed, and is refused once it is older
than `Security.CSRFTokenTTL` — twelve hours by default, so a leaked token cannot
be replayed forever. A reader with a long-open form is sent a fresh token rather
than refused; set `CSRFTokenTTL` negative to keep a token valid for as long as its
signature.

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
is not one either). The cleaning matters because `http.Redirect` cleans a rooted
path: `/./\evil.example` would leave as `/\evil.example`, which a browser reads as
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
read, for the page's renders, the actions on its URL and its fragment paths. A
guarded page goes out `private, no-cache`, and a guard's own answer `no-store`, so
a CDN never hands one reader's page or refusal to the next. A static export does
not write guarded pages at all.

## A cached page is nobody's in particular

A render stored in the cache is served to every reader whose request has its key,
so it is handed only what the key holds — path, host, the query parameters in the
key, the headers declared with `collage.Vary` — and never a reader's cookies,
credentials or address. The host is in the key, so a request naming another host
cannot poison the copy everyone else gets. A degraded render, an error page and a
page in development are never publicly cacheable. See
[What a shared render sees](/docs/caching#what-a-shared-render-sees).

## Bodies are bounded

A request body is limited — 4 MiB unless you say otherwise — before your middleware
runs, at the limit of the action it routes to. Multipart files spilled to disk are
removed when the action has answered, and a cookie the forgery check never signed
is refused before the body is read at all. See
[Request bodies are bounded](/docs/forms-and-actions#request-bodies-are-bounded).

## Files are what they say they are

An [asset mount](/docs/assets) types a file by its extension. One whose name says
nothing is typed from its content, so an extensionless image upload is still an
image — but never into a type that runs: one that looks like HTML or XML is
`text/plain`, and with `nosniff` the browser keeps to that, so an upload holding
markup is not served as a page. A mount never lists a directory, and never serves
or exports a dotfile — `.env`, `.git/` — except `.well-known`. `collage serve` does
the same.

## Errors say nothing in production

In production, collage's own error page shows the status and nothing else: no
error text, no stack, no path. The development overlay and the reload channel
exist only in development, and `collage dev` answers only a `Host` naming this machine, so a
page on another site cannot read them through DNS rebinding. A request path in a
log line is quoted when it holds a control character, so it cannot forge a line or
rewrite your terminal.

## Security headers

Every response carries `X-Content-Type-Options: nosniff` and
`X-Frame-Options: SAMEORIGIN` by default, so a site has MIME-sniff and clickjacking
protection on its forms without adding anything. `Security.FrameOptions` sets the
value — `"-"` sends none, anything else is sent verbatim — and `Security.NoSniff`
turns nosniff off when it points at `false`. A full `Content-Security-Policy`, HSTS
and the rest are what the [elagoht/secure](/docs/plugins#elagohtsecure) plugin adds,
and its headers override these; a production page carries no framework script, so a
strict policy needs no nonce.

## What is left to you
- **TLS**, at a proxy that passes `X-Forwarded-Proto` and the browser's `Host` on.
  See [Deployment](/docs/deployment#tls-behind-a-proxy).
- **Rate limits** — [elagoht/ratelimit](/docs/plugins#elagohtratelimit).
- **A handler mounted with `app.Handle`** is your own: no forgery check, no body
  limit, no cache.
- **What a data handler reads.** On a cacheable page, a value middleware put in the
  request context is hidden from the shared render, so one reader's cannot reach
  another's page; declare it with `collage.Vary` and read it with `collage.Varied`.

To report a vulnerability, use GitHub's private reporting on the
[collage repository](https://github.com/Elagoht/collage/security).
