package command

type BankAccountCommand struct {
	CreateBankAccount
	DepositeBalance
	WithdrawBalance
}

func NewBankAccountCommand(
	createBankAccount CreateBankAccount,
	depositeBalance DepositeBalance,
	withdrawBalance WithdrawBalance,
) *BankAccountCommand {
	return &BankAccountCommand{
		CreateBankAccount: createBankAccount,
		DepositeBalance:   depositeBalance,
		WithdrawBalance:   withdrawBalance,
	}
}
