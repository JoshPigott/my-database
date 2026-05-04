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

## I am just going to start building somthing every simple


## Next 
get this point before I commit 
I am going to assume the most simple input e.g. no space in input
db.Set("a", "1")
db.Set("a", "2")
db.Get("a") → "2"
db.Delete("a")


SET a 1
SET a 2
DELETE a

I think I will need to strus in





Some left over code aye
// func getData(scanner *bufio.Scanner) ([][]string, error) {
// 	// Note I am assume all key and value have no spaces
// 	var err error
// 	data := [][]string{}
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		command, info, found := strings.Cut(line, " ")

// 		if found == false {
// 			continue
// 		}

// 		if command == "SET" {
// 			// Gets key length
// 			keyLengthStr, info, found := strings.Cut(info, " ")
// 			if found == false {
// 				return data, errors.New("Unable to get data")
// 			}
// 			keyLength, err := strconv.Atoi(keyLengthStr)
// 			if err != nil {
// 				continue
// 			}
// 			// Gets key value
// 			keyContent := info[0:keyLength]

// 			// Gets value length
// 			_, info, found = strings.Cut(info[keyLength:], " ")
// 			if found == false {
// 				return data, errors.New("Unable to get data")
// 			}
// 			// Gets value length
// 			valueContent := info
// 			// Add to data
// 			output := []string{keyContent, valueContent}
// 			data = append(data, output)
// 		}
// 		if command == "DELETE" {
// 			key := info
// 			newData := make([][]string, 0, len(data))

// 			for i := 0; i < len(data); i++ {
// 				if data[i][0] != key {
// 					newData = append(newData, data[i])
// 				}
// 			}
// 			data = newData
// 		}
// 	}
// 	return data, err
// }

