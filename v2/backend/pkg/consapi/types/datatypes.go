package types

import (
	"errors"
	"strconv"
)

const (
	PendingInitialized ValidatorStatus = 0
	PendingQueued      ValidatorStatus = 1
	ActiveOngoing      ValidatorStatus = 2
	ActiveExiting      ValidatorStatus = 3
	ActiveSlashed      ValidatorStatus = 4
	ExitedUnslashed    ValidatorStatus = 5
	ExitedSlashed      ValidatorStatus = 6
	WithdrawalPossible ValidatorStatus = 7
	WithdrawalDone     ValidatorStatus = 8
	Active             ValidatorStatus = 9
	Pending            ValidatorStatus = 10
	Exited             ValidatorStatus = 11
	Withdrawal         ValidatorStatus = 12
)

type ValidatorStatus int8

func (s ValidatorStatus) IsActive() bool {
	return s == ActiveOngoing || s == ActiveExiting || s == ActiveSlashed || s == Active
}

// unmarshal validatorstatus
func (v *ValidatorStatus) UnmarshalJSON(b []byte) error {
	if b[0] == '"' || b[0] == '\'' {
		if len(b) == 1 || b[len(b)-1] != b[0] {
			return errors.New("uneven/missing quotes")
		}
		b = b[1 : len(b)-1]
	}
	result, err := NewValidatorStatusFromString(string(b))
	if err != nil {
		return err
	}
	*v = result
	return nil
}

func NewValidatorStatusFromString(s string) (ValidatorStatus, error) {
	switch s {
	case "pending_initialized":
		return PendingInitialized, nil
	case "pending_queued":
		return PendingQueued, nil
	case "active_ongoing":
		return ActiveOngoing, nil
	case "active_exiting":
		return ActiveExiting, nil
	case "active_slashed":
		return ActiveSlashed, nil
	case "exited_unslashed":
		return ExitedUnslashed, nil
	case "exited_slashed":
		return ExitedSlashed, nil
	case "withdrawal_possible":
		return WithdrawalPossible, nil
	case "withdrawal_done":
		return WithdrawalDone, nil
	case "active":
		return Active, nil
	case "pending":
		return Pending, nil
	case "exited":
		return Exited, nil
	case "withdrawal":
		return Withdrawal, nil
	}
	return 0, errors.New("invalid validator status: " + s)
}

func (v *ValidatorStatus) String() string {
	switch *v {
	case PendingInitialized:
		return "pending_initialized"
	case PendingQueued:
		return "pending_queued"
	case ActiveOngoing:
		return "active_ongoing"
	case ActiveExiting:
		return "active_exiting"
	case ActiveSlashed:
		return "active_slashed"
	case ExitedUnslashed:
		return "exited_unslashed"
	case ExitedSlashed:
		return "exited_slashed"
	case WithdrawalPossible:
		return "withdrawal_possible"
	case WithdrawalDone:
		return "withdrawal_done"
	case Active:
		return "active"
	case Pending:
		return "pending"
	case Exited:
		return "exited"
	case Withdrawal:
		return "withdrawal"
	}
	return "unknown"
}

func ConvertToStringSlice(status []ValidatorStatus) []string {
	strSlice := make([]string, len(status))
	for i, s := range status {
		strSlice[i] = s.String()
	}
	return strSlice
}

type Uint64Str uint64

func (s *Uint64Str) UnmarshalJSON(b []byte) error {
	return Uint64Unmarshal((*uint64)(s), b)
}

// Parse a uint64, with or without quotes, in any base, with common prefixes accepted to change base.
func Uint64Unmarshal(v *uint64, b []byte) error {
	if v == nil {
		return errors.New("nil dest in uint64 decoding")
	}
	if len(b) == 0 {
		return errors.New("empty uint64 input")
	}
	if b[0] == '"' || b[0] == '\'' {
		if len(b) == 1 || b[len(b)-1] != b[0] {
			return errors.New("uneven/missing quotes")
		}
		b = b[1 : len(b)-1]
	}
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = n
	return nil
}
