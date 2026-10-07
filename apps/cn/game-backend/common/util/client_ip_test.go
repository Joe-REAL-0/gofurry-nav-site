package util

import "testing"

func TestTrustedClientIP(t *testing.T) {
	for _, tc := range []struct{ name, remote, xff, real, want string }{
		{"untrusted", "192.0.2.1", "198.51.100.2", "", "192.0.2.1"},
		{"trusted", "127.0.0.1", "198.51.100.2, 127.0.0.2", "", "198.51.100.2"},
		{"untrusted prefix", "127.0.0.1", "203.0.113.10, 198.51.100.2", "", "198.51.100.2"},
		{"malformed", "127.0.0.1", "garbage, 198.51.100.2", "203.0.113.2", "127.0.0.1"},
		{"real", "127.0.0.1", "", "198.51.100.2", "198.51.100.2"},
		{"ipv6", "::1", "2001:db8::2", "", "2001:db8::2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveClientIP(tc.remote, tc.xff, tc.real, []string{"127.0.0.0/8", "::1/128"}); got != tc.want {
				t.Fatalf("%s != %s", got, tc.want)
			}
		})
	}
}
