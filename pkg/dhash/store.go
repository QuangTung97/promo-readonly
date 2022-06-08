package dhash

import (
	"context"
	"time"
)

type storeImpl struct {
	sess     *sessionImpl
	db       StoreDatabase
	pipeline CachePipeline
}

type storeGetAction struct {
	root *storeImpl
	ctx  context.Context
	key  string

	leaseGetFn func() (LeaseGetOutput, error)

	leaseWaitStarted   bool
	leaseWaitDurations []time.Duration

	data []byte
	err  error
}

func (s *storeGetAction) handleLeaseGet() {
	s.data, s.err = s.handleLeaseGetWithOutput()
}

func (s *storeGetAction) handleLeaseGetWithOutput() ([]byte, error) {
	output, err := s.leaseGetFn()
	if err != nil {
		return nil, err
	}

	if output.Type == LeaseGetTypeGranted {
		s.root.sess.storeMissCount++

		dbFn := s.root.db.Get(s.ctx, s.key)
		s.root.sess.addNextCall(func() {
			dbData, err := dbFn()
			if err != nil {
				s.err = err
				s.root.pipeline.Delete(s.key)
				return
			}
			s.data = dbData
			s.root.pipeline.LeaseSet(s.key, s.data, output.LeaseID, 0) // TODO TTL
		})
		return nil, nil
	}

	if output.Type == LeaseGetTypeRejected {
		s.root.sess.storeMissCount++

		sess := s.root.sess

		if !s.leaseWaitStarted {
			s.leaseWaitStarted = true
			s.leaseWaitDurations = sess.options.waitLeaseDurations
		}

		if len(s.leaseWaitDurations) == 0 {
			return nil, ErrLeaseNotGranted
		}
		duration := s.leaseWaitDurations[0]
		s.leaseWaitDurations = s.leaseWaitDurations[1:]

		s.leaseGetFn = s.root.pipeline.LeaseGet(s.key)

		sess.addDelayedCall(duration, func() {
			s.handleLeaseGet()
		})
		return nil, nil
	}

	return output.Data, nil
}

// Get ...
func (s *storeImpl) Get(ctx context.Context, key string) func() ([]byte, error) {
	s.sess.storeAccessCount++

	fn := s.pipeline.LeaseGet(key)
	action := &storeGetAction{
		root:       s,
		key:        key,
		ctx:        ctx,
		leaseGetFn: fn,
	}

	s.sess.addNextCall(func() {
		action.handleLeaseGet()
	})

	return func() ([]byte, error) {
		s.sess.processAllCalls()
		return action.data, action.err
	}
}

// Invalidate ...
func (s *storeImpl) Invalidate(_ context.Context, key string) func() error {
	return s.pipeline.Delete(key)
}
