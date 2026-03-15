package acl

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "acl_test.db")
	store, err := OpenStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestStore_GrantAndCheck(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.Grant("client1", "/jni_raw.JNIService/CallMethod", "admin"))

	allowed, err := store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	assert.True(t, allowed, "exact grant should allow the method")

	allowed, err = store.IsMethodAllowed("client1", "/jni_raw.JNIService/OtherMethod")
	require.NoError(t, err)
	assert.False(t, allowed, "different method should be denied")
}

func TestStore_WildcardGrant(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.Grant("client1", "/jni_raw.JNIService/*", "admin"))

	allowed, err := store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	assert.True(t, allowed, "service wildcard should match any method in that service")

	allowed, err = store.IsMethodAllowed("client1", "/jni_raw.JNIService/AnotherMethod")
	require.NoError(t, err)
	assert.True(t, allowed, "service wildcard should match another method in the same service")

	allowed, err = store.IsMethodAllowed("client1", "/other.Service/CallMethod")
	require.NoError(t, err)
	assert.False(t, allowed, "service wildcard should not match a different service")
}

func TestStore_GlobalWildcard(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.Grant("client1", "/*", "admin"))

	allowed, err := store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	assert.True(t, allowed, "global wildcard should allow any method")

	allowed, err = store.IsMethodAllowed("client1", "/other.Service/Method")
	require.NoError(t, err)
	assert.True(t, allowed, "global wildcard should allow methods in any service")
}

func TestStore_Revoke(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.Grant("client1", "/jni_raw.JNIService/CallMethod", "admin"))

	allowed, err := store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	require.True(t, allowed, "precondition: grant should be active")

	require.NoError(t, store.Revoke("client1", "/jni_raw.JNIService/CallMethod"))

	allowed, err = store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	assert.False(t, allowed, "revoked grant should deny the method")
}

func TestStore_RegisterClient(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.RegisterClient("client1", "-----BEGIN CERT-----\nfake\n-----END CERT-----", "AA:BB:CC"))

	clients, err := store.ListClients()
	require.NoError(t, err)
	require.Len(t, clients, 1)

	c := clients[0]
	assert.Equal(t, "client1", c.ClientID)
	assert.Equal(t, "-----BEGIN CERT-----\nfake\n-----END CERT-----", c.CertPEM)
	assert.Equal(t, "AA:BB:CC", c.Fingerprint)
	assert.False(t, c.RegisteredAt.IsZero(), "registered_at should be populated")
}

func TestStore_PendingRequests(t *testing.T) {
	store := openTestStore(t)

	methods := []string{"/jni_raw.JNIService/CallMethod", "/jni_raw.JNIService/GetField"}
	reqID, err := store.CreateRequest("client1", methods)
	require.NoError(t, err)
	assert.Greater(t, reqID, int64(0))

	pending, err := store.ListPendingRequests()
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, "client1", pending[0].ClientID)
	assert.Equal(t, methods, pending[0].Methods)
	assert.Equal(t, "pending", pending[0].Status)

	require.NoError(t, store.ApproveRequest(reqID, "admin"))

	// After approval: grants should exist.
	for _, m := range methods {
		allowed, err := store.IsMethodAllowed("client1", m)
		require.NoError(t, err)
		assert.True(t, allowed, "approved method %q should be allowed", m)
	}

	// After approval: request should no longer appear in pending.
	pending, err = store.ListPendingRequests()
	require.NoError(t, err)
	assert.Empty(t, pending, "approved request should not be in pending list")
}

func TestStore_DenyRequest(t *testing.T) {
	store := openTestStore(t)

	methods := []string{"/jni_raw.JNIService/CallMethod"}
	reqID, err := store.CreateRequest("client1", methods)
	require.NoError(t, err)

	require.NoError(t, store.DenyRequest(reqID))

	// No grants should be created.
	allowed, err := store.IsMethodAllowed("client1", "/jni_raw.JNIService/CallMethod")
	require.NoError(t, err)
	assert.False(t, allowed, "denied request should not create grants")

	// Request should not be pending.
	pending, err := store.ListPendingRequests()
	require.NoError(t, err)
	assert.Empty(t, pending, "denied request should not be in pending list")
}

func TestStore_DuplicateGrant(t *testing.T) {
	store := openTestStore(t)

	require.NoError(t, store.Grant("client1", "/jni_raw.JNIService/CallMethod", "admin"))
	require.NoError(t, store.Grant("client1", "/jni_raw.JNIService/CallMethod", "admin2"),
		"duplicate grant should not return an error (INSERT OR IGNORE)")

	grants, err := store.ListGrants("client1")
	require.NoError(t, err)
	assert.Len(t, grants, 1, "duplicate grant should not create a second row")
}
