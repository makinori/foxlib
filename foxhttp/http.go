package foxhttp

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
)

var ignoreEncodingPrefixes = []string{
	"image/png",
	"image/jpg",
	"image/jpeg",
	"video/",
}

var (
	// go tool air proxy wont work if encoding
	DisableContentEncodingForHTML = false

	ReportWarnings = false

	portRegexp = regexp.MustCompile(":[0-9]+$")
)

func InCommaSeperated(commaSeparated string, needle string) bool {
	if commaSeparated == "" {
		return needle == ""
	}
	for v := range strings.SplitSeq(commaSeparated, ",") {
		if needle == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}

func ServeOptimized(
	w http.ResponseWriter, r *http.Request,
	filename string, modTime time.Time, data []byte,
	allowCache bool,
) {
	// incase it was already set
	contentType := w.Header().Get("Content-Type")

	if allowCache {
		// found in http.ServeContent
		if !(modTime.IsZero() || modTime.Equal(time.Unix(0, 0))) {
			w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
		}

		// unset content type incase etag matches
		w.Header().Del("Content-Type")

		// this better be fast cause we're doing this for each request
		// including ones that have Content-Range
		// idk this is a dumb idea we should cache more
		// but stateless uoohh

		// etag := fmt.Sprintf(`W/"%x"`, xxhash.Sum64(data))
		etag := fmt.Sprintf(`"%x"`, xxhash.Sum64(data))

		ifMatch := r.Header.Get("If-Match")
		if ifMatch != "" {
			if !InCommaSeperated(ifMatch, etag) && !InCommaSeperated(ifMatch, "*") {
				w.WriteHeader(http.StatusPreconditionFailed)
				return
			}
		}

		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" {
			if InCommaSeperated(ifNoneMatch, etag) || InCommaSeperated(ifNoneMatch, "*") {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Header().Add("ETag", etag)
	} else {
		w.Header().Add("Cache-Control", "no-store")
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}
	}

	w.Header().Add("Content-Type", contentType)

	// rest is encoding related

	disableContentEncoding := false

	if DisableContentEncodingForHTML && strings.HasPrefix(contentType, "text/html") {
		disableContentEncoding = true
	} else {
		for _, ignoreEncodingPrefix := range ignoreEncodingPrefixes {
			if strings.HasPrefix(contentType, ignoreEncodingPrefix) {
				disableContentEncoding = true
				break
			}
		}

	}

	if disableContentEncoding {
		http.ServeContent(w, r, filename, modTime, bytes.NewReader(data))
		return
	}

	var err error
	var compressed []byte
	contentEncoding := ""

	acceptEncoding := r.Header.Get("Accept-Encoding")

	if strings.Contains(acceptEncoding, "zstd") {
		contentEncoding = "zstd"
		compressed, err = EncodeZstd(data)
	} else if strings.Contains(acceptEncoding, "br") {
		contentEncoding = "br"
		compressed, err = EncodeBrotli(data)
	}

	if err != nil {
		slog.Error("failed to encode", "name", filename, "err", err.Error())
		w.Write(data)
		return
	}

	if contentEncoding == "" || len(compressed) == 0 {
		w.Write(data)
		return
	}

	if len(compressed) < len(data) {
		w.Header().Add("Content-Encoding", contentEncoding)
		w.Write(compressed)
		return
	}

	if ReportWarnings {
		slog.Warn(
			"ineffecient compression!", "name", filename,
			"type", contentType,
		)
	}

	// w.Write(data)
	http.ServeContent(w, r, filename, modTime, bytes.NewReader(data))
}

// example usage: `http.HandleFunc("GET /{file...}", foxhttp.FileServerOptimized(publicFS))`
func FileServerOptimized(
	fs fs.FS, notFoundHandler ...func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("file")

		file, err := fs.Open(filename)
		if err != nil {
			if len(notFoundHandler) > 0 {
				notFoundHandler[0](w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		modTime := time.Unix(0, 0)
		stat, err := file.Stat()
		if err == nil {
			modTime = stat.ModTime()
		}

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("failed to read file", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ServeOptimized(w, r, filename, modTime, data, true)
	}
}

func GetIPAddress(r *http.Request) string {
	ipAddress := r.Header.Get("X-Real-IP")

	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	}

	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	ipAddress = strings.Split(ipAddress, ",")[0]
	ipAddress = strings.TrimSpace(ipAddress)

	ipAddress = portRegexp.ReplaceAllString(ipAddress, "")

	ipAddress = strings.TrimPrefix(ipAddress, "[")
	ipAddress = strings.TrimSuffix(ipAddress, "]")

	return ipAddress
}

func GetFullURL(r *http.Request) url.URL {
	fullUrl := *r.URL // shallow copy

	fullUrl.Scheme = r.Header.Get("X-Forwarded-Proto")
	if fullUrl.Scheme == "" {
		fullUrl.Scheme = "http"
	}
	fullUrl.Host = r.Host

	return fullUrl
}
