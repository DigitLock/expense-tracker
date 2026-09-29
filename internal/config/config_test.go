package config

import "testing"

func TestResolveJWTSecret(t *testing.T) {
	cases := []struct {
		name    string
		secret  string
		appEnv  string
		want    string
		wantErr bool
	}{
		{name: "empty secret, no env", secret: "", appEnv: "", wantErr: true},
		{name: "empty secret, non-dev env", secret: "", appEnv: "staging", wantErr: true},
		{name: "empty secret, dev env", secret: "", appEnv: "dev", want: devJWTSecret},
		{name: "secret set, non-dev env", secret: "s3cret", appEnv: "staging", want: "s3cret"},
		{name: "secret set, dev env", secret: "s3cret", appEnv: "dev", want: "s3cret"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveJWTSecret(tc.secret, tc.appEnv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got secret %q", got)
				}
				if err.Error() != "JWT_SECRET is not set" {
					t.Errorf("error = %q, want %q", err.Error(), "JWT_SECRET is not set")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("secret = %q, want %q", got, tc.want)
			}
		})
	}
}
