package main

import (
	"bubbly-database/internal/database"
	"fmt"
	"io"
)

var DB *database.DB

func main() {
	testDatabase()
}

func testDatabase() {
	var err error
	DB, err = database.Open()
	if err != nil {
		fmt.Println(err)
	}

	lookAtFile()
	// lookAtData()
	if err := DB.Close(); err != nil {
		fmt.Println(err)
	}
}

func lookAtFile() {
	info, err := DB.Pages.File.Stat()
	if err != nil {
		fmt.Println("failed to get file size: %w", err)
	}
	fileSize := info.Size()
	fmt.Println("info.Size()", fileSize)
	// for key, value := range data {
	// 	fmt.Println("key:", key, "value:", value)
	// }
	bytes := make([]byte, fileSize)
	if _, err := DB.Pages.File.Seek(0, io.SeekStart); err != nil {
		fmt.Println("failed to move start of file to read")
	}
	n, err := DB.Pages.File.Read(bytes)
	if err != nil {
		fmt.Println("failed to read bytes aye", err)
	}

	// I need to seek to the start here

	fmt.Println("Number of bytes read:", n)
	for i := 0; i < int(fileSize)/4096; i++ {
		fmt.Println()
		fmt.Println("page:", (i + 1))
		fmt.Println(bytes[i*4096 : (i+1)*4096])
	}
	// fmt.Println(bytes)

}

func lookAtData() {
	data, err := DB.SelectAll()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Map length:", len(data))
	for key, value := range data {
		fmt.Printf("key: %-12s value: %-22s\n", key, value)
	}
}

// For test adding lots of different key values
func addKeysValues() {
	pairs := map[string]string{
		"name":        "James Robertson",
		"city":        "Auckland New Zea",
		"email":       "james@example.co",
		"phone":       "021-555-7823-NZ",
		"country":     "New Zealand land",
		"language":    "English speaking",
		"hobby":       "Mountain climbing",
		"pet":         "Golden retriever!",
		"car":         "Toyota Corolla NZ",
		"job":         "Software engineer",
		"food":        "Butter chicken yum",
		"sport":       "Rugby union games",
		"color":       "Ocean blue shade",
		"music":       "Classic rock band",
		"book":        "Lord of the rings",
		"movie":       "The dark knight NZ",
		"travel":      "Backpacking Europe",
		"drink":       "Flat white coffee",
		"school":      "University of Otago",
		"team":        "All Blacks rugby",
		"weather":     "Sunny with clouds",
		"season":      "Summer in December",
		"instrument":  "Acoustic guitar jam",
		"game":        "Table tennis match",
		"fruit":       "Feijoa and kiwi",
		"vegetable":   "Kumara sweet potato",
		"brand":       "Kathmandu outdoor",
		"show":        "Flight of conchords",
		"animal":      "Kiwi bird unique",
		"landmark":    "Sky Tower Auckland",
		"x9#mQ":       "p!7vLw$2kR@nZoY",
		"fj32d**&":    "hello world 1234",
		"TrEe_77":     "wistful afternoon",
		"@sunrise!":   "burnt orange glow",
		"k4$$hBnZ":    "fjord mist rising",
		"ZIP_code":    "1010 AKL central",
		"MOON":        "crescent silver ok",
		"blogPost3":   "draft needs editing",
		"rnd_92!xP":   "velvet underground",
		"$$$profit":   "not financial advice",
		"wX.f99~z":    "quantum foam layer",
		"HELLO???":    "goodbye cruel world",
		"_flat_":      "pancake stack yum",
		"j32fd**&":    "xZ!9q#mLp$2wRkNv",
		"turtle99":    "slow and steady pace",
		"^_^emoji":    "kawaii desu ne yes",
		"ZZZZ_top":    "sharp dressed person",
		"fog::lamp":   "damp road at night",
		"NZ_brew42":   "craft ale hoppy ipa",
		"p!n3appl3":   "does not go on pizza",
		"__init__":    "python class method",
		"404found":    "ironic key name lol",
		"#!bash":      "shebang line script",
		"th3_v01d":    "staring back at you",
		"ReD_bUs77":   "route number twelve",
		"~wave~":      "sinusoidal function",
		"CAPSLOCK":    "accidentally on key",
		"3.14159":     "approximately pi ok",
		"<<EOF>>":     "heredoc terminator",
		"zzz...":      "sleeping process now",
		"nUlL_pTr":    "segfault incoming soon",
		"xX_gamer_Xx": "mountain dew supply",
		"[bracket]":   "syntax error maybe",
		"time_t":      "unix epoch seconds",
		"&ampersand":  "html entity escape",
		"f(x)=mx+b":   "linear equation form",
		"???unknown":  "value undetermined yet",
		"RuNtImE!!":   "error at line nine",
		"0xDEADBEEF":  "classic hex constant",
		"snow_day!":   "school cancelled yes",
		"::colon::":   "namespace separator",
		"w1ld_k4rd*":  "glob pattern match",
		"lemon#drop":  "cocktail or candy ok",
		"OVERCLOCKED": "cpu running very hot",
		"b4cksl4sh\\": "escape character fun",
		"@@@triple@":  "unknown symbol meaning",
		"fuzzy_wuzz":  "bear had no hair sad",
		"GR4PH_ql":    "query language api",
		"d33p_sp4ce":  "no one hears scream",
		"m00n_walk":   "armstrong footprint",
		"!important":  "css override cheat",
		"recursive_":  "see key recursive_",
		"NaN_value":   "not a number result",
		"QWERTY_top":  "keyboard row one ok",
		"zsh_alias":   "ll stands for ls al",
		"pH_level7":   "neutral water balance",
		"$tr1ng!!":    "concatenated result ok",
		"WARP_9.5":    "faster than light now",
		"fog_index":   "readability score low",
		"~tilde~op":   "home directory linux",
		"br0kenLink":  "404 page not found",
		"SNR_db40":    "signal noise ratio ok",
		"^caret^up":   "regex start anchor",
		"plnktn_99":   "microscopic organism",
		"wildcard":    "sql like clause fun",
		"oBj3ct_K3y":  "property value pair",
		"VOID_main":   "c entry point func",
		"s@ltedHash":  "bcrypt password safe",
		"tr1g_func":   "sine cosine tangent",
		"<<DEVNULL>>": "discard output pipe",
		"M4RS_2031":   "colonisation target",
		"raw_socket":  "tcp handshake begin",
		"[NULL_SET]":  "empty collection ok",
		"entropy!!":   "disorder increasing",
		"YOLO_flag":   "deploy to prod now",
		"d4rk_m4tt3r": "undetectable mass ok",
		"PKT_L0SS!!":  "connection unstable rn",
		"$h3ll_cmd":   "bash injection risk",
		"BUFFR_OVF":   "stack smashing found",
		"z3r0_d4y!!":  "patch it immediately",
		"~~ripple~~":  "water surface effect",
		"GC_pause_":   "garbage collected now",
		"MUTEX_lock":  "thread waiting pls",
		"f1b0_seq":    "one one two three",
		"[REDACTED]":  "classified info here",
		"s3cr3t_k3y":  "not actually secret",
		"BLOAT_ware":  "using 4gb of ram",
		"@deprecated": "use new method pls",
		"lazy_eval":   "computed when needed",
		"KERN_panic":  "system halted oops",
		"#hashtag99":  "trending topic now",
		"Xn--unicode": "punycode domain ok",
		"p0ly_fill":   "browser compat shim",
		"COLD_b00t":   "fresh start no cache",
		"^^XOR_gate":  "true false logic op",
	}

	for k, v := range pairs {
		DB.AddToPage(k, v)
	}
}

// func testBTreee() {
// 	var entries = []struct {
// 		key    string
// 		pageID uint32
// 		slotID uint16
// 	}{
// 		{key: "name", pageID: 3, slotID: 12},
// 		{key: "city", pageID: 7, slotID: 4},
// 		{key: "email", pageID: 1, slotID: 19},
// 		{key: "phone", pageID: 9, slotID: 27},
// 		{key: "country", pageID: 5, slotID: 8},
// 		{key: "language", pageID: 11, slotID: 3},
// 		{key: "hobby", pageID: 2, slotID: 22},
// 		{key: "pet", pageID: 8, slotID: 15},
// 		{key: "car", pageID: 4, slotID: 30},
// 		{key: "job", pageID: 6, slotID: 1},
// 		{key: "food", pageID: 10, slotID: 17},
// 		{key: "sport", pageID: 3, slotID: 25},
// 		{key: "color", pageID: 7, slotID: 9},
// 		{key: "music", pageID: 1, slotID: 14},
// 		{key: "book", pageID: 12, slotID: 6},
// 		{key: "movie", pageID: 5, slotID: 28},
// 		{key: "travel", pageID: 9, slotID: 2},
// 		{key: "drink", pageID: 2, slotID: 20},
// 		{key: "school", pageID: 6, slotID: 11},
// 		{key: "team", pageID: 4, slotID: 24},
// 		{key: "weather", pageID: 8, slotID: 5},
// 		{key: "season", pageID: 11, slotID: 16},
// 		{key: "instrument", pageID: 3, slotID: 29},
// 		{key: "game", pageID: 7, slotID: 7},
// 		{key: "fruit", pageID: 10, slotID: 13},
// 		{key: "vegetable", pageID: 1, slotID: 21},
// 		{key: "brand", pageID: 5, slotID: 18},
// 		{key: "show", pageID: 9, slotID: 10},
// 		{key: "animal", pageID: 2, slotID: 26},
// 		{key: "landmark", pageID: 6, slotID: 23},
// 	}

// 	t := database.NewBTree()
// 	for _, e := range entries {
// 		if e.key == "weather" {
// 			t.CheckStructure(1)
// 		}
// 		t.Insert(e.key, e.pageID, e.slotID)
// 	}
// 	t.CheckStructure(2)
// }
