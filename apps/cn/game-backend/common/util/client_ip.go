package util

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"net/netip"
	"strings"
)

// GetClientIP starts from the socket peer, independently of Fiber's proxy header
// configuration. Walk from the nearest proxy toward the client, ignoring any
// attacker-supplied prefix beyond the first untrusted hop.
func GetClientIP(c fiber.Ctx) string {
	return ResolveClientIP(c.RequestCtx().RemoteIP().String(), c.Get("X-Forwarded-For"), c.Get("X-Real-IP"), env.GetServerConfig().Server.TrustedProxyCIDRs)
}
func ResolveClientIP(remote, forwarded, real string, cidrs []string) string {
	peer, e := netip.ParseAddr(strings.TrimSpace(remote))
	if e != nil {
		return ""
	}
	peer = peer.Unmap()
	trusted := func(ip netip.Addr) bool {
		for _, v := range cidrs {
			p, e := netip.ParsePrefix(v)
			if e == nil && p.Contains(ip) {
				return true
			}
		}
		return false
	}
	fallback := peer.String()
	if !trusted(peer) {
		return fallback
	}
	if len(forwarded) > 4096 {
		return fallback
	}
	if strings.TrimSpace(forwarded) == "" {
		if ip, e := netip.ParseAddr(strings.TrimSpace(real)); e == nil && ip.Zone() == "" {
			return ip.Unmap().String()
		}
		return fallback
	}
	parts := strings.Split(forwarded, ",")
	if len(parts) > 32 {
		return fallback
	}
	chain := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		ip, e := netip.ParseAddr(strings.TrimSpace(part))
		if e != nil || ip.Zone() != "" {
			return fallback
		}
		chain = append(chain, ip.Unmap())
	}
	for i := len(chain) - 1; i >= 0; i-- {
		if !trusted(peer) {
			break
		}
		peer = chain[i]
	}
	return peer.String()
}
