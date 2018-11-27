Design
======

A *Document* is a string of characters.
- SEE BELOW FOR DATASTRUCTURES

A *Splice* consists of a start character index, an end character index, and an inserted string.
- Start is inclusive, End is exclusive.

A *Changeset* consists of a splice and the revision number of the History it is based on.

A *History* is a fixed circular buffer of HISTORY_LENGTH Splices, along with a current revision number.

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

A *GetDocument* is an O(DOCUMENT_LENGTH) operation that:
- Acquires ReadLock
- Converts the Document to a String
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
    func GetDocument()
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
    - Connection.Send(Group.SyncDoc.GetDocument())
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
            - Perhaps during insertion and deletion steps, be clever about it
                - i.e. when inserting, can copy to a new node, overwrite, modify links and lengths.
- Which things need to be locked?
    - mostly done
