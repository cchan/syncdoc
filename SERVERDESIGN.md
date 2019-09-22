https://blog.cloudflare.com/how-to-receive-a-million-packets/
https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb

- A fast typist can hit 10 keystrokes per second.
- A busy document might have 20 users.
- 50,000 packets per second per core is a rule-of-thumb processing capacity estimate.

Design goal is to handle as many such documents as possible. A reasonable goal might be 100,000 keystrokes per second across 500 open documents on a (dual-core??) Raspberry Pi on a high-latency connection (plus Nginx for ssl termination). (Since it'll be running on something like a micro or nano instance on AWS/GCP, shared with a bunch of other servers, this is probably reasonable.)

The goal of the project is to provide a collaborative editing service similar to Google Docs but:
- faster, more server-efficient
- better controlled (markdown is nice)
- with latex integration for hw assignments? (KaTeX is capable of live compile with delims https://github.com/KaTeX/KaTeX/issues/712)
Must-have:
- has Google SSO
- has cursor

Things to add
- A benchmarker
- Mirror frontend operations onto LocalStorage, persistent through restarts, which can be copied to the remote when possible, and only remove them once ack'd by the server.
- fasthttp, etc.
- Template app.html with preloaded content (valyala/quicktemplate)
- Audit keystroke hotpaths in the code for mem allocs, loops, etc.
  - Truly zero mem allocs requires using something like Flat Buffers on the browser side.
  - Remove all locking too...
  - Could just use https://github.com/valyala/fastjson instead.
  - Again, benchmark benchmark benchmark. The bottleneck is probably the locks and the json.
