package ipcs

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func IsIpv4(ip string) bool {
	net.ParseIP(ip).To4()
	if net.ParseIP(ip).To4() != nil {
		return true
	}
	return false
}

func Ip2Int(ip string) uint {
	s2ip := net.ParseIP(ip).To4()
	return uint(binary.BigEndian.Uint32(s2ip))
}

func Int2Ip(ipint uint) string {
	ip := net.IP{byte(ipint >> 24), byte(ipint >> 16), byte(ipint >> 8), byte(ipint)}
	return ip.String()
}

func NewIP(ip interface{}) *IP {
	switch ip.(type) {
	case uint, int:
		ipint := uint(ip.(uint))
		return &IP{IP: net.IP{byte(ipint >> 24), byte(ipint >> 16), byte(ipint >> 8), byte(ipint)}.To4()}
	default:
		if i := net.ParseIP(ip.(string)); i != nil {
			return &IP{IP: i.To4()}
		} else {
			return nil
		}

	}
}

// ParseIP parse host to ip and validate ip format
func ParseIP(target string) (*IP, error) {
	target = strings.TrimSpace(target)
	if IsIpv4(target) {
		return &IP{IP: net.ParseIP(target).To4()}, nil
	}

	iprecords, err := net.LookupIP(target)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve domain name:" + target + ". SKIPPED!")
	}

	for _, ip := range iprecords {
		if ip.To4() != nil {
			//Log.Important("parse domain SUCCESS, map " + target + " to " + ip.String())
			return &IP{ip.To4(), target}, nil
		}
	}
	return nil, fmt.Errorf("not found Ip address")
}

type IP struct {
	IP   net.IP
	Host string
}

//func (ip IP) IsIPv4() bool {
//	if ip.IP.To4() != nil {
//		return true
//	}
//	return false
//}

func (ip IP) Int() uint {
	return uint(binary.BigEndian.Uint32(ip.IP.To4()))
}

func (ip IP) String() string {
	return ip.IP.To4().String()
}

func (ip IP) Mask(mask int) IP {
	return IP{IP: ip.IP.Mask(net.CIDRMask(mask, 32))}
}

// NewIPs parse string to ip , auto skip wrong ip
func NewIPs(input []string) IPs {
	ips := make(IPs, len(input))
	for _, ip := range input {
		i, err := ParseIP(ip)
		if err != nil {
			continue
		}
		ips = append(ips, i)
	}
	return ips
}

type IPs []*IP

func (is IPs) Less(i, j int) bool {
	ipi := is[i].Int()
	ipj := is[j].Int()
	if ipi < ipj {
		return true
	} else {
		return false
	}
}

func (is IPs) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

func (is IPs) Len() int {
	return len(is)
}

func (is IPs) Strings() []string {
	s := make([]string, len(is))
	for i, cidr := range is {
		s[i] = cidr.String()
	}
	return s
}

func (is IPs) Approx() CIDRs {
	cidrMap := make(map[string]*CIDR)

	for _, ip := range is {
		fmt.Println(ip.String())
		if n, ok := cidrMap[ip.Mask(24).String()]; ok {
			var baseNet byte
			var nowN, newN byte
			for i := 8; i > 0; i-- {
				nowN = n.IP.IP[3] & (1 << uint(i-1)) >> uint(i-1)
				newN = ip.IP[3] & (1 << uint(i-1)) >> uint(i-1)
				if nowN&newN == 1 {
					baseNet += 1 << uint(i-1)
				}
				if nowN^newN == 1 {
					n.Mask = 32 - i
					n.IP.IP[3] = baseNet
					break
				}
			}
		} else {
			cidrMap[ip.Mask(24).String()] = &CIDR{ip, 32}
		}
	}

	approxed := make(CIDRs, len(cidrMap))
	var index int
	for _, cidr := range cidrMap {
		approxed[index] = cidr
		index++
	}

	return approxed
}
