package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	INR = "INR"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, INR, EUR, CAD:
		return true
	}
	return false
}