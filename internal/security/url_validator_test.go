package security

import (
	"context"
	"net/netip"
	"strings"
	"testing"
)

func TestIsBlockedAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		address string
		blocked bool
	}{
		{
			name:    "public IPv4",
			address: "8.8.8.8",
			blocked: false,
		},
		{
			name:    "public IPv6",
			address: "2606:4700:4700::1111",
			blocked: false,
		},
		{
			name:    "unspecified IPv4",
			address: "0.0.0.0",
			blocked: true,
		},
		{
			name:    "loopback IPv4",
			address: "127.0.0.1",
			blocked: true,
		},
		{
			name:    "private class A",
			address: "10.10.10.10",
			blocked: true,
		},
		{
			name:    "private class B",
			address: "172.16.20.30",
			blocked: true,
		},
		{
			name:    "private class C",
			address: "192.168.1.10",
			blocked: true,
		},
		{
			name:    "carrier grade NAT",
			address: "100.64.0.1",
			blocked: true,
		},
		{
			name:    "cloud metadata IPv4",
			address: "169.254.169.254",
			blocked: true,
		},
		{
			name:    "documentation IPv4",
			address: "203.0.113.10",
			blocked: true,
		},
		{
			name:    "multicast IPv4",
			address: "224.0.0.1",
			blocked: true,
		},
		{
			name:    "loopback IPv6",
			address: "::1",
			blocked: true,
		},
		{
			name:    "private IPv6",
			address: "fd00::1",
			blocked: true,
		},
		{
			name:    "link local IPv6",
			address: "fe80::1",
			blocked: true,
		},
		{
			name:    "multicast IPv6",
			address: "ff02::1",
			blocked: true,
		},
		{
			name:    "IPv4 mapped loopback",
			address: "::ffff:127.0.0.1",
			blocked: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				address := netip.MustParseAddr(
					testCase.address,
				)

				actual := IsBlockedAddress(address)

				if actual != testCase.blocked {
					t.Fatalf(
						"IsBlockedAddress(%q) = %v, diharapkan %v",
						testCase.address,
						actual,
						testCase.blocked,
					)
				}
			},
		)
	}
}

func TestValidatePublicHTTPURLRejectsUnsafeInput(
	t *testing.T,
) {
	t.Parallel()

	testCases := []struct {
		name        string
		rawURL      string
		errorString string
	}{
		{
			name:        "empty URL",
			rawURL:      "",
			errorString: "URL wajib diisi",
		},
		{
			name:        "invalid URL",
			rawURL:      "bukan sebuah url",
			errorString: "format URL tidak valid",
		},
		{
			name:        "unsupported scheme",
			rawURL:      "ftp://8.8.8.8/file.txt",
			errorString: "scheme URL hanya boleh http atau https",
		},
		{
			name:        "URL credentials",
			rawURL:      "https://user:password@8.8.8.8/",
			errorString: "URL tidak boleh mengandung kredensial",
		},
		{
			name:        "localhost hostname",
			rawURL:      "http://localhost/",
			errorString: "hostname tidak boleh dipindai",
		},
		{
			name:        "localhost subdomain",
			rawURL:      "http://service.localhost/",
			errorString: "hostname lokal tidak boleh dipindai",
		},
		{
			name:        "local domain",
			rawURL:      "http://internal.local/",
			errorString: "hostname lokal tidak boleh dipindai",
		},
		{
			name:        "loopback IPv4",
			rawURL:      "http://127.0.0.1/",
			errorString: "hostname mengarah ke alamat IP yang tidak diizinkan",
		},
		{
			name:        "private IPv4",
			rawURL:      "http://192.168.1.1/",
			errorString: "hostname mengarah ke alamat IP yang tidak diizinkan",
		},
		{
			name:        "cloud metadata IPv4",
			rawURL:      "http://169.254.169.254/latest/meta-data/",
			errorString: "hostname mengarah ke alamat IP yang tidak diizinkan",
		},
		{
			name:        "loopback IPv6",
			rawURL:      "http://[::1]/",
			errorString: "hostname mengarah ke alamat IP yang tidak diizinkan",
		},
		{
			name:        "private IPv6",
			rawURL:      "http://[fd00::1]/",
			errorString: "hostname mengarah ke alamat IP yang tidak diizinkan",
		},
		{
			name:        "blocked metadata hostname",
			rawURL:      "http://metadata.google.internal/",
			errorString: "hostname tidak boleh dipindai",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := ValidatePublicHTTPURL(
					context.Background(),
					testCase.rawURL,
				)

				if err == nil {
					t.Fatalf(
						"ValidatePublicHTTPURL(%q) seharusnya menghasilkan error",
						testCase.rawURL,
					)
				}

				if !strings.Contains(
					err.Error(),
					testCase.errorString,
				) {
					t.Fatalf(
						"error = %q, diharapkan mengandung %q",
						err.Error(),
						testCase.errorString,
					)
				}
			},
		)
	}
}

func TestValidatePublicHTTPURLAcceptsPublicIP(
	t *testing.T,
) {
	t.Parallel()

	testCases := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "public IPv4 HTTP",
			rawURL:   "  http://8.8.8.8/health?source=test  ",
			expected: "http://8.8.8.8/health?source=test",
		},
		{
			name:     "public IPv4 HTTPS",
			rawURL:   "https://1.1.1.1/path",
			expected: "https://1.1.1.1/path",
		},
		{
			name:     "public IPv6 HTTPS",
			rawURL:   "https://[2606:4700:4700::1111]/dns-query",
			expected: "https://[2606:4700:4700::1111]/dns-query",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				actual, err := ValidatePublicHTTPURL(
					context.Background(),
					testCase.rawURL,
				)
				if err != nil {
					t.Fatalf(
						"ValidatePublicHTTPURL(%q) menghasilkan error: %v",
						testCase.rawURL,
						err,
					)
				}

				if actual != testCase.expected {
					t.Fatalf(
						"hasil = %q, diharapkan %q",
						actual,
						testCase.expected,
					)
				}
			},
		)
	}
}

func TestValidatePublicHTTPURLRejectsOversizedURL(
	t *testing.T,
) {
	t.Parallel()

	rawURL := "https://8.8.8.8/" +
		strings.Repeat("a", 2049)

	_, err := ValidatePublicHTTPURL(
		context.Background(),
		rawURL,
	)

	if err == nil {
		t.Fatal(
			"URL yang terlalu panjang seharusnya ditolak",
		)
	}

	if !strings.Contains(
		err.Error(),
		"URL maksimal 2048 karakter",
	) {
		t.Fatalf(
			"error = %q, diharapkan error panjang URL",
			err.Error(),
		)
	}
}

func TestResolvePublicAddressesRejectsPrivateIP(
	t *testing.T,
) {
	t.Parallel()

	_, err := ResolvePublicAddresses(
		context.Background(),
		"192.168.10.10",
	)

	if err == nil {
		t.Fatal(
			"alamat IP private seharusnya ditolak",
		)
	}

	if !strings.Contains(
		err.Error(),
		"alamat IP tidak diizinkan",
	) {
		t.Fatalf(
			"error = %q, diharapkan error alamat IP",
			err.Error(),
		)
	}
}
