package comment

import (
	_ "embed"
)

// func TestDefaultComment_Generate(t *testing.T) {
// 	type fields struct {
// 		layers  []configv1alpha1.TerraformLayer
// 		storage storage.Storage
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		want    string
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &DefaultComment{
// 				layers:  tt.fields.layers,
// 				storage: tt.fields.storage,
// 			}
// 			got, err := c.Generate()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DefaultComment.Generate() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("DefaultComment.Generate() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
