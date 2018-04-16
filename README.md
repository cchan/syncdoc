# SyncDoc

A collaborative editing tool written in Go by [Clive Chan](https://clive.io).

Copyright (c) 2018, all rights reserved, for now.

## TODO

x separate this into neater files
x ISSUE: when you newline above a line with content, that line gets cleared
- ISSUE: onconnect full-doc update duplicates the document if the client disconnects & reconnects; SOLUTION: have separate Edit and FullUpdate events
  - Maintain a local queue of changes while offline
- ISSUE: bounds checks for change events before slicing
- ISSUE: ctrl-a delete doesn't sync
x periodically dump History results to plaintext file, and use that as the base
  - keep enough history (1000 entries?) that it can still merge any latecomers
  - full-file updates periodically (every 10s?)
- OT is pretty easy; to test it just give a setTimeout delay before js send edit to server
- collapse history entries (hard to preserve indexes) - should probably happen mainly on js side? for now don't worry about load
- WSS LetsEncrypt - not just CloudFlare
- Use a linkedlist for Document connections?
- Add cursors for other users (cursorPosition events)
