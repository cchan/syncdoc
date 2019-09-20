# Design 2: No Operational Transformation

A dumber and potentially lower-latency way to do this is to consider each writer as a remote keyboard.
This could also have decentralization potential if you use WebRTC, but let's start simple, with the server as a dumb relay.

## Idea
Each person has a cursor which has a start and end and moves when typing happens - essentially a virtual character.
Each person's keystrokes are copied to everyone else.

- Does "offline mode" work in this? You'd need a complete history of every keystroke and every move...
    well... not really. You just need to re-sync where your cursor is in the opinion of other users.
    - Ugh. No. This should happen simply as a matter of course in the protocol.
    - You could do the complete-history route and just cut off the history where we know everyone's up to date. How do we know this - someone closing the tab vs. going suddenly offline
- What happens when you move your cursor into a region where text is changing? Do you position it relative to the context? The previous cursor?
- Maybe there's insert mode and cursor-move mode, and switching between the modes requires a full blocking sync.


