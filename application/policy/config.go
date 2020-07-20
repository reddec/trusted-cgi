package policy

import (
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/internal"
	"os"
	"sync"
)

type naiveFileStorePayload struct {
	Policies []application.Policy `json:"policies"`
}

func FileConfig(filename string) *naiveFileStore {
	return &naiveFileStore{file: filename}
}

type naiveFileStore struct {
	file string
	lock sync.RWMutex
}

func (nfs *naiveFileStore) SetPolicies(policies []application.Policy) error {
	nfs.lock.Lock()
	defer nfs.lock.Unlock()
	return internal.AtomicWriteJson(nfs.file, &naiveFileStorePayload{Policies: policies})
}

func (nfs *naiveFileStore) GetPolicies() ([]application.Policy, error) {
	nfs.lock.RLock()
	defer nfs.lock.RUnlock()
	var pd naiveFileStorePayload
	err := internal.ReadJson(nfs.file, &pd)
	if err == nil {
		return pd.Policies, nil
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, err
}

func Mock(policies ...application.Policy) *mockStore {
	return &mockStore{policies: policies}
}

type mockStore struct {
	lock     sync.RWMutex
	policies []application.Policy
}

func (msc *mockStore) SetPolicies(policies []application.Policy) error {
	msc.lock.Lock()
	defer msc.lock.Unlock()
	msc.policies = make([]application.Policy, len(policies))
	copy(msc.policies, policies)
	return nil
}

func (msc *mockStore) GetPolicies() ([]application.Policy, error) {
	msc.lock.RLock()
	defer msc.lock.RUnlock()
	out := make([]application.Policy, len(msc.policies))
	copy(out, msc.policies)
	return out, nil
}
