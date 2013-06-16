package preseeder

import (
	"os/exec"
	"regexp"
	"strings"
)

var macAddressPat *regexp.Regexp

func init() {
	macAddressPat = regexp.MustCompile(
		`[0-9a-fA-F]{1,2}:[0-9a-fA-F]{1,2}:[0-9a-fA-F]{1,2}:` +
			`[0-9a-fA-F]{1,2}:[0-9a-fA-F]{1,2}:[0-9a-fA-F]{1,2}`)
}

func arpCommandMAC(addr string) string {
	out, err := exec.Command("arp", "-n", addr).Output()
	// log.Printf("arp: %s", out)
	if err != nil {
		return ""
	}
	rawAddr := macAddressPat.FindString(string(out))
	if rawAddr == "" {
		return ""
	}

	fragments := strings.Split(rawAddr, ":")
	for index, frag := range fragments {
		if len(frag) == 1 {
			fragments[index] = "0" + frag
		}
	}
	return strings.Join(fragments, ":")
}

func getRemoteMacAddress(addr string) string {
	return arpCommandMAC(addr)
	//arpF, err := os.Open("/proc/net/arp")
	//if err == nil {
	//	return procMAC(arpF, addr)
	//} else {
	//	return arpCommandMAC(addr)
	//}
}
