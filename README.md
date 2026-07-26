## My database
- In this project it creates a database

## current constraints
- metadata file must set before hand
- Key value format
- Note I am assume all key have no spaces
    - If there are space it will break (I will fix this later on)
    - commard (SET DELETE), key, value

## Data stucture
- The database is made up of pages
- Three type of pages root page
Each page size is 4096 bytes or 4KB


**Page structure**

![slot page structure image](/screenshots/sloted-page.png)

**Slot structure**

![slot structure](/screenshots/slot.png)

**Page metadata**
- metadata
    - page type 
    - page id
    - numEntries 
    - freeSpaceStart
    - freeSpaceEnd

**Database metadata page struture**
- page metadata
- page id of b+tree root 
- total number of pages
- next free page
- page id of last used data page

**Data pages struture**
- page metadata
- slots
    - Note this starts at the bottom of the metadata and fills downwards
    - The point of the slot is to store data about how the data is store and if it relevant
    - [offset] [length] [flag]
        - offset stores where in the page the data starts
        - length stores how long the entry is 
        - flag stores is normal or deleted ect
- free space
- data
    - Note this start of the page and fills upwards
    - [keyLength] [valueLength] [key] [value]


## Free Pages
- Pages with type `0` are free (unallocated).
- Free pages are stored as a singly linked list.
- Each free page stores the `nextFreePageID`.
- The head of the free page linked list is stored in database metadata.
- Insert/delete at head of linked list → O(1).

## B+ Tree Features & Structure

### Overview
- A B+ tree stores many keys per node, reducing tree height and disk reads.
- All operations (search, insert, delete) are **O(log n)** due to shallow depth.
- Internal nodes store **only separator keys** (no data).
- Leaf nodes store **all actual data pointers** (`pageID` + `slotID`) in `keyLocations`.
- Leaf nodes are linked by `pageID` as a linked list for fast range queries.
- Tree has a single **root node**, linked store in database metadata.
- Each node maps to a **disk page** (one node = one page).

### Key and node ordering
- Internal nodes use separator keys to route searches:
  - Left child: keys `< separator`
  - Right child: keys `>= separator`
- Key comparison uses `strings.Compare(a, b)` (lexicographic / ASCII order).
- All keys are sorted inside each node

---

## Operations / features

### `Insert`
- Start at root → descend to correct leaf node.
- Insert key in sorted order.
- If node overflows:
  - Split node into two.
  - Promote separator key (usually first key of right node) to parent.
- Parent may recursively split (can propagate to root).

---

### Search `Select`
- Start at root.
- Traverse internal nodes using separator keys.
- Reach leaf node.
- Find key → get `(pageID, slot)`.
- Fetch data from disk in data page and return it.

---

### Range Query `SelectWhere`
- Uses linked leaf nodes.
- `>` or `>=`: find first matching leaf, scan right through linked list.
- `<` or `<=`: start at leftmost leaf, scan until boundary key.

---

### Delete
- Find key via normal traversal.
- Remove key from leaf node.
- Handle underflow if node has too few keys:
  - `redistribution` with sibling, or
  - `merge` with sibling node.
- Update parent separator keys if structure changes.
- When `merge` right node get delete and page becomes free page

---

### Node page structure
**Leaf node**
- page metadata
- Next leaf pageID
- key 0 length
- key 0 
- PageID (where value 0 is)
- slotID (where value 0 is)
- key 1 length
- key 1
- PageID (where value 1 is)
- slotID (where value 1 is)

**Internal node**
- page metadata
- child 0 PageID 
- key 0 length
- key 0
- child 1 PageID
- key 1 length
- key 1
- child 2 PageID

## Page Cache
- Keeps recently used pages in memory to avoid disk reads  
- Fixed size; old pages are removed when new ones are needed  
- Tracks usage with a double-linked list  

## WAL & Transactions
- WAL ensures changes are all applied together or not at all  
- Changes are written to WAL before being applied to the database  
- If a crash happens before commit, changes are discarded  
- On restart, WAL is used to recover or ignore incomplete changes  
- Manual transactions apply this all-or-nothing behavior across multiple operations  


## Further things
- Allow logic operations with select. I already have the AST; I just need to turn AND into one search combining the conditions, and OR into separate searches. Or just add results and and just combine filters.

## Know promblems 
- duplicates keys cannot be added to the b+tree will regret second key
- Reading of slots and data could just be one read but they are two atm
- Manual transitions can crash if too many database changes occur, as the page cache is limited and there’s currently no WAL undo support.
## Notes 
- Some tests where maybe by AI for increased development speed 
