package minima

type NodeStatus struct {
	Status      string `json:"status"`
	Synced      bool   `json:"synced"`
	BlockHeight int64  `json:"block_height"`
	Peers       int    `json:"peers"`
	Uptime      int64  `json:"uptime"`
	Role        string `json:"role"`
	Version     string `json:"version"`
	ChainID     string `json:"chain_id"`
}

type Balance struct {
	Confirmed   string `json:"confirmed"`
	Unconfirmed string `json:"unconfirmed"`
	Asset       string `json:"asset"`
	TokenID     string `json:"token_id"`
}

type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Asset     string `json:"asset"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Block     int64  `json:"block"`
	Fee       string `json:"fee"`
}

type CoinDetails struct {
	TokenID  string `json:"tokenid"`
	Amount   string `json:"amount"`
	Asset    string `json:"asset"`
	Creator  string `json:"creator"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	Script   string `json:"script"`
}

type ContractResult struct {
	ContractID string `json:"contractid"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	State      string `json:"state"`
}

type KeyPair struct {
	PublicKey  string `json:"publickey"`
	PrivateKey string `json:"privatekey"`
	Address    string `json:"address"`
}

type MiniDAPPInfo struct {
	UID         string `json:"uid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Icon        string `json:"icon"`
	Status      string `json:"status"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}
