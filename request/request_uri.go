package request

import (
	"fmt"
	"path"
	"strings"
)

func MakeRequestURI(prefix, p string) (string, error) {
	if !strings.HasPrefix(prefix, "/") {
		return "", fmt.Errorf("prefix must starts with \"/\"")
	}
	var uri string
	if p == "" {
		uri = path.Join(prefix+"a", "")
	} else {
		uri = path.Join(prefix, p+"a")
	}
	return uri[:len(uri)-1], nil
}
