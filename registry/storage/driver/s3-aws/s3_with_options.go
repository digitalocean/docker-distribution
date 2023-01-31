package s3

import (
	"context"
	"io"

	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

type Option func(s *WithOptions)
type RequestContextCancelledFunc func(err error) error

type WithOptions struct {
	s3                    storagedriver.StorageDriver
	checkForContextCancel RequestContextCancelledFunc
}

func NewWithOptions(s3 storagedriver.StorageDriver, opts ...Option) storagedriver.StorageDriver {
	sd := &WithOptions{
		s3,
		checkS3RequestContextCancellation,
	}
	for _, opt := range opts {
		opt(sd)
	}
	return sd
}

func (w *WithOptions) Name() string {
	return w.s3.Name()
}

func (w *WithOptions) GetContent(ctx context.Context, path string) ([]byte, error) {
	data, err := w.s3.GetContent(ctx, path)
	return data, w.checkForContextCancel(err)
}

func (w *WithOptions) PutContent(ctx context.Context, path string, content []byte) error {
	return checkS3RequestContextCancellation(w.s3.PutContent(ctx, path, content))
}

func (w *WithOptions) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	reader, err := w.s3.Reader(ctx, path, offset)
	return reader, checkS3RequestContextCancellation(err)
}

func (w *WithOptions) Writer(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error) {
	writer, err := w.s3.Writer(ctx, path, append)
	return writer, checkS3RequestContextCancellation(err)
}

func (w *WithOptions) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	info, err := w.s3.Stat(ctx, path)
	return info, w.checkForContextCancel(err)
}

func (w *WithOptions) List(ctx context.Context, path string) ([]string, error) {
	listRes, err := w.s3.List(ctx, path)
	return listRes, w.checkForContextCancel(err)
}

func (w *WithOptions) Move(ctx context.Context, sourcePath string, destPath string) error {
	return w.checkForContextCancel(w.s3.Move(ctx, sourcePath, destPath))
}

func (w *WithOptions) Delete(ctx context.Context, path string) error {
	return w.checkForContextCancel(w.s3.Delete(ctx, path))
}

func (w *WithOptions) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return w.s3.URLFor(ctx, path, options)
}

func (w *WithOptions) Walk(ctx context.Context, path string, f storagedriver.WalkFn) error {
	return w.s3.Walk(ctx, path, f)
}

func WithRequestContextCancelled(f RequestContextCancelledFunc) func(d *WithOptions) {
	return func(d *WithOptions) {
		d.checkForContextCancel = f
	}
}

func checkS3RequestContextCancellation(err error) error {
	return err
}
