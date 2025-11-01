package mappers

import (
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/dto"
)

func BankAccountToMongoProjection(bankAccount *domain.BankAccountAggregate) *domain.BankAccountMongoProjection {
	return &domain.BankAccountMongoProjection{
		AggregateID: bankAccount.BankAccount.AggregateID,
		Version:     bankAccount.Version,
		Email:       bankAccount.BankAccount.Email,
		FirstName:   bankAccount.BankAccount.FirstName,
		LastName:    bankAccount.BankAccount.LastName,
		Balance: domain.Balance{
			Amount:   bankAccount.BankAccount.Balance.AsMajorUnits(),
			Currency: bankAccount.BankAccount.Balance.Currency().Code,
		},
		Status: bankAccount.BankAccount.Status,
	}
}

func BankAccountMongoProjectionToHttp(bankAccount *domain.BankAccountMongoProjection) *dto.HttpBankAccountResponse {
	return &dto.HttpBankAccountResponse{
		AggregateID: bankAccount.AggregateID,
		Email:       bankAccount.Email,
		FirstName:   bankAccount.FirstName,
		LastName:    bankAccount.LastName,
		Balance:     bankAccount.Balance,
		Status:      bankAccount.Status,
	}
}
