package queuemanager

import (
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/internal"
	"os"
	"sync"
)

type naiveFileStorePayload struct {
	Queues []application.Queue `json:"queues"`
}

func FileConfig(filename string) *naiveFileStore {
	return &naiveFileStore{file: filename}
}

type naiveFileStore struct {
	file string
	lock sync.RWMutex
}

func (nfs *naiveFileStore) SetQueues(queues []application.Queue) error {
	nfs.lock.Lock()
	defer nfs.lock.Unlock()
	return internal.AtomicWriteJson(nfs.file, &naiveFileStorePayload{Queues: queues})
}

func (nfs *naiveFileStore) GetQueues() ([]application.Queue, error) {
	nfs.lock.RLock()
	defer nfs.lock.RUnlock()
	var pd naiveFileStorePayload
	err := internal.ReadJson(nfs.file, &pd)
	if err == nil {
		return pd.Queues, nil
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, err
}

func Mock(queues ...application.Queue) *mockStore {
	return &mockStore{queues: queues}
}

type mockStore struct {
	lock   sync.RWMutex
	queues []application.Queue
}

func (msc *mockStore) SetQueues(queues []application.Queue) error {
	msc.lock.Lock()
	defer msc.lock.Unlock()
	msc.queues = make([]application.Queue, len(queues))
	copy(msc.queues, queues)
	return nil
}

func (msc *mockStore) GetQueues() ([]application.Queue, error) {
	msc.lock.RLock()
	defer msc.lock.RUnlock()
	out := make([]application.Queue, len(msc.queues))
	copy(out, msc.queues)
	return out, nil
}
