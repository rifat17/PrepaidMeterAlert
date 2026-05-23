package dpdc

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type BalanceQueryRequest struct {
	Query string `json:"query"`
}

type PostBalanceDetailsResponse struct {
	Data struct {
		PostBalanceDetails struct {
			CustomerName      string  `json:"customerName"`
			BalanceRemaining  float64 `json:"balanceRemaining,string"`
			ErrorMessage      string  `json:"errorMessage"`
			BalanceLatestDate string  `json:"balanceLatestDate"`
		} `json:"postBalanceDetails"`
	} `json:"data"`
}
