package foxhtml

import (
	"context"
	"io"

	"github.com/makinori/foxlib/foxcss"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// nice! if only we can do this from foxcss itself lol
type StackCSS string

func (_ StackCSS) Render(_ io.Writer) error {
	return nil
}

func stack(
	ctx context.Context, flexDir string, children ...Node,
) Node {
	class := foxcss.Class(ctx, `
		display: flex;
		flex-direction: `+flexDir+`;
		gap: 8px;
	`)

	for _, node := range children {
		switch css := node.(type) {
		case StackCSS:
			class += " " + foxcss.Class(ctx, string(css))
		}
	}

	return Div(
		Class(class),
		Group(children),
	)
}
func HStack(ctx context.Context, children ...Node) Node {
	return stack(ctx, "row", children...)
}

func VStack(ctx context.Context, children ...Node) Node {
	return stack(ctx, "column", children...)
}
