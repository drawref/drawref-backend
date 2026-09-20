package drawref

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCleanSessionTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "null",
			input:    "null",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "empty json object",
			input:    "{}",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "empty json array",
			input:    "[]",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "all empty tag arrays",
			input:    `{"bodies":[],"clothing":[],"energy":[],"cross_contours":[],"mode":[]}`,
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:  "only mode tag specified",
			input: `{"bodies":[],"clothing":[],"energy":[],"cross_contours":[],"mode":["Suggestive"]}`,
			expected: map[string]interface{}{
				"mode": []interface{}{"Suggestive"},
			},
			wantErr: false,
		},
		{
			name:  "multiple populated tags with empty tags filtered out",
			input: `{"bodies":["Women"],"clothing":[],"energy":["Action"],"cross_contours":[],"mode":["Suggestive"]}`,
			expected: map[string]interface{}{
				"bodies": []interface{}{"Women"},
				"energy": []interface{}{"Action"},
				"mode":   []interface{}{"Suggestive"},
			},
			wantErr: false,
		},
		{
			name:     "nil values",
			input:    `{"bodies":null,"mode":[]}`,
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "invalid json",
			input:    `{invalid_json}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanSessionTags(json.RawMessage(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("CleanSessionTags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Fatalf("Failed to unmarshal output: %v", err)
			}

			if len(gotMap) == 0 && len(tt.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(gotMap, tt.expected) {
				t.Errorf("CleanSessionTags() = %v, want %v", gotMap, tt.expected)
			}
		})
	}
}
