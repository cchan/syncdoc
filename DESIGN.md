Design
======

A *Document* is a string of characters.
- SEE BELOW FOR DATASTRUCTURES

A *Splice* consists of a start character index, an end character index, and an inserted string.
- Start is inclusive, End is exclusive.

A *Changeset* consists of a splice and the revision number of the History it is based on.

A *History* is a fixed circular buffer of HISTORY_LENGTH Splices, along with a current revision number. This revision number is monotonically increasing for the entire syncdoc.

A *Syncdoc* consists of current up-to-date Document and a History of recent changes.

An *Onto* is an O(1) operation of a Splice CURR using a Splice BASE that:
- (CCBB) IF CURR.end < BASE.start THEN do nothing
- (BBCC) IF CURR.start > BASE.end THEN CURR.start += BASE.str.len - (BASE.end - BASE.start)
                                       CURR.end   += BASE.str.len - (BASE.end - BASE.start)
- (CBCB) CURR.end = BASE.start
- (BCBC) CURR.start = BASE.start + BASE.str.len
- (BCCB) CURR.start = BASE.start + BASE.str.len
         CURR.end   = BASE.start + BASE.str.len
- (CBBC) CURR.end += BASE.str.len - (BASE.end - BASE.start)

A *Rebase* is an O(HISTORY_LENGTH) operation on a Changeset and a History that:
- Performs an Onto for the last (History.revNum - Changeset.baseRev). O(HISTORY_LENGTH)
- Fails if the Changeset baseRev is more than History.splices.length older than the revNum of the History.

An *Append* is an O(1) operation that:
- Appends the Splice of a Changeset to a History.
- Fails if the Changeset baseRev is not equal to the History revNum.

An *Apply* is an O(DOCUMENT_LENGTH) operation that:
- Modifies a Document according to a Splice. (i.e. delete the range start-end, insert the string there)

A *CherryPick* is an O(HISTORY_LENGTH + DOCUMENT_LENGTH) operation that:
- Acquires ReadLock
- Rebases a Changeset on a Syncdoc.History O(HISTORY_LENGTH)
- Upgrades ReadLock to WriteLock
- Appends a Changeset to a Syncdoc.History O(1)
- Applies a Changeset.Splice to a Syncdoc.Document O(DOCUMENT_LENGTH)
- Releases WriteLock

A *ToStringWithRevNum* is an O(DOCUMENT_LENGTH) operation that:
- Acquires ReadLock
- Converts the Syncdoc.Document to a String
- Retrieves Syncdoc.History.revNum
- Releases ReadLock

```
struct Syncdoc{
    struct Document{
        contents string

        func Apply(Splice)
    }
    struct Splice{
        startIndex int
        endIndex   int
        insertion  string

        func Onto(Splice)
    }
    struct Changeset{
        splice Splice
        baseRev int

        func Rebase(History)
    }
    struct History{
        splices circular_buffer<Splice>
        revNum int
        
        func Append(Changeset)
    }
    func CherryPick(Changeset)
    func ToString()
    doc Document
    hist History
}
```

An *ID* is a string that uniquely identifies a Syncdoc.
- Must be alphanumeric + _ + -

A *Connection* is a user connected by WebSockets.
- When connecting to an ID, Server attaches them to the corresponding Group.

A *Packet* contains an ID and a Changeset.

A *Group* is a group of Connections associated with a certain Syncdoc.
- Maintains list of Connections
- Group.Attach(Connection)
    - Group.connectionList.append(Connection)
    - Connection.Send(Group.SyncDoc.ToStringWithRevNum)
- Group.Receive(Changeset)
    - Group.Syncdoc.CherryPick(Packet.Changeset)
    - for each Group.Connections: Connection.Send(Changeset)

A *Server* contains a mapping of IDs to Groups.
- When receiving a Packet, it uses the ID to direct Packet.Changeset to Group.Receive()

Issues to Resolve
-----------------
- Apply is inefficient.
    - Document data structure candidates:
        - LinkedList cannot be efficiently range searched.
        - LinkedHashMap cannot be efficiently range searched.
        - (Self-balancing) order statistic tree has O(logn) range delete, O(logn) insert
            - O(n) space
            - Leaves are up to 32 or 64 byte strings or something? Not sure what's optimal
        - Skip list has O(logn) range delete, O(logn) insert
            - O(nlogn) space
        - Problem with linked data structures is that it's very difficult to commit to disk efficiently.
            - I may just have to be clever and DIY mempage allocation
            - Database needs to store ID => Document mapping. That's it.
            - Maaaaybe it just stores the string? That could work fine tbh lol
        - Another problem with any of these data structures is that they need to be pruned periodically.
            - i.e. keeping the document in chunks of up to 1kb will gradually fragment smaller and smaller
        - Insertion splits the current string and copies to a new node.
    - Is this write-optimized?
    - The data structure requires:
        - An extremely efficient RangeSplice() operation.
        - Efficient committing to disk.
        - Doesn't need efficient readouts to string, since those are rarer.
    - What if it's just a string plus a list of splices? Then periodically the list of splices can be resolved back down to length zero.
        - Repeated splices still aren't that efficient without a treelike algorithm.
        - Commit to disk every time it's resolved. (Don't commit the tree itself, just the document with revnum.)
            - Committing to disk means writing it to a temp file, then renaming over it.
            - Use this for commits: https://github.com/natefinch/atomic
        - Needs to writelock during every Resolve operation until it's written to disk.
    - Use an order-statistic tree of strings.
        - Use normal string addition and splicing per-node, unless it's over a certain limit. Then split.
        - Hmm so just insert-after? I guess it will have to sort on some arbitrary index attached to each string.
            - https://github.com/FX-HAO/GoOST
            - I really don't want to implement AVL balancing lol.
    - OH WHAT IF I SORT THE EDITS BY START-INDEX (stable sort so tiebreak by time) and then StringBuilder it?
        - this is a lot smarter lol
        - just have to deal with all the rebases...
        - one pass algorithm: have a fenwick tree that stores char diff lengths by position within the base Document, then for each splice:
            - SUBTRACT the running sum in the fenwick tree from the current splice, effectively rebasing it onto the base string
                - note that we want the running sum up to AND INCLUDING the searched-for index.
            - store the splice into the fenwick tree by the start index and the number of characters of diff (pos or neg)
        - THEN AFTER THIS... stable-sort the rebased splices by start index, and then StringBuilder.
        - WAIT maybe this doesn't work rip
            - You can't reposition within inserted segments. Like if you want to start with a blank doc and A@0, B@1, C@2, D@1, you can't. It'll just result in A@0, B@0, C@0, D@0, and then insert by historical order and result in ABCD.
        - Maybe come up with a better way of resolving the order between things at the same position.
- Undoability
    - Be able to create inverses of operations.
    - https://hackernoon.com/operational-transformation-the-real-time-collaborative-editing-algorithm-bf8756683f66
    - Somehow we need consistency no matter what order the operations are applied in.
        - What if both clients send edits at the same time? What does the server send back to indicate in which order the edits should be applied?
- Every keystroke for every person will be sent.
    - Merge some keystrokes together
    - Client-side caching of word-typing
        - Fixed setTimeout-based
        - The way Google Docs does it is it waits for an ACK before sending the next Packet.
- Todo: show others' cursors
- Client-side conflict resolution.
    - If I make an edit and it gets rebased onto other stuff server-side... we'd have to lock and resolve before allowing the user to continue typing? Maybe not.
    - First of all, how do I even apply edits to a textarea? It would be full-string replacement of the entire textarea every time.
        - The way GDocs does it is by a DOM tree. Each DOM node additionally has an ID.
        - We can do it similarly. Every insertion will affect only the leafmost DOM node.
        - Let's do it by lines. It should be good enough. Lines can be inserted and deleted at will, the only thing that actually has to be replaced is the cursor.
        - Can I do a contenteditable div with StackEdit? (Do I need stackedit? all I need is efficient live transformation of Markdown + KaTeX)
            - Can I do differential application of edits to LaTeX formulas? Would that require implementing KaTeX myself essentially?
- Which things need to be locked?
    - mostly done


OKAY BACK TO THE DRAWING BOARD A LITTLE BIT.
--------------------------------------------
What kind of operation would be idempotent AND commutative? <= scratch that.
Okay, given that the client will wait for a server ACK before sending the next changeset, we can get by with causal consistency.
Context-based changesets, instead of position-based.
(sidenote: might be good to use nodejs now? bc this is gonna be a lot of shared code)
How do we resolve A@0 and B@0 sent from two clients?
- Suppose we decide by order of arrival, and OT based on that.
    - Client1: A@0. Client2: B@0. Client3: C@0. Server: A@0 B@1 C@2.
- We then need to send these results back. What does this update packet look like?
    - Ideally, Client1 += B@1 and Client2 an A@0 packet.
    - This could be done quite efficiently by bubbling events backwards in the history, and OT'ing for each swap. i.e. bubbling C@2 backwards would mean
        - A@0 B@1 C@2
        - A@0 C@1 B@1
        - C@0 A@0 B@1
    - Suppose C@1 happens on Client3 in the meanwhile. So it looks like C@0, C@1, and then receives A@0, B@1.
    - Again we'd bubble them through each other until all events in the received packet have their desired parents (C@0, A@0, B@1, ), and then we send any events hanging off the edge (in this case C@1) to the server.
        - Wait so what's our ordering? Do we always assume edits sent from server happened before our own edits? Does this work?
            - Yes I think so.
        - How do we determine which parent is desired? We could use edit numbers. How would those sync across server and client?
            - Client3: 0[C@0] 1[C@1] receives 0[C@0] 1[A@0] 2[B@1]
- Could we do a system where every node including server and clients must do all the OT magic themselves? Same code shared. The server contains an OTClient that provides initialization to new connections, and the OTServer is just a dumb forwarder.
    - Kind of like git - distributed version control.
    - No we can't completely. OTServer must at least determine an ordering over all edits. Or does it? Can we have commutative edits?

