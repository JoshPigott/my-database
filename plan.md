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
 

## deletion logic
So there are three cases to think about

1. simple case leaf has enough key so just delete

2. Borrow from a silbling left or right (this includes updating speartores)

3. No silbings have any keys so you have to merge nodes

plus update the linked list after


I think there will be some edge cases when the node gets really small

Look at the linked list