package address

import (
	"hash/fnv"
	"math/big"
	"math/rand"
	"net"
)

// MaskSize returns the number of IP addresses in the network.
func MaskSize(m *net.IPMask) big.Int {
	var size big.Int

	maskBits, totalBits := m.Size()
	addrBits := totalBits - maskBits
	size.Lsh(big.NewInt(1), uint(addrBits))

	return size
}

// RandomIPv6 returns a random IPv6 address in the given network.
func RandomIPv6(c *net.IPNet, seed ...rand.Rand) net.IP {
	ip := make(net.IP, len(c.IP))

	copy(ip, c.IP)

	for i := len(ip) - 1; i >= 0; i-- {
		if c.Mask[i] == 0xff {
			ip[i] = c.IP[i]
		} else if c.Mask[i] == 0 {
			if len(seed) > 0 {
				ip[i] = byte(seed[0].Intn(256))
			} else {
				ip[i] = byte(rand.Intn(256))
			}
		} else {
			subnetValue := c.IP[i] & c.Mask[i]

			var hostValue byte
			if len(seed) > 0 {
				hostValue = byte(seed[0].Intn(256) & int(^c.Mask[i]))
			} else {
				hostValue = byte(rand.Intn(256) & int(^c.Mask[i]))
			}

			ip[i] = subnetValue | hostValue
		}
	}

	return ip
}

// RandomSeededIPv6 returns a random IPv6 address in the given network, seeded
func RandomSeededIPv6(c *net.IPNet, seed string) net.IP {
	hash := fnv.New32a()
	hash.Write([]byte(seed))
	r := rand.New(rand.NewSource(int64(hash.Sum32())))

	return RandomIPv6(c, *r)
}
