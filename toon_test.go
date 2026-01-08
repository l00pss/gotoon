package toon_test

import (
	"strings"
	"testing"

	toon "github.com/l00pss/gotoon"
)

type Context struct {
	Task     string `toon:"task"`
	Location string `toon:"location"`
	Season   string `toon:"season"`
}

type Hike struct {
	ID            int     `toon:"id"`
	Name          string  `toon:"name"`
	DistanceKm    float64 `toon:"distanceKm"`
	ElevationGain int     `toon:"elevationGain"`
	Companion     string  `toon:"companion"`
	WasSunny      bool    `toon:"wasSunny"`
}

type HikesData struct {
	Context Context  `toon:"context"`
	Friends []string `toon:"friends"`
	Hikes   []Hike   `toon:"hikes"`
}

func TestMarshalSimple(t *testing.T) {
	data := struct {
		Name  string `toon:"name"`
		Age   int    `toon:"age"`
		Email string `toon:"email"`
	}{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	result, err := toon.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "name: Alice\nage: 30\nemail: alice@example.com\n"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(result))
	}
}

func TestMarshalNestedStruct(t *testing.T) {
	data := struct {
		Context Context `toon:"context"`
	}{
		Context: Context{
			Task:     "Our favorite hikes together",
			Location: "Boulder",
			Season:   "spring_2025",
		},
	}

	result, err := toon.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "context:\n  task: Our favorite hikes together\n  location: Boulder\n  season: spring_2025\n"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(result))
	}
}

func TestMarshalPrimitiveArray(t *testing.T) {
	data := struct {
		Friends []string `toon:"friends"`
	}{
		Friends: []string{"ana", "luis", "sam"},
	}

	result, err := toon.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "friends[3]: ana,luis,sam\n"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(result))
	}
}

func TestMarshalTabularArray(t *testing.T) {
	data := struct {
		Hikes []Hike `toon:"hikes"`
	}{
		Hikes: []Hike{
			{ID: 1, Name: "Blue Lake Trail", DistanceKm: 7.5, ElevationGain: 320, Companion: "ana", WasSunny: true},
			{ID: 2, Name: "Ridge Overlook", DistanceKm: 9.2, ElevationGain: 540, Companion: "luis", WasSunny: false},
			{ID: 3, Name: "Wildflower Loop", DistanceKm: 5.1, ElevationGain: 180, Companion: "sam", WasSunny: true},
		},
	}

	result, err := toon.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Tabular output:\n%s", string(result))

	if len(result) == 0 {
		t.Error("Expected non-empty output")
	}

	// Check it contains tabular format indicators
	output := string(result)
	if !strings.Contains(output, "hikes[3]") {
		t.Error("Expected array length declaration")
	}
	if !strings.Contains(output, "{id,name,distanceKm,elevationGain,companion,wasSunny}") {
		t.Error("Expected field declaration")
	}
}

func TestMarshalFullExample(t *testing.T) {
	data := HikesData{
		Context: Context{
			Task:     "Our favorite hikes together",
			Location: "Boulder",
			Season:   "spring_2025",
		},
		Friends: []string{"ana", "luis", "sam"},
		Hikes: []Hike{
			{ID: 1, Name: "Blue Lake Trail", DistanceKm: 7.5, ElevationGain: 320, Companion: "ana", WasSunny: true},
			{ID: 2, Name: "Ridge Overlook", DistanceKm: 9.2, ElevationGain: 540, Companion: "luis", WasSunny: false},
			{ID: 3, Name: "Wildflower Loop", DistanceKm: 5.1, ElevationGain: 180, Companion: "sam", WasSunny: true},
		},
	}

	result, err := toon.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Full TOON output:\n%s", string(result))

	if len(result) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestMarshalWithTabDelimiter(t *testing.T) {
	data := struct {
		Numbers []int `toon:"numbers"`
	}{
		Numbers: []int{1, 2, 3, 4, 5},
	}

	opts := toon.MarshalOptions{
		Indent:     2,
		Delimiter:  toon.DelimiterTab,
		UseTabular: true,
	}

	result, err := toon.MarshalWithOptions(data, opts)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "numbers[5]: 1\t2\t3\t4\t5\n"
	if string(result) != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, string(result))
	}
}

func TestUnmarshalSimple(t *testing.T) {
	input := "name: Alice\nage: 30\nemail: alice@example.com\n"

	var result struct {
		Name  string `toon:"name"`
		Age   int    `toon:"age"`
		Email string `toon:"email"`
	}

	if err := toon.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("Expected Name=Alice, got %s", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("Expected Age=30, got %d", result.Age)
	}
	if result.Email != "alice@example.com" {
		t.Errorf("Expected Email=alice@example.com, got %s", result.Email)
	}
}

func TestUnmarshalNested(t *testing.T) {
	input := "context:\n  task: Our favorite hikes together\n  location: Boulder\n  season: spring_2025\n"

	var result struct {
		Context Context `toon:"context"`
	}

	if err := toon.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Context.Task != "Our favorite hikes together" {
		t.Errorf("Expected Task='Our favorite hikes together', got '%s'", result.Context.Task)
	}
	if result.Context.Location != "Boulder" {
		t.Errorf("Expected Location='Boulder', got '%s'", result.Context.Location)
	}
}

func TestUnmarshalPrimitiveArray(t *testing.T) {
	input := "friends[3]: ana,luis,sam\n"

	var result struct {
		Friends []string `toon:"friends"`
	}

	if err := toon.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result.Friends) != 3 {
		t.Errorf("Expected 3 friends, got %d", len(result.Friends))
	}
	expected := []string{"ana", "luis", "sam"}
	for i, f := range expected {
		if i < len(result.Friends) && result.Friends[i] != f {
			t.Errorf("Expected friend[%d]=%s, got %s", i, f, result.Friends[i])
		}
	}
}

func TestUnmarshalTabularArray(t *testing.T) {
	input := `hikes[3]{id,name,distanceKm,elevationGain,companion,wasSunny}:
  1,Blue Lake Trail,7.5,320,ana,true
  2,Ridge Overlook,9.2,540,luis,false
  3,Wildflower Loop,5.1,180,sam,true
`

	var result struct {
		Hikes []Hike `toon:"hikes"`
	}

	if err := toon.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result.Hikes) != 3 {
		t.Errorf("Expected 3 hikes, got %d", len(result.Hikes))
	}

	if len(result.Hikes) > 0 {
		hike := result.Hikes[0]
		if hike.ID != 1 || hike.Name != "Blue Lake Trail" || hike.DistanceKm != 7.5 {
			t.Errorf("First hike incorrect: %+v", hike)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	original := HikesData{
		Context: Context{
			Task:     "Our favorite hikes together",
			Location: "Boulder",
			Season:   "spring_2025",
		},
		Friends: []string{"ana", "luis", "sam"},
		Hikes: []Hike{
			{ID: 1, Name: "Blue Lake Trail", DistanceKm: 7.5, ElevationGain: 320, Companion: "ana", WasSunny: true},
		},
	}

	data, err := toon.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded HikesData
	if err := toon.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Compare basic fields
	if decoded.Context.Task != original.Context.Task {
		t.Errorf("Task mismatch: expected %s, got %s", original.Context.Task, decoded.Context.Task)
	}
	if len(decoded.Friends) != len(original.Friends) {
		t.Errorf("Friends length mismatch: expected %d, got %d", len(original.Friends), len(decoded.Friends))
	}
}

func TestValid(t *testing.T) {
	validToon := "name: Alice\nage: 30\n"
	if !toon.Valid([]byte(validToon)) {
		t.Error("Expected valid TOON to be valid")
	}

	invalidToon := "invalid syntax here"
	if toon.Valid([]byte(invalidToon)) {
		t.Error("Expected invalid TOON to be invalid")
	}
}

func BenchmarkMarshal(b *testing.B) {
	data := HikesData{
		Context: Context{
			Task:     "Our favorite hikes together",
			Location: "Boulder",
			Season:   "spring_2025",
		},
		Friends: []string{"ana", "luis", "sam"},
		Hikes: []Hike{
			{ID: 1, Name: "Blue Lake Trail", DistanceKm: 7.5, ElevationGain: 320, Companion: "ana", WasSunny: true},
			{ID: 2, Name: "Ridge Overlook", DistanceKm: 9.2, ElevationGain: 540, Companion: "luis", WasSunny: false},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = toon.Marshal(data)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	input := []byte("context:\n  task: Our favorite hikes together\n  location: Boulder\nfriends[3]: ana,luis,sam\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result struct {
			Context Context  `toon:"context"`
			Friends []string `toon:"friends"`
		}
		_ = toon.Unmarshal(input, &result)
	}
}
