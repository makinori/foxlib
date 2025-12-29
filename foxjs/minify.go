package foxjs

import (
	"bytes"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func Minify(in string) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := js.Minify(minify.New(), buf, bytes.NewReader([]byte(in)), nil)
	return buf.String(), err
}

func MustMinify(in string) string {
	out, err := Minify(in)
	if err != nil {
		panic(err)
	}
	return out
}
