package handlestore

import (
	"sync"
	"sync/atomic"

	"github.com/AndroidGoLab/jni"
)

// HandleStore maps opaque int64 handles to JNI global references, allowing
// Java objects to be referenced across gRPC RPC boundaries.
//
// When a server-side RPC returns a JNI object, it stores the object as a
// global reference in the HandleStore and returns the int64 handle to the
// client. When the client later passes that handle in another RPC, the
// server retrieves the global reference from the HandleStore.
//
// HandleStore is safe for concurrent use.
type HandleStore struct {
	mu      sync.RWMutex
	objects map[int64]*jni.Object
	nextID  atomic.Int64
}

// New creates a new empty HandleStore.
func New() *HandleStore {
	return &HandleStore{
		objects: make(map[int64]*jni.Object),
	}
}

// Put creates a JNI global reference for the given object and returns an
// opaque int64 handle. The caller must be inside a VM.Do callback (env is
// required to create the global reference). Returns 0 if obj is nil.
func (s *HandleStore) Put(env *jni.Env, obj *jni.Object) int64 {
	if obj == nil {
		return 0
	}
	globalRef := env.NewGlobalRef(obj)
	id := s.nextID.Add(1)
	s.mu.Lock()
	s.objects[id] = globalRef
	s.mu.Unlock()
	return id
}

// Get retrieves the JNI global reference for the given handle. Returns nil
// if the handle is invalid or 0 (the zero handle represents nil).
func (s *HandleStore) Get(handle int64) *jni.Object {
	if handle == 0 {
		return nil
	}
	s.mu.RLock()
	obj := s.objects[handle]
	s.mu.RUnlock()
	return obj
}

// Release deletes the JNI global reference for the given handle and removes
// it from the store. The caller must be inside a VM.Do callback. Does nothing
// if the handle is 0 or not found.
func (s *HandleStore) Release(env *jni.Env, handle int64) {
	if handle == 0 {
		return
	}
	s.mu.Lock()
	obj, ok := s.objects[handle]
	if ok {
		delete(s.objects, handle)
	}
	s.mu.Unlock()
	if ok {
		env.DeleteGlobalRef(obj)
	}
}

// Len returns the number of active handles.
func (s *HandleStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.objects)
}
