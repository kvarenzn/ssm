// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package locale

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func GetSystemLocale() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsLocale()
	case "darwin":
		return getMacLocale()
	case "linux":
		return getLinuxLocale()
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

func getWindowsLocale() (string, error) {
	cmd := exec.Command("powershell", "Get-Culture | Select -ExpandProperty Name")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	lang := os.Getenv("LANG")
	if lang != "" {
		return lang, nil
	}

	return "ja-JP", nil
}

func getMacLocale() (string, error) {
	cmd := exec.Command("defaults", "read", "-g", "AppleLocale")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("locale", "-a")
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
		locales := strings.Split(string(output), "\n")
		if len(locales) > 0 {
			return locales[0], nil
		}
		return "", fmt.Errorf("no locale found")
	}
	return strings.TrimSpace(string(output)), nil
}

func getLinuxLocale() (string, error) {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang != "" {
		parts := strings.Split(lang, ".")
		if len(parts) > 0 {
			return parts[0], nil
		}
		return lang, nil
	}

	cmd := exec.Command("locale")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.HasPrefix(line, "LANG=") || strings.HasPrefix(line, "LANGUAGE=") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				langPart := parts[1]
				if strings.Contains(langPart, ":") {
					langPart = strings.Split(langPart, ":")[0]
				}
				if strings.Contains(langPart, ".") {
					langPart = strings.Split(langPart, ".")[0]
				}
				return langPart, nil
			}
		}
	}

	return "ja_JP", nil
}
