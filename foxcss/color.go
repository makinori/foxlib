package foxcss

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
)

var (
	hexToRGBCache = sync.Map{}

	regexpHexColorShort = regexp.MustCompile(
		`(?i)^#?([0-9a-f])([0-9a-f])([0-9a-f])$`,
	)
	regexpHexColorLong = regexp.MustCompile(
		`(?i)^#?([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$`,
	)
)

func HexToRGB(input string) string {
	out, ok := hexToRGBCache.Load(input)
	if ok {
		return out.(string)
	}

	switch len(input) {
	case 3, 4:
		matches := regexpHexColorShort.FindStringSubmatch(input)
		if len(matches) == 0 {
			slog.Warn("invalid hex color " + input)
			return "0,0,0"
		}
		r, _ := strconv.ParseUint(matches[1]+matches[1], 16, 8)
		g, _ := strconv.ParseUint(matches[2]+matches[2], 16, 8)
		b, _ := strconv.ParseUint(matches[3]+matches[3], 16, 8)
		out = fmt.Sprintf("%d,%d,%d", r, g, b)

	case 6, 7:
		matches := regexpHexColorLong.FindStringSubmatch(input)
		if len(matches) == 0 {
			slog.Warn("invalid hex color " + input)
			return "0,0,0"
		}
		r, _ := strconv.ParseUint(matches[1], 16, 8)
		g, _ := strconv.ParseUint(matches[2], 16, 8)
		b, _ := strconv.ParseUint(matches[3], 16, 8)
		out = fmt.Sprintf("%d,%d,%d", r, g, b)
	}

	if out == "" {
		slog.Warn("invalid hex color " + input)
		return "0,0,0"
	}

	hexToRGBCache.Store(input, out)
	return out.(string)
}

func RGBABackground(rgb, alpha string) string {
	return fmt.Sprintf(
		"linear-gradient(0deg,rgba(%s,%s),rgba(%[1]s,%[2]s))",
		rgb, alpha,
	)
}

func HexAlphaBackground(hex, alpha string) string {
	return fmt.Sprintf(
		"linear-gradient(0deg,rgba(%s,%s),rgba(%[1]s,%[2]s))",
		HexToRGB(hex), alpha,
	)
}
