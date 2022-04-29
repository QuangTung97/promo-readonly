package dhash

import "context"

type hashImpl struct {
	sess *sessionImpl

	mem        MemTable
	pipeline   CachePipeline
	db         HashDatabase
	namespace  string
	sizeLogKey string
}

// SelectEntries ...
func (h *hashImpl) SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error) {
	action := &hashSelectAction{
		root: h,
		ctx:  ctx,
		hash: hash,
	}

	sizeLogNum, ok := h.mem.GetNum(h.namespace)
	if !ok {
		action.handleMemSizeLogNotExisted()
	} else {
		sizeLog := int(sizeLogNum)
		action.sizeLog = sizeLog

		action.handleMemSizeLogExisted()
	}

	return func() ([]Entry, error) {
		h.sess.processAllCalls()
		return action.results, action.err
	}
}

// InvalidateSizeLog ...
func (h *hashImpl) InvalidateSizeLog(_ context.Context) func() error {
	return h.pipeline.Delete(h.sizeLogKey)
}

// InvalidateEntry ...
func (h *hashImpl) InvalidateEntry(_ context.Context, sizeLog uint64, hash uint32) func() error {
	fn1 := h.pipeline.Delete(computeBucketKey(h.namespace, int(sizeLog-1), hash))
	fn2 := h.pipeline.Delete(computeBucketKey(h.namespace, int(sizeLog), hash))

	return func() error {
		if err := fn1(); err != nil {
			return err
		}
		return fn2()
	}
}
