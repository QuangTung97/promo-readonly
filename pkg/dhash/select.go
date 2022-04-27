package dhash

import (
	"context"
	"fmt"
	"strconv"
)

type hashSelectAction struct {
	root *hashImpl
	ctx  context.Context
	hash uint32

	sizeLogFn      func() (LeaseGetOutput, error)
	bucketFn1      func() (GetOutput, error)
	bucketFn2      func() (GetOutput, error)
	sizeLogDBFn    func() (uint64, error)
	bucketLeaseGet func() (LeaseGetOutput, error)
	entriesDBFn    func() ([]Entry, error)

	sizeLog        int
	sizeLogLeaseID uint64
	bucketLeaseID  uint64

	results []Entry
	err     error
}

func (h *hashSelectAction) getSizeLogFromClient() {
	sizeLogFn := h.root.pipeline.LeaseGet(h.root.sizeLogKey)
	h.sizeLogFn = sizeLogFn
}

func computeBucketKey(ns string, sizeLog int, hash uint32) string {
	return fmt.Sprintf("%s:%d:%x", ns, sizeLog, startOfSlot(hash, sizeLog))
}

func (h *hashSelectAction) getBuckets() {
	key1 := computeBucketKey(h.root.namespace, h.sizeLog-1, h.hash)
	key2 := computeBucketKey(h.root.namespace, h.sizeLog, h.hash)

	h.bucketFn1 = h.root.pipeline.Get(key1)
	h.bucketFn2 = h.root.pipeline.Get(key2)
}

func (h *hashSelectAction) handleMemSizeLogNotExisted() {
	h.err = h.handleMemSizeLogNotExistedWithError()
}

func (h *hashSelectAction) handleSizeLogFromDB() {
	dbSizeLog, err := h.sizeLogDBFn()
	if err != nil {
		h.err = err
		return
	}
	h.root.mem.SetNum(h.root.namespace, dbSizeLog)
	h.sizeLog = int(dbSizeLog)
	h.root.pipeline.LeaseSet(
		h.root.sizeLogKey,
		[]byte(strconv.FormatUint(dbSizeLog, 10)),
		h.sizeLogLeaseID, 0, // TODO Customize TTL
	)
}

func (h *hashSelectAction) handleSizeLogFromClient(callback func()) {
	h.err = h.handleSizeLogFromClientWithError(callback)
}

func (h *hashSelectAction) handleSizeLogFromClientWithError(callback func()) error {
	newSizeLogOutput, err := h.sizeLogFn()
	if err != nil {
		return err
	}

	if newSizeLogOutput.Type == LeaseGetTypeGranted {
		h.sizeLogDBFn = h.root.db.GetSizeLog(h.ctx)
		h.sizeLogLeaseID = newSizeLogOutput.LeaseID
		h.root.sess.addNextCall(func() {
			h.handleSizeLogFromDB()
		})
		callback()
	}

	sizeLog, err := strconv.ParseInt(string(newSizeLogOutput.Data), 10, 64)
	if err != nil {
		return err
	}
	// TODO compare with previous, handle difference
	h.sizeLog = int(sizeLog)

	callback()

	return nil
}

func (h *hashSelectAction) handleMemSizeLogNotExistedWithError() error {
	h.handleSizeLogFromClient(func() {
		h.getBuckets()
		h.root.sess.addNextCall(func() {
			h.handleBuckets()
		})
	})
	return nil
}

func (h *hashSelectAction) handleMemSizeLogExisted() {
	h.handleSizeLogFromClient(func() {
		h.handleBuckets()
	})
}

func (h *hashSelectAction) handleBuckets() {
	h.results, h.err = h.handleBucketsWithOutput()
}

func (h *hashSelectAction) handleBucketsWithOutput() ([]Entry, error) {
	bucket1Output, err := h.bucketFn1()
	if err != nil {
		return nil, err
	}

	var data []byte
	if bucket1Output.Found {
		data = bucket1Output.Data
	}

	bucket2Output, err := h.bucketFn2()
	if err != nil {
		return nil, err
	}
	if bucket2Output.Found {
		data = bucket2Output.Data
	}

	if len(data) == 0 {
		key := computeBucketKey(h.root.namespace, h.sizeLog, h.hash)
		h.bucketLeaseGet = h.root.pipeline.LeaseGet(key)
		h.root.sess.addNextCall(func() {
			h.handleGetBucketFromDB()
		})
		return nil, nil
	}

	entries, err := unmarshalEntries(data)
	if err != nil {
		return nil, err
	}

	var result []Entry
	for _, entry := range entries {
		if entry.Hash != h.hash {
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}

func (h *hashSelectAction) handleGetBucketFromDB() {
	h.err = h.handleGetBucketFromDBWithError()
}

func (h *hashSelectAction) handleGetBucketFromDBWithError() error {
	bucketGetOutput, err := h.bucketLeaseGet()
	if err != nil {
		return err
	}

	// TODO Handle OK / Rejected
	if bucketGetOutput.Type == LeaseGetTypeGranted {
	}

	h.bucketLeaseID = bucketGetOutput.LeaseID

	begin := startOfSlot(h.hash, h.sizeLog)
	end := nextSlot(h.hash, h.sizeLog)
	h.entriesDBFn = h.root.db.SelectEntries(h.ctx, begin, end)

	h.root.sess.addNextCall(func() {
		h.handleBucketDataFromDB()
	})
	return nil
}

func (h *hashSelectAction) handleBucketDataFromDB() {
	h.results, h.err = h.handleBucketDataFromDBWithOutput()
}

func (h *hashSelectAction) handleBucketDataFromDBWithOutput() ([]Entry, error) {
	dbEntries, err := h.entriesDBFn()
	if err != nil {
		return nil, err
	}
	key := computeBucketKey(h.root.namespace, h.sizeLog, h.hash)
	h.root.pipeline.LeaseSet(key, marshalEntries(dbEntries), h.bucketLeaseID, 0) // TODO TTL
	return dbEntries, nil
}
