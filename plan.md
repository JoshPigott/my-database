# I low key need to build a database but I got no idea on how to build a database

## Feature 
- Insert
- select
- update
- delete

- Like filtering like where, and, or

## Thing I will need to do
- I will need to try and understand the sturcture
- Like what a b-tree is
- Make sure no chagne inject stuff into the data
- Lock database


## Could good to watch when I am tired
https://www.youtube.com/watch?v=kQ3GJuflJN4

## useful info
https://mohllal.github.io/database-internals-part-1/?

## Next 
Next when the database get too big I will rewrite it


I am using this of somewhat of a plan
https://chatgpt.com/share/69f99c49-4b50-8323-9ccc-19574937e236

Basically 
Phase 1 — Log-based key-value store
Phase 2 — Persistence + crash recovery  (here)
Phase 3 — Page-based storage
Phase 4 — B-Tree index
Phase 5 — Tables
Phase 6 — Simple query engine
Phase 7 — Transactions (hard but important)
Phase 8 — Indexes

Maybe I should update docs at some point

But  B tree package next

Could be good to watch
https://www.youtube.com/watch?v=09E-tVAUqQw&pp=ygURYisgdHJlZSBleHBsYWluZWQ%3D

To implement B trees 
I will just look at page insertion
And worry about deletion later


Insertion flow
Search
↓
Insert
↓
Split
↓
Range scan (leaf links)
↓
Pages / disk
↓
Delete

Deletion flow
Delete key
↓
Node underflow?
↓
Borrow from sibling
↓
Update parent
↓
Can't borrow?
↓
Merge nodes
↓
Parent underflow?
↓
Repeat upward

b - tree
It is just like a binary tree I think but with muilplte key on one node
So maybe just make a b tree to start off with 
So work your way down untill leaf node then try and add value in order
If is does not fit, split the tree and the parent becomes the key
but the key also stay in the leaf as well

With deletion is b trees


I want to test the speed of some funcation at some point


So I want to add a child right.
