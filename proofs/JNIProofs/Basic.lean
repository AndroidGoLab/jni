-- Basic definitions shared across all JNI proofs.

-- A HandlerID is a positive natural number. The Go code uses
-- atomic.Int64.Add(1) starting from 0, so IDs are always ≥ 1.
abbrev HandlerID := Nat

-- JNI return codes modeled as an inductive type.
inductive JNIResult where
  | ok
  | err
  | edetached
  | eversion
  | enomem
  | eexist
  | einval
  deriving DecidableEq, Repr
