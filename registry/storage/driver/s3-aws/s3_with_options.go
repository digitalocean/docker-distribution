package s3

import (
	"context"
	"errors"
	"io"
	"strings"

	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

type Option func(s *WithOptions)
type RequestContextCancelledFunc func(driverName string, path string, err error) error

type WithOptions struct {
	s3                    storagedriver.StorageDriver
	checkForContextCancel RequestContextCancelledFunc
}

func NewWithOptions(opts ...Option) (storagedriver.StorageDriver, error) {
	sd := &WithOptions{
		nil,
		checkS3RequestContextCancellation,
	}
	for _, opt := range opts {
		opt(sd)
	}
	if sd.s3 == nil {
		return nil, errors.New("failed creating storage driver with options - no S3 driver set")
	}
	return sd, nil
}

func (w *WithOptions) Name() string {
	return w.s3.Name()
}

func (w *WithOptions) GetContent(ctx context.Context, path string) ([]byte, error) {
	data, err := w.s3.GetContent(ctx, path)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), path, err)
	}
	return data, err
}

func (w *WithOptions) PutContent(ctx context.Context, path string, content []byte) error {
	err := w.s3.PutContent(ctx, path, content)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), path, err)
	}
	return err
}

func (w *WithOptions) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	return w.s3.Reader(ctx, path, offset)
}

func (w *WithOptions) Writer(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error) {
	return w.s3.Writer(ctx, path, append)
}

func (w *WithOptions) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	info, err := w.s3.Stat(ctx, path)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), path, err)
	}
	return info, err
}

func (w *WithOptions) List(ctx context.Context, path string) ([]string, error) {
	listRes, err := w.s3.List(ctx, path)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), path, err)
	}
	return listRes, err
}

func (w *WithOptions) Move(ctx context.Context, sourcePath string, destPath string) error {
	err := w.s3.Move(ctx, sourcePath, destPath)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), sourcePath, err)
	}
	return err
}

func (w *WithOptions) Delete(ctx context.Context, path string) error {
	err := w.s3.Delete(ctx, path)
	if err != nil {
		err = w.checkForContextCancel(w.Name(), path, err)
	}
	return err
}

func (w *WithOptions) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return w.s3.URLFor(ctx, path, options)
}

func (w *WithOptions) Walk(ctx context.Context, path string, f storagedriver.WalkFn) error {
	return w.s3.Walk(ctx, path, f)
}

func WithWrappedS3Driver(s storagedriver.StorageDriver) func(d *WithOptions) {
	return func(d *WithOptions) {
		d.s3 = s
	}
}

func WithRequestContextCancelled(f RequestContextCancelledFunc) func(d *WithOptions) {
	return func(d *WithOptions) {
		d.checkForContextCancel = f
	}
}

func checkS3RequestContextCancellation(driverName string, path string, err error) error {
	if strings.Contains(err.Error(), "RequestCanceled") {
		return &storagedriver.RequestContextCancelledError{
			DriverName: driverName,
			Path:       path,
			StatusCode: 499,
			Enclosed:   err,
		}
	}
	return err
}
