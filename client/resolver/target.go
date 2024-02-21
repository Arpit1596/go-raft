package resolver

import (
	"strings"

	"google.golang.org/grpc/resolver"
)

func split2(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return "", "", false
	}
	return spl[0], spl[1], true
}

func ParseTarget(target string) (ret resolver.Target) {
	var ok bool
	ret.URL.Scheme, ret.URL.Host, ok = split2(target, ":///")
	if !ok {
		ret.URL.Host = target	
		return ret
	}
	return ret
}
