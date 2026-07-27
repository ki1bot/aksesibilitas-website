package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func ValidatePublicHTTPURL(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("URL wajib diisi")
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", errors.New("format URL tidak valid")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("scheme URL hanya boleh http atau https")
	}

	if parsedURL.Host == "" {
		return "", errors.New("hostname URL wajib diisi")
	}

	if parsedURL.User != nil {
		return "", errors.New("URL tidak boleh mengandung kredensial")
	}

	hostname := strings.TrimSuffix(
		strings.ToLower(parsedURL.Hostname()),
		".",
	)

	if hostname == "" {
		return "", errors.New("hostname URL tidak valid")
	}

	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return "", errors.New("localhost tidak boleh dipindai")
	}

	addresses, err := resolveAddresses(ctx, hostname)
	if err != nil {
		return "", err
	}

	if len(addresses) == 0 {
		return "", errors.New("hostname tidak menghasilkan alamat IP")
	}

	for _, address := range addresses {
		if isBlockedAddress(address) {
			return "", fmt.Errorf(
				"hostname mengarah ke alamat IP yang tidak diizinkan: %s",
				address.String(),
			)
		}
	}

	parsedURL.Fragment = ""

	return parsedURL.String(), nil
}

func resolveAddresses(ctx context.Context, hostname string) ([]netip.Addr, error) {
	if address, err := netip.ParseAddr(hostname); err == nil {
		return []netip.Addr{address}, nil
	}

	lookupContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addresses, err := net.DefaultResolver.LookupNetIP(
		lookupContext,
		"ip",
		hostname,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan resolusi DNS: %w", err)
	}

	return addresses, nil
}

func isBlockedAddress(address netip.Addr) bool {
	address = address.Unmap()

	if !address.IsValid() {
		return true
	}

	if !address.IsGlobalUnicast() {
		return true
	}

	if address.IsPrivate() ||
		address.IsLoopback() ||
		address.IsUnspecified() ||
		address.IsMulticast() ||
		address.IsLinkLocalUnicast() ||
		address.IsLinkLocalMulticast() {
		return true
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}
