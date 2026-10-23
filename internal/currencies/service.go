package currencies

import "context"

type service struct{
	repo Repository
}

type Service interface{
	ListCurrency(ctx context.Context)([]CurrencyResponse,error)
}

func NewService(repo Repository) Service{
	return &service{repo : repo}
}


func _serialize_currency(currency Currency)CurrencyResponse{
	return CurrencyResponse{
		Code : currency.Code,
		Name :currency.Name,
		Exponent: currency.Exponent,
	}
}

func (s *service) ListCurrency(ctx context.Context)([]CurrencyResponse, error){
	// list all currency
	currency, err := s.repo.ListActive(ctx)
	if err!=nil{
		return nil, err
	}

	var currencies []CurrencyResponse
	for _, value := range currency{
		response := _serialize_currency(value)
		currencies = append(currencies, response)
	}

	return currencies, nil

}