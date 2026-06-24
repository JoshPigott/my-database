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

**Root page struture**

**Metadata page struture**

**Routing page struture**
- The point of this page is not link key to slotID to eliminate
 the need for seraching the page/s to find a where a key slot is
- metadata
    - page type (MetadataPage)
    - numEntries 
    - freeSpaceStart
    - freeSpaceEnd
- routes
    - Note this may be 100% atm
    - [boundaryKey] [pageID] [slotID]

**Data / leaf node struture**
- metadata
    - page type (DataPage)
    - numEntries 
    - freeSpaceStart
    - freeSpaceEnd
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

**Node disk strutrue leaf**
- metadata
- Next leaf pageID
- key 0 length
- key 0 
- PageID (where value 0 is)
- slotID (where value 0 is)
- key 1 length
- key 1
- PageID (where value 1 is)
- slotID (where value 1 is)


**Node disk strutrue internal**
- metadata
- child 0 PageID 
- key 0 length
- key 0
- child 1 PageID
- key 1 length
- key 1
- child 2 PageID

note duplicates keys cannot be added to the tree will regret second keys