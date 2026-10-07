package river

import "encoding/json"

type oracleArgs struct {
	Payload json.RawMessage `json:"payload`
}

func (oracleArgs) Kind() string { return "market-data-oracle" }
