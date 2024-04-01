package agent

import (
	"context"
	"net"
)

func LookUp(url string) (string, error) {
	ips, err := net.DefaultResolver.LookupIP(context.TODO(), "ip4", url)
	if err != nil {
		return "", err
	}
	return ips[0].String(), nil
}
