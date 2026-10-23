package coinmarketcap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/merloot/market-data/internal/domain/market"
)

type cmcErrorResponse struct {
	Status struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return market.ErrRateLimited
	case http.StatusNotFound:
		return market.ErrNotFound
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var apiErr cmcErrorResponse
	_ = json.Unmarshal(body, &apiErr)

	if apiErr.Status.ErrorMessage != "" {
		return fmt.Errorf("%w: status=%d code=%d message=%q",
			market.ErrUpstream,
			resp.StatusCode,
			apiErr.Status.ErrorCode,
			apiErr.Status.ErrorMessage,
		)
	}
	return fmt.Errorf("%w: status=%d body=%s",
		market.ErrUpstream,
		resp.StatusCode,
		string(body),
	)
}
