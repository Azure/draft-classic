package azure

import "testing"

func TestGetRegistryName(t *testing.T) {
	var registryNameTests = []struct {
		in  string
		out string
	}{
		{"acrdevname.azurecr.io", "acrdevname"},
		{"jpalma.azurecr.io", "jpalma"},
		{"usdraftacr.azurecr.io", "usdraftacr"},
	}

	for _, tt := range registryNameTests {
		t.Run(tt.in, func(t *testing.T) {
			actual := getRegistryName(tt.in)
			if actual != tt.out {
				t.Errorf("expected %v but got %v", tt.out, actual)
			}
		})
	}
}
