package providers

import "testing"

func TestNewProviderFromEnvironmentRecognizesProviderModes(t *testing.T) {
	testCases := []struct {
		name    string
		mode    string
		want    string
		wantErr string
	}{
		{name: "unset defaults to azure", mode: "", want: "azure"},
		{name: "azure explicit", mode: "azure", want: "azure"},
		{name: "azure trims and normalizes case", mode: " Azure ", want: "azure"},
		{name: "static explicit", mode: "static", want: "static"},
		{name: "static trims and normalizes case", mode: " Static ", want: "static"},
		{name: "whitespace only defaults to azure", mode: "   ", want: "azure"},
		{name: "invalid mode", mode: "bogus", wantErr: `invalid AZUREFOX_PROVIDER value "bogus"; valid values: azure, static`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(providerModeEnv, testCase.mode)

			provider, err := NewProviderFromEnvironment()
			if testCase.wantErr != "" {
				if err == nil {
					t.Fatalf("NewProviderFromEnvironment() error = nil, want %q", testCase.wantErr)
				}
				if err.Error() != testCase.wantErr {
					t.Fatalf("NewProviderFromEnvironment() error = %q, want %q", err.Error(), testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewProviderFromEnvironment() error = %v", err)
			}

			switch testCase.want {
			case "azure":
				if _, ok := provider.(AzureProvider); !ok {
					t.Fatalf("NewProviderFromEnvironment() provider = %T, want providers.AzureProvider", provider)
				}
			case "static":
				if _, ok := provider.(StaticProvider); !ok {
					t.Fatalf("NewProviderFromEnvironment() provider = %T, want providers.StaticProvider", provider)
				}
			default:
				t.Fatalf("test case %q has unknown want %q", testCase.name, testCase.want)
			}
		})
	}
}
