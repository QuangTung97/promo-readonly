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
		action.getSizeLogFromClient()

		h.sess.addNextCall(func() {
			action.handleMemSizeLogNotExisted()
		})
	} else {
		sizeLog := int(sizeLogNum)
		action.sizeLog = sizeLog

		action.getSizeLogFromClient()
		action.getBuckets()
		h.sess.addNextCall(func() {
			action.handleMemSizeLogExisted()
		})
	}

	return func() ([]Entry, error) {
		h.sess.processAllCalls()
		return action.results, action.err
	}
}

// InvalidateSizeLog ...
func (h *hashImpl) InvalidateSizeLog(_ context.Context) func() error {
	return func() error {
		return nil
	}
}

// InvalidateEntry ...
func (h *hashImpl) InvalidateEntry(_ context.Context, _ uint64, _ uint32) func() error {
	return func() error {
		return nil
	}
}
