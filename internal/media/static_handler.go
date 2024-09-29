package media

import (
	"errors"
	"net/http"
	"os"
	"path"
)

func indexHTMLWhenNotFound(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, err := fs.Open(path.Clean(req.URL.Path)) // Do not allow path traversals.
		if errors.Is(err, os.ErrNotExist) {
			http.ServeFile(resp, req, "./web/index.html")
			return
		}
		fileServer.ServeHTTP(resp, req)
	})
}
