package types

import "cosmossdk.io/errors"

var (
	ErrZeroDeposit          = errors.Register(ModuleName, 2, "zero deposit")
	ErrExceedShieldLimit    = errors.Register(ModuleName, 3, "exceeds single tx limit")
	ErrPoolFull             = errors.Register(ModuleName, 4, "pool full")
	ErrZeroCommitment       = errors.Register(ModuleName, 5, "zero commitment")
	ErrUnknownRoot          = errors.Register(ModuleName, 6, "unknown root")
	ErrAlreadySpent         = errors.Register(ModuleName, 7, "already spent")
	ErrZeroAmount           = errors.Register(ModuleName, 8, "zero amount")
	ErrInsufficientPool     = errors.Register(ModuleName, 9, "insufficient pool")
	ErrZeroRecipient        = errors.Register(ModuleName, 10, "zero recipient")
	ErrDuplicateNullifiers  = errors.Register(ModuleName, 11, "duplicate nullifiers")
	ErrDuplicateCommitments = errors.Register(ModuleName, 12, "duplicate commitments")
	ErrNotRegistered        = errors.Register(ModuleName, 13, "not registered")
	ErrAlreadyRegistered    = errors.Register(ModuleName, 14, "already registered")
	ErrCommitmentTaken      = errors.Register(ModuleName, 15, "commitment taken")
	ErrInvalidProof         = errors.Register(ModuleName, 16, "invalid proof")
	ErrKeyNotRegistered     = errors.Register(ModuleName, 17, "verifying key not registered")
	ErrTreeFull             = errors.Register(ModuleName, 18, "commitment tree full")
)
