package oauth

import "testing"

func TestClaimsSatisfyRequirements(t *testing.T) {
	tests := []struct {
		name     string
		claims   map[string]interface{}
		required map[string][]string
		want     bool
	}{
		{
			name:     "no required claims always passes",
			claims:   map[string]interface{}{},
			required: nil,
			want:     true,
		},
		{
			name:     "missing claim fails",
			claims:   map[string]interface{}{"email": "user@example.com"},
			required: map[string][]string{"groups": {"burrito-admins"}},
			want:     false,
		},
		{
			name:     "string claim matches",
			claims:   map[string]interface{}{"role": "admin"},
			required: map[string][]string{"role": {"admin", "operator"}},
			want:     true,
		},
		{
			name:     "string claim mismatch",
			claims:   map[string]interface{}{"role": "viewer"},
			required: map[string][]string{"role": {"admin", "operator"}},
			want:     false,
		},
		{
			name: "array claim matches one of the allowed values",
			claims: map[string]interface{}{
				"groups": []interface{}{"engineering", "burrito-admins"},
			},
			required: map[string][]string{"groups": {"burrito-admins"}},
			want:     true,
		},
		{
			name: "array claim mismatch",
			claims: map[string]interface{}{
				"groups": []interface{}{"engineering", "sales"},
			},
			required: map[string][]string{"groups": {"burrito-admins"}},
			want:     false,
		},
		{
			name: "multiple required claims must all be satisfied",
			claims: map[string]interface{}{
				"role":   "admin",
				"groups": []interface{}{"engineering"},
			},
			required: map[string][]string{
				"role":   {"admin"},
				"groups": {"burrito-admins"},
			},
			want: false,
		},
		{
			name: "multiple required claims all satisfied",
			claims: map[string]interface{}{
				"role":   "admin",
				"groups": []interface{}{"burrito-admins"},
			},
			required: map[string][]string{
				"role":   {"admin"},
				"groups": {"burrito-admins"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := claimsSatisfyRequirements(tt.claims, tt.required)
			if got != tt.want {
				t.Errorf("claimsSatisfyRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}
