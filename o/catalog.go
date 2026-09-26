package o

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var bundleName = regexp.MustCompile(`[A-Za-z0-9_\-\.\(\)]{6,}\.bundle`)

const suffixLength = len(".bundle")

type CatalogNameMap map[string]string

func CatalogNameMapOf(path string) (CatalogNameMap, error) {
	result := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}

	founds := bundleName.FindAll(data, -1)
	for _, f := range founds {
		a := f[:len(f)-suffixLength]
		l := len(a)
		if l >= 32 {
			result[string(a[l-32:])+"_"+string(a[:min(maxBundleName, l)])] = string(f)
		}
	}

	return result, nil
}

func (m CatalogNameMap) argOf(fp string) string {
	filename := filepath.Base(fp)
	if m != nil && len(filename) > 24 {
		key := filename[:len(filename)-24]
		if v, ok := m[key]; ok {
			return v
		}
	}
	return addrNameFromDisk(filename)
}

func (m CatalogNameMap) NonceOf(fp string) []byte {
	return nonceOf(m.argOf(fp))
}

func addrNameFromDisk(s string) string {
	base := s[:len(s)-suffixLength]
	tmp := strings.SplitN(base, "_", 2)
	rest := tmp[1]
	midIndex := strings.LastIndex(rest, "_")
	mid := rest
	if midIndex != -1 {
		mid = mid[:midIndex]
	}
	return mid + ".bundle"
}
