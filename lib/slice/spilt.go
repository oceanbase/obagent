package slice

import "github.com/pkg/errors"

func SpiltBatch(size, batchSize int, batchFn func(start, end int)) error {
	if batchSize <= 0 {
		return errors.Errorf("batchSize %d is invalid", batchSize)
	}
	times := size / batchSize
	for i := 0; i < times; i++ {
		batchFn(i*batchSize, (i+1)*batchSize)
	}
	if size%batchSize > 0 {
		batchFn(size/batchSize*batchSize, size)
	}

	return nil
}
