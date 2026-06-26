package database

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"testing"
)

/*
Note a few of these tests have been write by AI
*/

/*
linked list Once have the range querry in qurrys
*/

func TestOpenClose(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
	}
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database")
	}
}

func TestOneNodeInsert(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
		t.Errorf("failed to close database")
	}
}

func TestInsert(t *testing.T) {
	amount := 2000
	var insertedKeys []string
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
		t.Errorf("failed to close database")
	}
}

func TestInsertWithClose(t *testing.T) {
	amount := 2000
	var insertedKeys []string
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
				t.Errorf("failed to close database")
			}
			db, err = openDefault(filename)
			if err != nil {
				t.Fatal("failed to open database")
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
		t.Errorf("failed to close database")
	}
}

func TestOneNodeDelete(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
		t.Errorf("failed to close database")
	}
}

func TestDelete(t *testing.T) {
	amount := 2000
	var insertedKeys []string
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
		t.Errorf("failed to close database")
	}
}

func TestDeleteWithClose(t *testing.T) {
	amount := 2000
	var insertedKeys []string
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")
	db, err := openDefault(filename)
	if err != nil {
		t.Fatal("failed to open database")
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
				t.Errorf("failed to close database")
			}
			db, err = openDefault(filename)
			if err != nil {
				t.Fatal("failed to open database")
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
				t.Errorf("failed to close database")
			}
			db, err = openDefault(filename)
			if err != nil {
				t.Fatal("failed to open database")
			}
		}
		if err := db.Delete(key); err != nil {
			t.Errorf("failed to delete %q: %v", key, err)
		}
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database")
	}
}

func TestSelectInsertDeleteReopen(t *testing.T) {
	amount := 1000

	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")

	db, err := openDefault(filename)
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

			db, err = openDefault(filename)
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

	db, err = openDefault(filename)
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
		t.Errorf("failed to close database")
	}
}

func TestSelect(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "bubbly-test.db")

	db, err := openDefault(filename)
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
	db, err = openDefault(filename)
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

	db, err := openDefault(filename)
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

	db, err := openDefault(filename)
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

	db, err := openDefault(filename)
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
