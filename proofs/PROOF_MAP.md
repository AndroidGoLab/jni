# Proof Map: Go Code to Lean Models

When modifying a Go function listed here, update the corresponding
Lean model and re-run `make prove` (or `cd proofs && lake build`).

The differential tests (`jni_difftest_test.go`) will catch translation
divergences between Lean models and Go implementations at `go test` time.

## Covered functions

| Go file | Go function | Lean module | Lean function | Fidelity |
|---------|------------|-------------|---------------|----------|
| `thread.go` | `VM.Do` | `ThreadAttachment` | `doOp` | Medium |
| `thread.go` | `getOrAttachCurrentThread` | `ThreadAttachment` | `getOrAttach` | High |
| `proxy.go` | `registerProxy` | `ProxyRegistry` | `register` | High |
| `proxy.go` | `registerProxyFull` | `ProxyRegistry` | `registerFull` | High |
| `proxy.go` | `lookupProxy` | `ProxyRegistry` | `lookup` | High |
| `proxy.go` | `lookupProxyFull` | `ProxyRegistry` | `lookupFull` | High |
| `proxy.go` | `unregisterProxy` | `ProxyRegistry` | `unregister` | High |
| `proxy.go` | `dispatchProxyInvocation` | `ProxyRegistry` | `dispatch` | Medium |
| `proxy.go` | `extractGoString` | `StringConversion` | `extractGoString` | Medium |
| `string_conversion.go` | `Env.GoString` | `StringConversion` | `goStringImpl` | Medium |
| `internal/jnierr/exception.go` | `CheckException` | `ExceptionProtocol` | `checkException` | Medium |
| `value.go` (generated) | `IntValue`, etc. | `JValueUnion` | `writeInt`, etc. | Medium |

## Not yet covered

| Go file | Go function | Why |
|---------|------------|-----|
| `proxy.go` | `doProxyInit` | Complex JNI init with ClassLoader fallback; hard to model without JVM |
| `proxy.go` | `NewProxy` / `NewProxyFull` | Multi-step JNI interaction; only nil-jvalue fix is modeled |
| `internal/jnierr/exception.go` | `extractClassName` / `extractMessage` | JNI call sequences; would need JVM mock model |
| `app/context.go` | `ensureContextInit` / `GetSystemService` | Android-specific; requires Android API model |

## Fidelity levels

- **High**: Lean model is a direct 1:1 translation of the Go logic.
  Changes to the Go function require updating the Lean model.
- **Medium**: Lean model captures the algorithm's core invariants but
  abstracts away unsafe pointer ops, CGo, or JNI runtime details.
  Changes to the Go algorithm require updating the Lean model;
  changes to the JNI plumbing do not.
