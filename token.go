package gosporestack

type TokenInfo struct {
	BalanceCents     int    `json:"balance_cents"`
	BalanceUSD       string `json:"balance_usd"`
	BurnRateCents    int    `json:"burn_rate_cents"`
	BurnRateUSD      string `json:"burn_rate_usd"`
	DaysRemaining    int    `json:"days_remaining"`
	Servers          int    `json:"servers"`
	AutorenewServers int    `json:"autorenew_servers"`
	SuspendedServers int    `json:"suspended_servers"`
}

type Payment struct {
	PaymentURI     string  `json:"payment_uri"`
	Cryptocurrency string  `json:"cryptocurrency"`
	Amount         int     `json:"amount"`
	FiatPerCoin    string  `json:"fiat_per_coin"`
	Created        int64   `json:"created"`
	Expires        int64   `json:"expires"`
	Paid           int64   `json:"paid"`
	TxID           string  `json:"txid"`
	AffiliateToken *string `json:"affiliate_token"` // Use pointer to handle null values
	ID             string  `json:"id"`
	Expired        bool    `json:"expired"`
}

type TokenInfoService struct {
	client *Client
}

func (s *TokenInfoService) Get() (*TokenInfo, error) {
	req, err := s.client.NewRequest("GET", "/token/"+s.client.token+"/info", nil)
	if err != nil {
		return nil, err
	}

	tokenInfo := &TokenInfo{}
	if err := s.client.DoRequest(req, tokenInfo); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *TokenInfoService) GetInvoices() (*Payment, error) {
	req, err := s.client.NewRequest("GET", "/token/"+s.client.token+"/invoices", nil)
	if err != nil {
		return nil, err
	}

	invoices := &Payment{}
	if err := s.client.DoRequest(req, invoices); err != nil {
		return nil, err
	}

	return invoices, nil
}
