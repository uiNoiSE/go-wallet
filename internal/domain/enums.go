package domain

type OperationType string

const (
	OpDeposit  OperationType = "DEPOSIT"
	OpWithdraw OperationType = "WITHDRAW"
)

func (o OperationType) IsValid() bool {
	switch o {
	case OpDeposit, OpWithdraw:
		return true
	}
	return false
}
