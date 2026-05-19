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
			AccountId        string  `json:"accountId"`
			CustomerName     string  `json:"customerName"`
			BalanceRemaining float64 `json:"balanceRemaining,string"`
			// Add other needed fields as required
		} `json:"postBalanceDetails"`
	} `json:"data"`
}
