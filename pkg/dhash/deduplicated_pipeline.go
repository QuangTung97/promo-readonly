package dhash

type deduplicatedPipeline struct {
	getFuncs      map[string]func() (GetOutput, error)
	leaseGetFuncs map[string]func() (LeaseGetOutput, error)
	leaseSetKeys  map[string]uint64
	CachePipeline
}

func newDeduplicatedPipeline(root CachePipeline) *deduplicatedPipeline {
	d := &deduplicatedPipeline{
		CachePipeline: root,
	}
	d.reset()
	return d
}

func (d *deduplicatedPipeline) Get(key string) func() (GetOutput, error) {
	lastFunc, existed := d.getFuncs[key]
	if existed {
		return lastFunc
	}

	fn := d.CachePipeline.Get(key)

	executed := false
	var output GetOutput
	var err error

	getFunc := func() (GetOutput, error) {
		if executed {
			return output, err
		}
		executed = true
		output, err = fn()
		return output, err
	}
	d.getFuncs[key] = getFunc
	return getFunc
}

func (d *deduplicatedPipeline) LeaseGet(key string) func() (LeaseGetOutput, error) {
	lastFunc, existed := d.leaseGetFuncs[key]
	if existed {
		return lastFunc
	}

	executed := false
	var output LeaseGetOutput
	var err error

	fn := d.CachePipeline.LeaseGet(key)
	getFunc := func() (LeaseGetOutput, error) {
		if executed {
			return output, err
		}
		executed = true
		output, err = fn()
		return output, err
	}
	d.leaseGetFuncs[key] = getFunc

	return getFunc
}

func (d *deduplicatedPipeline) LeaseSet(key string, value []byte, leaseID uint64, ttl uint32) func() error {
	existedLeaseID, existed := d.leaseSetKeys[key]
	if existed && leaseID == existedLeaseID {
		return func() error { return nil }
	}
	d.leaseSetKeys[key] = leaseID
	return d.CachePipeline.LeaseSet(key, value, leaseID, ttl)
}

func (d *deduplicatedPipeline) reset() {
	d.getFuncs = map[string]func() (GetOutput, error){}
	d.leaseGetFuncs = map[string]func() (LeaseGetOutput, error){}
	d.leaseSetKeys = map[string]uint64{}
}
