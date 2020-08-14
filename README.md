## Impatience - 0.0.1
----------------------

HTTP2 Static WebServer in GO!

### Why?
---------
* With the maturity of the HTTP2 Protocol + the large support for ES6
import / export syntax + js modules in browser I've decided to build
something that would simplify serving web front projects.

* ATM you're required to use a bundler or at least implement a 'build'
step when testing your projects, but by taking advantage of HTTP2 push 
capabilities Impatience can analyze all server files looking for 
dependencies and serve them in a single push as needed!

* I like the NodeJS environment but I also like speed! So I'm trying to
solve the HTTP2 AutoPush feature in a compiled language, if compatilibity
with NodeJS env proves to be rough I shall consider also creating a TS
version of Impatience!

### Disclaimer
--------------
Version 0.0.1 isn't there for nothing, this project is something that I
think that can be handy, someday, but is also an experiment! I'm mastering
the GoLang as I develop the WebServer.

You can use it AS IS with NO GUARANTIES! This should not hit production for
a while!

### The core idea
-----------------
[] At server boot crawl into the public root searching for files  
[] Query each known file type (using extension) for related files/dependencies  
[] When a file is requested by the client check dependencies and try to push 
as many as bandwidth allows  
[] Serve node_modules libraries that are required by the JS/TS files - Developer should be aware if the library is browser compatible! 
[] ?? Apply file transformations with the option to serve them from memory or
from file dump  (ex: TS -> JS, TSX -> JSX, Vue -> JS, etc.)
[] ?? Hold information on a CacheStore about the recently served and not modified files
[] ?? Keep track of modified files to invalidate cache (watch file system)