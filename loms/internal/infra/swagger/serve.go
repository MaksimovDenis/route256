package serve

import (
	"context"
	"io"
	"net/http"
	"route256/loms/internal/infra/logger"

	"github.com/rakyll/statik/fs"
)

func SwaggerFile(ctx context.Context, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Infof(ctx, "Serving swagger file: %s", path)

		statikFs, err := fs.New()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Infof(ctx, "Open swagger file: %s", path)

		file, err := statikFs.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		logger.Infof(ctx, "Read swagger file: %s", path)

		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Infof(ctx, "Write swagger file: %s", path)

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Infof(ctx, "Served swagger file: %s", path)
	}
}
