package cli

import (
	"io"

	render "github.com/tamnd/twitch-cli/pkg"
)

type Format = render.Format

const (
	FormatTable = render.FormatTable
	FormatJSON  = render.FormatJSON
	FormatJSONL = render.FormatJSONL
	FormatCSV   = render.FormatCSV
	FormatTSV   = render.FormatTSV
	FormatURL   = render.FormatURL
	FormatRaw   = render.FormatRaw
)

func NewRenderer(w io.Writer, format Format, fields []string, noHeader bool, tmpl string) *render.Renderer {
	return render.New(w, format, fields, noHeader, tmpl)
}
