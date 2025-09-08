package pagination

import (
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/stretchr/testify/assert"
)

type testDomainType struct {
	Name string
	Idx  int // field used for pagination
}

type testDomainTypeCursor struct {
	Idx int
}

func TestHandle(t *testing.T) {
	itemA := testDomainType{Name: "item", Idx: 4}
	itemB := testDomainType{Name: "item", Idx: 9}
	itemC := testDomainType{Name: "item", Idx: 23}
	cursorA, _ := toBase64JSONString(transform(itemA))
	cursorB, _ := toBase64JSONString(transform(itemB))
	cursorC, _ := toBase64JSONString(transform(itemC))
	tests := []struct {
		name            string
		cursorStr       string
		pageSize        int
		wantItems       []testDomainType
		wantPaging      model.Paging
		shouldFetchFail bool
		wantErr         bool
	}{
		{
			name:      "first page, no cursor, has next page",
			cursorStr: "",
			pageSize:  5,
			wantItems: []testDomainType{
				{Name: "item", Idx: 0},
				{Name: "item", Idx: 1},
				{Name: "item", Idx: 2},
				{Name: "item", Idx: 3},
				itemA,
			},
			wantPaging: model.Paging{
				NextCursor: cursorA,
			},
			shouldFetchFail: false,
			wantErr:         false,
		},
		{
			name:      "second page, with cursor, starts from next item, has next page",
			cursorStr: cursorA,
			pageSize:  5,
			wantItems: []testDomainType{
				{Name: "item", Idx: 5},
				{Name: "item", Idx: 6},
				{Name: "item", Idx: 7},
				{Name: "item", Idx: 8},
				itemB,
			},
			wantPaging: model.Paging{
				NextCursor: cursorB,
			},
			shouldFetchFail: false,
			wantErr:         false,
		},
		{
			name:      "last page, with cursor, no next page",
			cursorStr: cursorC,
			pageSize:  5,
			wantItems: []testDomainType{
				{Name: "item", Idx: 24},
				{Name: "item", Idx: 25},
			},
			wantPaging:      model.Paging{},
			shouldFetchFail: false,
			wantErr:         false,
		},
		{
			name:            "invalid cursor returns error",
			cursorStr:       "invalid_base64",
			pageSize:        5,
			wantItems:       nil,
			wantPaging:      model.Paging{},
			shouldFetchFail: false,
			wantErr:         true,
		},
		{
			name:            "fetch failure returns error",
			cursorStr:       "",
			pageSize:        5,
			wantItems:       nil,
			wantPaging:      model.Paging{},
			shouldFetchFail: true,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, gotPaging, err := Handle(
				tt.cursorStr,
				tt.pageSize,
				transform,
				func(cursor *testDomainTypeCursor, pageSize int) ([]testDomainType, error) {
					return mockFetch(tt.shouldFetchFail, cursor, pageSize)
				},
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantItems, gotItems)
				assert.Equal(t, tt.wantPaging, gotPaging)
			}
		})
	}
}

func TestHandle_ToBase64Error(t *testing.T) {
	transform := func(item testDomainType) chan int {
		return make(chan int)
	}
	mockFetch := func(cursor *(chan int), pageSize int) ([]testDomainType, error) {
		return make([]testDomainType, pageSize+1), nil // ensure there's a next page
	}
	// Handle should try to encode the channel as base64 and return an error
	_, _, err := Handle(
		"",
		5,
		transform,
		mockFetch,
	)
	assert.Error(t, err)
}

// mockFetch simulates fetching ascending item Idx from a data source, with a existing items Idxs 0 to 25.
// It uses the cursor to determine the starting point, and returns up to pageSize items.
// If shouldFail is true, it returns an error instead.
func mockFetch(shouldFail bool, cursor *testDomainTypeCursor, pageSize int) ([]testDomainType, error) {
	if shouldFail {
		return nil, assert.AnError
	}
	const maxItemIdx = 25 // will not return items with Idx > 25
	startIdx := 0
	if cursor != nil {
		startIdx = cursor.Idx + 1
	}
	items := []testDomainType{}
	for i := startIdx; i < startIdx+pageSize; i++ {
		if i > maxItemIdx {
			break
		}
		items = append(items, testDomainType{
			Name: "item",
			Idx:  i,
		})
	}
	return items, nil
}

func transform(item testDomainType) testDomainTypeCursor {
	return testDomainTypeCursor{
		Idx: item.Idx,
	}
}

func TestFromBase64JSONString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *testDomainTypeCursor
		wantErr bool
	}{
		{
			name:    "empty string returns nil cursor",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:  "valid base64 JSON string returns correct cursor",
			input: "eyJJZHgiOjl9AA", // base64 of {"Idx":9}
			want: &testDomainTypeCursor{
				Idx: 9,
			},
			wantErr: false,
		},
		{
			name:    "invalid base64 string returns error",
			input:   "invalid_base64",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid base64 but invalid JSON returns error",
			input:   "aW52YWxpZF9qc29u", // base64 of "invalid_json"
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid base64 JSON but wrong type returns error",
			input:   "eyJJZHgiOjEsIk5hbWUiOiJUZXN0In0", // base64 of {"Idx":1,"Name":"Test"}
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromBase64JSONString[testDomainTypeCursor](tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestToBase64JSONString(t *testing.T) {
	input := testDomainTypeCursor{
		Idx: 5,
	}
	want := "eyJJZHgiOjV9"
	got, err := toBase64JSONString(input)

	assert.NoError(t, err)
	assert.Equal(t, want, got)

	ch := make(chan int)
	_, err = toBase64JSONString(ch)
	assert.Error(t, err)
}
