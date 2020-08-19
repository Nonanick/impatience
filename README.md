## Impatience - 0.0.2

---

HTTP2 Static WebServer in GO!

### Why?

---

* With the maturity of the HTTP2 Protocol + the large support for ES6
  import / export syntax + js modules in browser I've decided to build
  something that would simplify serving web front projects in dev environments.
* ATM you're required to use a bundler or at least implement a 'build'
  step when testing your projects which can be CPU/time consuming !
  But by taking advantage of HTTP2 push capabilities Impatience can analyze
  all static server files looking for dependencies and serve them in a single
  push!
* I like the NodeJS environment but I also like speed! So I'm trying to
  solve the HTTP2 AutoPush feature in a compiled language, if compatilibity
  with NodeJS env proves to be rough I shall consider also creating a TS
  version of Impatience!
  ** ATM i'm trying to implement file "transformers", the general ideia
  is to be able to parse a file before serving it and then hold a copy
  of the parsed file in memory until the underlying file changes.
  This would allow to serve on demand .ts, .tsx, .vue and so on
  Since most of the transpilers already exists in the NodeJS environment I
  would like to bridge it by consuming the transpiled output to Impatience
  in GO using the command line!

### Pros

---

- Smart caching of files by using a file watcher (fsnotify)
- Compiled language web server
- Use the full power of ES6 import/export syntax
- Quickly hosts a dev project without a build step
- NodeModules served transparently in the browser (when implemented)

### Cons

---

- Still to young, bugs might be numerous
- Begginer dev in GO, I've literally used this project to learn GO...
- Written in a new language (for those coming from JS), so you rely on
  others if something isn't working
- No tree-shaking
- Bundlers also applies minification to code
- In some scenarios (probably on the real world not on localhost) HTTP2
  autopush proves to have equivalent performance

So Impatience server or any HTTP2 webservers in the wild that offers auto-push might still
not be the standart for delivering your web products, anyway devs might experience
a gain in quality-of-life while developing and having better response times

### Disclaimer

---

Version 0.0.2 isn't there for nothing, this project is something that I
think that can be handy, someday, but is also an experiment! I'm mastering
the GoLang as I develop the WebServer.

You can use it AS IS with NO GUARANTIES! This should not hit production for
a while!

### The core idea

---

* [ ] Accept command-line configurations
* [ ] Accept JSON / JS / TS config file ( to act like command-line config )
* [ ] At server boot, crawl into the public root searching for files
* [ ] Query each known file type (using extension) for related files/dependencies
* [ ] When a file is requested by the client check dependencies and try to push
  as many as bandwidth allows
* [ ] Serve node_modules libraries that are required by the JS/TS files - The Developer
  should be aware if the library can be run in the browser!
* [ ] Hold information on a CacheStore about the recently served and not modified files

* [ ] Cookie strategy implemented!
* [ ] ServiceWorker strategy showing some difficulties:

Only one service worker per context (might collide with your own!)

- Self-Signed certs + SW don't get along nicely D:
- Once installed upgrading the SW based on server content is hard


[X] Keep track of modified files to invalidate cache (watch file system) - Done using fsnotify

### Extras + TODOs

---

[ ] ?? Apply file "transformations" with the option to serve them from memory or
from file dump  (ex: TS -> JS, TSX -> JSX, Vue -> JS, etc.)  - This feature will
probably envolve calling Node using exec and then consuming the output
[ ] ?? Hot Module Reload, auto reload page when one of the current page dependencies
update, apart from the keep-alive connection working along-side the HTTP2 server
the way things are structured implementing it shouldn't be hard!
-- Requires JS code to contact server ( fake path? /impatience/listen/fileChanges)
-- Requires Server to emit events to client
