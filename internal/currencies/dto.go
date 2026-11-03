package currencies


type CurrencyResponse struct{
	Code string `json:"code"`
	Name string `json:"name"`
	Exponent int16 `json:"exponent"`
}