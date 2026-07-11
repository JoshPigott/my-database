package database

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

/*
Note a few of these tests have been write by AI
*/

func TestOpenClose(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")
	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestBasicInsert(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")
	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	key := "the_key"
	value := "the_value"
	if err := db.AddToPage(key, value); err != nil {
		t.Errorf("failed to add %q -> %q: %v", key, value, err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestOneNodeInsert(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")
	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	tests := []struct {
		key string
		val string
	}{
		{"dataset_user_profile_preferences_theme_selection_identifier_v1", "01"},
		{"dataset_analytics_session_retention_tracking_metadata_field_v2", "ab"},
		{"customer_account_notification_delivery_settings_record_key_v1", "7f"},
		{"system_generated_user_permission_mapping_configuration_entry", "ff"},
		{"application_cache_invalidation_dependency_reference_key_v3", "10"},
		{"long_form_transaction_processing_status_tracking_identifier", "de"},
		{"background_worker_checkpoint_progress_metadata_reference_key", "42"},
		{"internal_index_page_split_debugging_test_identifier_record", "99"},
		{"persistent_storage_serialization_validation_entry_identifier", "aa"},
		{"temporary_btree_leaf_node_overflow_testing_reference_field", "be"},
	}

	for _, tc := range tests {
		if err := db.AddToPage(tc.key, tc.val); err != nil {
			t.Errorf("failed to add %q -> %q: %v", tc.key, tc.val, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestInsert(t *testing.T) {
	amount := 1500
	var insertedKeys []string

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle them
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Insert in random order
	for _, i := range indices {
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		// fmt.Printf("DB.AddToPage(\"%s\", \"%s\")\n", key, val)
		db.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestInsertWithClose(t *testing.T) {
	amount := 2000
	var insertedKeys []string

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle them
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Insert in random order
	for _, i := range indices {
		if i == amount/2 {
			// Closes and reopens file / database
			if err := db.Close(); err != nil {
				t.Errorf("failed to close database: %v", err)
			}
			db, err = openDefault(filename, walFilename)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
		}
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		// fmt.Printf("DB.AddToPage(\"%s\", \"%s\")\n", key, val)
		db.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestOneNodeDelete(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")
	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	tests := []struct {
		key string
		val string
	}{
		{"dataset_user_profile_preferences_theme_selection_identifier_v1", "01"},
		{"dataset_analytics_session_retention_tracking_metadata_field_v2", "ab"},
		{"customer_account_notification_delivery_settings_record_key_v1", "7f"},
		{"system_generated_user_permission_mapping_configuration_entry", "ff"},
		{"application_cache_invalidation_dependency_reference_key_v3", "10"},
		{"long_form_transaction_processing_status_tracking_identifier", "de"},
		{"background_worker_checkpoint_progress_metadata_reference_key", "42"},
		{"internal_index_page_split_debugging_test_identifier_record", "99"},
		{"persistent_storage_serialization_validation_entry_identifier", "aa"},
		{"temporary_btree_leaf_node_overflow_testing_reference_field", "be"},
	}

	for _, tc := range tests {
		if err := db.AddToPage(tc.key, tc.val); err != nil {
			t.Errorf("failed to add %q -> %q: %v", tc.key, tc.val, err)
		}
	}
	for _, tc := range tests {
		if err := db.Delete(tc.key); err != nil {
			t.Errorf("failed to delete %q: %v", tc.key, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestDelete(t *testing.T) {
	amount := 2000
	var insertedKeys []string

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle them
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Insert in random order
	for _, i := range indices {
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		// fmt.Printf("DB.AddToPage(\"%s\", \"%s\")\n", key, val)
		db.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
	}

	// Shuffle them
	rand.Shuffle(len(insertedKeys), func(i, j int) {
		insertedKeys[i], insertedKeys[j] = insertedKeys[j], insertedKeys[i]
	})

	for _, key := range insertedKeys {
		if err := db.Delete(key); err != nil {
			t.Errorf("failed to delete %q: %v", key, err)
		}
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close databasen: %v", err)
	}
}

func TestDeleteWithClose(t *testing.T) {
	amount := 2000
	var insertedKeys []string

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle them
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Insert in random order
	for _, i := range indices {
		if i == amount/2 {
			// Closes and reopens file / database
			if err := db.Close(); err != nil {
				t.Errorf("failed to close database: %v", err)
			}
			db, err = openDefault(filename, walFilename)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
		}
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		// fmt.Printf("DB.AddToPage(\"%s\", \"%s\")\n", key, val)
		db.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
	}

	// Shuffle them
	rand.Shuffle(len(insertedKeys), func(i, j int) {
		insertedKeys[i], insertedKeys[j] = insertedKeys[j], insertedKeys[i]
	})

	for i, key := range insertedKeys {
		if i == amount/2 {
			// Closes and reopens file / database
			if err := db.Close(); err != nil {
				t.Errorf("failed to close database: %v", err)
			}
			db, err = openDefault(filename, walFilename)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
		}
		if err := db.Delete(key); err != nil {
			t.Errorf("failed to delete %q: %v", key, err)
		}
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestSelectInsertDeleteReopen(t *testing.T) {
	amount := 1000

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	inserted := make([]string, 0, amount)

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// -------------------------
	// INSERT + SELECT CHECK
	// -------------------------
	for _, i := range indices {
		key := fmt.Sprintf("key_%04d", i)
		val := fmt.Sprintf("val_%04d", i)

		if err := db.AddToPage(key, val); err != nil {
			t.Fatalf("insert failed: %v", err)
		}

		inserted = append(inserted, key)

		// UPDATED SELECT API
		backValue, isFound, err := db.Select(key)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if !isFound {
			t.Fatalf("expected key to exist: %s", key)
		}
		if backValue != val {
			t.Fatalf("wrong value: got %s want %s", backValue, val)
		}

		// halfway reopen
		if i == amount/2 {
			if err := db.Close(); err != nil {
				t.Fatalf("close failed: %v", err)
			}

			db, err = openDefault(filename, walFilename)
			if err != nil {
				t.Fatalf("reopen failed: %v", err)
			}
		}
	}

	// -------------------------
	// FINAL VERIFY AFTER INSERT
	// -------------------------
	for _, key := range inserted {
		expected := "val_" + key[4:]

		backValue, isFound, err := db.Select(key)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if !isFound {
			t.Fatalf("missing key: %s", key)
		}
		if backValue != expected {
			t.Fatalf("bad value: got %s want %s", backValue, expected)
		}
	}

	// -------------------------
	// DELETE HALF
	// -------------------------
	rand.Shuffle(len(inserted), func(i, j int) {
		inserted[i], inserted[j] = inserted[j], inserted[i]
	})

	toDelete := inserted[:amount/2]

	for _, key := range toDelete {
		if err := db.Delete(key); err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		// UPDATED SELECT CHECK
		_, isFound, err := db.Select(key)
		if err != nil {
			t.Fatalf("select after delete failed: %v", err)
		}
		if isFound {
			t.Fatalf("expected key to be deleted: %s", key)
		}
	}

	// -------------------------
	// REOPEN AGAIN
	// -------------------------
	if err := db.Close(); err != nil {
		t.Fatalf("final close failed: %v", err)
	}

	db, err = openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("final reopen failed: %v", err)
	}

	// -------------------------
	// FINAL CHECK AFTER REOPEN
	// -------------------------
	for i, key := range inserted {
		expected := "val_" + key[4:]

		backValue, isFound, err := db.Select(key)

		if i < amount/2 {
			// deleted keys
			if err != nil {
				t.Fatalf("select failed: %v", err)
			}
			if isFound {
				t.Fatalf("expected deleted key missing: %s", key)
			}
			continue
		}

		// existing keys
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if !isFound {
			t.Fatalf("missing key after reopen: %s", key)
		}
		if backValue != expected {
			t.Fatalf("bad value after reopen: got %s want %s", backValue, expected)
		}
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestSelect(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	key := "basic_key"
	val := "basic_value"

	// -------------------------
	// INSERT
	// -------------------------
	if err := db.AddToPage(key, val); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	// -------------------------
	// SELECT (first check)
	// -------------------------
	backValue, isFound, err := db.Select(key)
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if !isFound {
		t.Fatalf("expected key to exist")
	}
	if backValue != val {
		t.Fatalf("wrong value: got %s want %s", backValue, val)
	}

	// -------------------------
	// CLOSE DB
	// -------------------------
	if err := db.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// -------------------------
	// REOPEN DB
	// -------------------------
	db, err = openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}

	// -------------------------
	// SELECT AGAIN (persistence check)
	// -------------------------
	backValue, isFound, err = db.Select(key)
	if err != nil {
		t.Fatalf("select after reopen failed: %v", err)
	}
	if !isFound {
		t.Fatalf("expected key to persist after reopen")
	}
	if backValue != val {
		t.Fatalf("wrong value after reopen: got %s want %s", backValue, val)
	}

	// -------------------------
	// CLEAN CLOSE
	// -------------------------
	if err := db.Close(); err != nil {
		t.Fatalf("final close failed: %v", err)
	}
}

// -------------------------
// Select all testing
// -------------------------

func TestSelectAll_Basic(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer db.Close()

	cases := []struct {
		key string
		val string
	}{
		{"user_1", "alice"},
		{"user_2", "bob"},
		{"user_3", "charlie"},
	}

	for _, tc := range cases {
		if err := db.AddToPage(tc.key, tc.val); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	got, err := db.SelectAll()
	if err != nil {
		t.Fatalf("SelectAll failed: %v", err)
	}

	if len(got) != len(cases) {
		t.Fatalf("expected %d items, got %d", len(cases), len(got))
	}

	for _, tc := range cases {
		if got[tc.key] != tc.val {
			t.Errorf("key %s: expected %s, got %s", tc.key, tc.val, got[tc.key])
		}
	}
}

func TestSelectAll_WithDeletes(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer db.Close()

	expected := map[string]string{}

	// insert
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("k_%d", i)
		val := fmt.Sprintf("v_%d", i)

		db.AddToPage(key, val)
		expected[key] = val
	}

	// delete half
	for i := 0; i < 500; i += 2 {
		key := fmt.Sprintf("k_%d", i)

		if err := db.Delete(key); err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		delete(expected, key)
	}

	got, err := db.SelectAll()
	if err != nil {
		t.Fatalf("SelectAll failed: %v", err)
	}

	for k, v := range expected {
		if got[k] != v {
			t.Errorf("key %s mismatch", k)
		}
	}

	for k := range got {
		if _, ok := expected[k]; !ok {
			t.Errorf("unexpected key returned: %s", k)
		}
	}
}

func TestSelectAll_Stress(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer db.Close()

	const N = 2000

	expected := map[string]string{}

	// INSERT PHASE
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key_%05d", i)
		val := fmt.Sprintf("val_%05d", i)

		if err := db.AddToPage(key, val); err != nil {
			t.Fatalf("insert failed at %d: %v", i, err)
		}

		expected[key] = val
	}

	// OVERWRITE SOME KEYS
	for i := 0; i < N; i += 10 {
		key := fmt.Sprintf("key_%05d", i)
		val := "overwritten"

		if err := db.AddToPage(key, val); err != nil {
			t.Fatalf("overwrite failed: %v", err)
		}

		expected[key] = val
	}

	// SELECT ALL
	got, err := db.SelectAll()
	if err != nil {
		t.Fatalf("SelectAll failed: %v", err)
	}

	// VERIFY
	if len(got) != len(expected) {
		t.Fatalf("size mismatch: expected %d got %d", len(expected), len(got))
	}

	for k, v := range expected {
		if got[k] != v {
			t.Errorf("mismatch key=%s expected=%s got=%s", k, v, got[k])
		}
	}
}

// -------------------------
// Select where testing
// -------------------------

func TestSelectWhere(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	const (
		base   = 1000000000000000000
		amount = 500
	)

	// Insert keys:
	// 1000000000000000000 -> "0"
	// ...
	// 1000000000000000499 -> "499"
	for i := 0; i < amount; i++ {
		key := strconv.Itoa(base + i)
		val := strconv.Itoa(i)

		if err := db.AddToPage(key, val); err != nil {
			t.Fatalf("failed to add %q -> %q: %v", key, val, err)
		}
	}

	// Delete a couple to make sure deleted records
	// never appear in results.
	deleted1 := strconv.Itoa(base + 132)
	deleted2 := strconv.Itoa(base + 250)

	if err := db.Delete(deleted1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := db.Delete(deleted2); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	tests := []struct {
		name      string
		condition MathConditions
		boundary  string
		wantCount int
	}{
		{
			name:      "greater than",
			condition: GreaterThan,
			boundary:  strconv.Itoa(base + 250),
			wantCount: 249, // 251..499
		},
		{
			name:      "greater than or equal",
			condition: GreaterThanOrEqualTo,
			boundary:  strconv.Itoa(base + 250),
			wantCount: 249, // 250 deleted
		},
		{
			name:      "less than",
			condition: LessThan,
			boundary:  strconv.Itoa(base + 250),
			wantCount: 249, // 0..249 except 132
		},
		{
			name:      "less than or equal",
			condition: LessThanOrEqualTo,
			boundary:  strconv.Itoa(base + 250),
			wantCount: 249, // 250 deleted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.SelectWhere(tt.condition, tt.boundary)
			if err != nil {
				t.Fatalf("SelectWhere failed: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Fatalf(
					"expected %d rows, got %d",
					tt.wantCount,
					len(got),
				)
			}

			// Deleted records should never appear.
			for _, d := range got {
				if d.key == deleted1 {
					t.Fatalf("deleted key %q returned", deleted1)
				}
				if d.key == deleted2 {
					t.Fatalf("deleted key %q returned", deleted2)
				}
			}
		})
	}
}
func TestSelectWhere_ReopenDB(t *testing.T) {
	db, dir, deleted1, deleted2 := setupSelectWhereDB(t)
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db.Close()

	// reopen DB from disk
	db2, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to reopen db: %v", err)
	}
	defer db2.Close()

	tests := []struct {
		name      string
		condition MathConditions
		boundary  string
		wantCount int
	}{
		{
			name:      "greater than",
			condition: GreaterThan,
			boundary:  strconv.Itoa(1000000000000000000 + 250),
			wantCount: 249,
		},
		// same table...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db2.SelectWhere(tt.condition, tt.boundary)
			if err != nil {
				t.Fatalf("SelectWhere failed: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Fatalf("expected %d, got %d", tt.wantCount, len(got))
			}

			for _, d := range got {
				if d.key == deleted1 || d.key == deleted2 {
					t.Fatalf("deleted key returned after reopen: %q", d.key)
				}
			}
		})
	}
}

// Note this utils funcation for TestSelectWhere_ReopenDB
func setupSelectWhereDB(t *testing.T) (db *DB, dir string, deleted1, deleted2 string) {
	dir = t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	const (
		base   = 1000000000000000000
		amount = 500
	)

	for i := 0; i < amount; i++ {
		key := strconv.Itoa(base + i)
		val := strconv.Itoa(i)

		if err := db.AddToPage(key, val); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	deleted1 = strconv.Itoa(base + 132)
	deleted2 = strconv.Itoa(base + 250)

	if err := db.Delete(deleted1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := db.Delete(deleted2); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	return db, dir, deleted1, deleted2
}

func TestCrashHelperEntrypoint(t *testing.T) {
	if os.Getenv("BUBBLY_HELPER_PROCESS") != "1" {
		t.Skip("helper entrypoint — only runs as a subprocess")
	}
	mode := os.Getenv("BUBBLY_HELPER_MODE")
	filename := os.Getenv("BUBBLY_HELPER_DB_FILE")
	walFilename := os.Getenv("BUBBLY_HELPER_WAL_FILE")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		fmt.Println("ERR open:", err)
		os.Exit(1)
	}

	switch mode {
	case "insert":
		for i := 0; ; i++ {
			key := fmt.Sprintf("crash_key_%05d", i)
			if err := db.AddToPage(key, crashValueFor(key)); err != nil {
				fmt.Println("ERR add:", err)
				os.Exit(1)
			}
			fmt.Println("OK", key)
		}
	case "delete":
		for i := 0; ; i++ {
			key := fmt.Sprintf("crash_key_%05d", i)
			if err := db.Delete(key); err != nil {
				fmt.Println("ERR delete:", err)
				os.Exit(1)
			}
			fmt.Println("OK", key)
		}
	default:
		fmt.Println("ERR unknown mode:", mode)
		os.Exit(1)
	}
}

func crashValueFor(key string) string {
	return strings.Repeat(key+"|", 8)
}

func isWellFormedCrashKey(key string) bool {
	const prefix = "crash_key_"
	if !strings.HasPrefix(key, prefix) {
		return false
	}
	suffix := key[len(prefix):]
	if len(suffix) != 5 {
		return false
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func runAndKillMidway(t *testing.T, mode, filename, walFilename string, maxOps int) []string {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCrashHelperEntrypoint$")
	cmd.Env = append(os.Environ(),
		"BUBBLY_HELPER_PROCESS=1",
		"BUBBLY_HELPER_MODE="+mode,
		"BUBBLY_HELPER_DB_FILE="+filename,
		"BUBBLY_HELPER_WAL_FILE="+walFilename,
	)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper subprocess: %v", err)
	}

	killAfter := rand.IntN(maxOps) + 1
	var confirmed []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		key, ok := strings.CutPrefix(scanner.Text(), "OK ")
		if !ok {
			continue
		}
		confirmed = append(confirmed, key)
		if len(confirmed) >= killAfter {
			break
		}
	}

	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	return confirmed
}

func TestCrashMidInsert(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-crash-test.db")
	walFilename := filepath.Join(dir, "bubbly-crash-wal-test.db")
	const rounds = 8

	for round := 0; round < rounds; round++ {
		confirmed := runAndKillMidway(t, "insert", filename, walFilename, 60)
		t.Logf("round %d: helper confirmed ~%d inserts before being killed", round, len(confirmed))

		db, err := openDefault(filename, walFilename)
		if err != nil {
			t.Fatalf("round %d: reopen after simulated crash failed: %v", round, err)
		}
		got, err := db.SelectAll()
		if err != nil {
			t.Fatalf("round %d: SelectAll after crash failed: %v", round, err)
		}
		for key, val := range got {
			if !isWellFormedCrashKey(key) {
				t.Errorf("round %d: recovered malformed key %q -> %q", round, key, val)
				continue
			}
			if want := crashValueFor(key); val != want {
				t.Errorf("round %d: corrupted value for %q: got %q want %q", round, key, val, want)
			}
		}
		if err := db.Close(); err != nil {
			t.Fatalf("round %d: close failed: %v", round, err)
		}
		os.Remove(filename)
		os.Remove(walFilename)
	}
}

func TestCrashMidDelete(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-crash-test.db")
	walFilename := filepath.Join(dir, "bubbly-crash-wal-test.db")
	const seedCount = 200
	const rounds = 8

	for round := 0; round < rounds; round++ {
		seedDB, err := openDefault(filename, walFilename)
		if err != nil {
			t.Fatalf("round %d: seed open failed: %v", round, err)
		}
		for i := 0; i < seedCount; i++ {
			key := fmt.Sprintf("crash_key_%05d", i)
			if err := seedDB.AddToPage(key, crashValueFor(key)); err != nil {
				t.Fatalf("round %d: seed insert failed: %v", round, err)
			}
		}
		if err := seedDB.Close(); err != nil {
			t.Fatalf("round %d: seed close failed: %v", round, err)
		}

		confirmed := runAndKillMidway(t, "delete", filename, walFilename, seedCount/2)
		t.Logf("round %d: helper confirmed ~%d deletes before being killed", round, len(confirmed))

		db, err := openDefault(filename, walFilename)
		if err != nil {
			t.Fatalf("round %d: reopen after simulated crash failed: %v", round, err)
		}
		got, err := db.SelectAll()
		if err != nil {
			t.Fatalf("round %d: SelectAll after crash failed: %v", round, err)
		}
		for key, val := range got {
			if !isWellFormedCrashKey(key) {
				t.Errorf("round %d: recovered malformed key %q -> %q after delete-crash", round, key, val)
			}
		}
		for i := 0; i < seedCount; i++ {
			key := fmt.Sprintf("crash_key_%05d", i)
			val, present := got[key]
			if !present {
				continue
			}
			if want := crashValueFor(key); val != want {
				t.Errorf("round %d: surviving key %q is corrupted: got %q want %q", round, key, val, want)
			}
		}
		if err := db.Close(); err != nil {
			t.Fatalf("round %d: close failed: %v", round, err)
		}
		os.Remove(filename)
		os.Remove(walFilename)
	}
}

func TestTransaction(t *testing.T) {
	amount := 100
	var insertedKeys []string
	expected := make(map[string]string)

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	walFilename := filepath.Join(dir, "bubbly-wal-test.db")

	db, err := openDefault(filename, walFilename)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	indices := make([]int, amount)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle them
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	if err := db.BeginTX(); err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Insert in random order, all within one transaction
	for _, i := range indices {
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		db.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
		expected[key] = val
	}

	if err := db.EndTX(); err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	// Verify everything made it to disk after commit
	got, err := db.SelectAll()
	if err != nil {
		t.Fatalf("failed to select all: %v", err)
	}

	if len(got) != len(expected) {
		t.Errorf("expected %d keys after commit, got %d", len(expected), len(got))
	}

	for key, wantVal := range expected {
		gotVal, ok := got[key]
		if !ok {
			t.Errorf("key %q missing after transaction commit", key)
			continue
		}
		if gotVal != wantVal {
			t.Errorf("key %q: expected value %q, got %q", key, wantVal, gotVal)
		}
	}

	// Now delete everything inside a second transaction
	rand.Shuffle(len(insertedKeys), func(i, j int) {
		insertedKeys[i], insertedKeys[j] = insertedKeys[j], insertedKeys[i]
	})

	if err := db.BeginTX(); err != nil {
		t.Fatalf("failed to begin delete transaction: %v", err)
	}

	for _, key := range insertedKeys {
		if err := db.Delete(key); err != nil {
			t.Errorf("failed to delete %q: %v", key, err)
		}
	}

	if err := db.EndTX(); err != nil {
		t.Fatalf("failed to commit delete transaction: %v", err)
	}

	got, err = db.SelectAll()
	if err != nil {
		t.Fatalf("failed to select all after deletes: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 keys after deleting everything, got %d", len(got))
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}
