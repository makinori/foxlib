package foxcss

import "strings"

// replace all & with class name afterwards
func preprocess(input string) (css string) {
	css = "&{"
	outOfMain := false

	for line := range strings.SplitSeq(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// ignore comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// parse nesting
		// can only nest from root though
		if strings.HasSuffix(line, "{") {
			if !outOfMain {
				outOfMain = true
				css += "}"
			}
			if !strings.HasPrefix(line, "@") &&
				!strings.HasPrefix(line, "&") {
				css += "& "
			}
		}

		// each css rule
		if !strings.HasSuffix(line, "}") && !strings.HasSuffix(line, "{") &&
			!strings.HasSuffix(line, ":") && !strings.HasSuffix(line, ",") {
			// make sure ends with semicolon
			if !strings.HasSuffix(line, ";") {
				line += ";"
			}
		}

		css += line
	}

	if !outOfMain {
		css += "}"
	}

	return
}
