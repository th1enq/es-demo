package mappers

import (
	"github.com/Rhymond/go-money"
	"github.com/th1enq/es-demo/internal/domain"
)

func BalanceFromMoney(money *money.Money) domain.Balance {
	return domain.Balance{
		Amount:   money.AsMajorUnits(),
		Currency: money.Currency().Code,
	}
}
