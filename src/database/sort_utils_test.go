package database

import (
	"math"
	"testing"
)

// ==================== SortBetween ====================

func TestSortBetween_Normal(t *testing.T) {
	result := SortBetween(10, 20)
	if result != 15 {
		t.Fatalf("expected 15, got %d", result)
	}
}

func TestSortBetween_Adjacent(t *testing.T) {
	result := SortBetween(10, 11)
	if result != 10 {
		t.Fatalf("expected 10 (integer division), got %d", result)
	}
}

func TestSortBetween_SameValues(t *testing.T) {
	result := SortBetween(10, 10)
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

func TestSortBetween_LargeGap(t *testing.T) {
	result := SortBetween(0, 1000)
	if result != 500 {
		t.Fatalf("expected 500, got %d", result)
	}
}

func TestSortBetween_NegativeValues(t *testing.T) {
	result := SortBetween(-20, -10)
	if result != -15 {
		t.Fatalf("expected -15, got %d", result)
	}
}

func TestSortBetween_NoOverflow(t *testing.T) {
	// before + (after-before)/2 should avoid overflow
	result := SortBetween(math.MaxInt-1, math.MaxInt)
	expected := math.MaxInt - 1 // integer division
	if result != expected {
		t.Fatalf("expected %d, got %d", expected, result)
	}
}

// ==================== SortSequence ====================

func TestSortSequence_Zero(t *testing.T) {
	seq := SortSequence(0)
	if len(seq) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(seq))
	}
}

func TestSortSequence_One(t *testing.T) {
	seq := SortSequence(1)
	if len(seq) != 1 {
		t.Fatalf("expected 1 item, got %d", len(seq))
	}
	if seq[0] != SortStep {
		t.Fatalf("expected %d, got %d", SortStep, seq[0])
	}
}

func TestSortSequence_Multiple(t *testing.T) {
	seq := SortSequence(5)
	expected := []int{10, 20, 30, 40, 50}
	if len(seq) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(seq))
	}
	for i, v := range seq {
		if v != expected[i] {
			t.Fatalf("seq[%d]: expected %d, got %d", i, expected[i], v)
		}
	}
}

// ==================== SortStep constant ====================

func TestSortStep_Value(t *testing.T) {
	if SortStep != 10 {
		t.Fatalf("expected SortStep=10, got %d", SortStep)
	}
}

// ==================== SortInsert ====================

func TestSortInsert_BothNil(t *testing.T) {
	result := SortInsert(nil, nil)
	if result != SortStep {
		t.Fatalf("expected %d, got %d", SortStep, result)
	}
}

func TestSortInsert_BeforeNil(t *testing.T) {
	after := 20
	result := SortInsert(nil, &after)
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

func TestSortInsert_AfterNil(t *testing.T) {
	before := 20
	result := SortInsert(&before, nil)
	if result != 30 {
		t.Fatalf("expected 30, got %d", result)
	}
}

func TestSortInsert_BothSet(t *testing.T) {
	before := 10
	after := 30
	result := SortInsert(&before, &after)
	if result != 20 {
		t.Fatalf("expected 20, got %d", result)
	}
}

// ==================== SortInsertSafe ====================

func TestSortInsertSafe_NoConflict(t *testing.T) {
	before := 10
	after := 30
	result, conflict := SortInsertSafe(&before, &after)
	if conflict {
		t.Fatal("expected no conflict")
	}
	if result != 20 {
		t.Fatalf("expected 20, got %d", result)
	}
}

func TestSortInsertSafe_ConflictWithBefore(t *testing.T) {
	// When before and after are adjacent, integer division may cause conflict
	before := 10
	after := 11
	result, conflict := SortInsertSafe(&before, &after)
	// SortBetween(10, 11) = 10 + (11-10)/2 = 10, which equals before → conflict
	if !conflict {
		t.Fatal("expected conflict")
	}
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

func TestSortInsertSafe_ConflictWithAfter(t *testing.T) {
	// before and after very close: SortBetween may equal after
	before := 10
	after := 10
	result, conflict := SortInsertSafe(&before, &after)
	if !conflict {
		t.Fatal("expected conflict")
	}
	if result != 10 {
		t.Fatalf("expected 10, got %d", result)
	}
}

// ==================== toSnakeCase ====================

func TestToSnakeCase_Simple(t *testing.T) {
	result := toSnakeCase("NodeGroup")
	if result != "node_group" {
		t.Fatalf("expected 'node_group', got '%s'", result)
	}
}

func TestToSnakeCase_AllLower(t *testing.T) {
	result := toSnakeCase("profile")
	if result != "profile" {
		t.Fatalf("expected 'profile', got '%s'", result)
	}
}

func TestToSnakeCase_AllUpper(t *testing.T) {
	result := toSnakeCase("ABC")
	if result != "a_b_c" {
		t.Fatalf("expected 'a_b_c', got '%s'", result)
	}
}

func TestToSnakeCase_SingleWord(t *testing.T) {
	result := toSnakeCase("Profile")
	if result != "profile" {
		t.Fatalf("expected 'profile', got '%s'", result)
	}
}

func TestToSnakeCase_MultipleWords(t *testing.T) {
	result := toSnakeCase("RoutingRule")
	if result != "routing_rule" {
		t.Fatalf("expected 'routing_rule', got '%s'", result)
	}
}

func TestToSnakeCase_StrategyGroup(t *testing.T) {
	result := toSnakeCase("StrategyGroup")
	if result != "strategy_group" {
		t.Fatalf("expected 'strategy_group', got '%s'", result)
	}
}

func TestToSnakeCase_AppSetting(t *testing.T) {
	result := toSnakeCase("AppSetting")
	if result != "app_setting" {
		t.Fatalf("expected 'app_setting', got '%s'", result)
	}
}

// ==================== safeAdd / safeSub ====================

func TestSafeAdd_Normal(t *testing.T) {
	result, ok := safeAdd(10, 20)
	if !ok || result != 30 {
		t.Fatalf("expected (30, true), got (%d, %v)", result, ok)
	}
}

func TestSafeAdd_Overflow(t *testing.T) {
	_, ok := safeAdd(math.MaxInt, 1)
	if ok {
		t.Fatal("expected overflow detection")
	}
}

func TestSafeAdd_Underflow(t *testing.T) {
	_, ok := safeAdd(math.MinInt, -1)
	if ok {
		t.Fatal("expected underflow detection")
	}
}

func TestSafeSub_Normal(t *testing.T) {
	result, ok := safeSub(30, 10)
	if !ok || result != 20 {
		t.Fatalf("expected (20, true), got (%d, %v)", result, ok)
	}
}

func TestSafeSub_Overflow(t *testing.T) {
	_, ok := safeSub(math.MaxInt, -1)
	if ok {
		t.Fatal("expected overflow detection")
	}
}

func TestSafeSub_Underflow(t *testing.T) {
	_, ok := safeSub(math.MinInt, 1)
	if ok {
		t.Fatal("expected underflow detection")
	}
}

// ==================== mustAdd / mustSub ====================

func TestMustAdd_Normal(t *testing.T) {
	result := mustAdd(10, 20)
	if result != 30 {
		t.Fatalf("expected 30, got %d", result)
	}
}

func TestMustAdd_Overflow(t *testing.T) {
	result := mustAdd(math.MaxInt, 1)
	if result != 0 {
		t.Fatalf("expected 0 on overflow, got %d", result)
	}
}

func TestMustSub_Normal(t *testing.T) {
	result := mustSub(30, 10)
	if result != 20 {
		t.Fatalf("expected 20, got %d", result)
	}
}

func TestMustSub_Overflow(t *testing.T) {
	result := mustSub(math.MinInt, 1)
	if result != 0 {
		t.Fatalf("expected 0 on overflow, got %d", result)
	}
}
