package toss

import (
	"context"
	"fmt"
)

type MakePaymentReq struct {
	OrderNo        string `json:"orderNo"`
	ProductDesc    string `json:"productDesc"`
	Amount         int64  `json:"amount"`
	AmountTaxFree  int64  `json:"amountTaxFree"`
	Installment    string `json:"installment,omitempty"`
	IsTestPayment  bool   `json:"isTestPayment"`
}

type MakePaymentSuccess struct {
	PayToken string `json:"payToken"`
}

func (c *Client) MakePayment(ctx context.Context, tossUserKey string, idemKey string, req MakePaymentReq) (MakePaymentSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/pay/make-payment", c.PayBaseURL)
	headers := map[string]string{
		"X-Toss-User-Key": tossUserKey,
	}
	if idemKey != "" {
		headers["Idempotency-Key"] = idemKey
	}
	env, raw, err := doJSON[MakePaymentSuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return MakePaymentSuccess{}, raw, err
	}
	if env.Success == nil {
		return MakePaymentSuccess{}, raw, fmt.Errorf("no success: resultType=%s", env.ResultType)
	}
	return *env.Success, raw, nil
}

type ExecutePaymentReq struct {
	PayToken       string `json:"payToken"`
	OrderNo        string `json:"orderNo"`
	IsTestPayment  bool   `json:"isTestPayment"`
}

type ExecutePaymentSuccess struct {
	// Toss Pay docs return many fields; we keep the envelope raw as well.
	PayToken string `json:"payToken,omitempty"`
}

func (c *Client) ExecutePayment(ctx context.Context, tossUserKey string, req ExecutePaymentReq) (ExecutePaymentSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/pay/execute-payment", c.PayBaseURL)
	headers := map[string]string{
		"X-Toss-User-Key": tossUserKey,
	}
	env, raw, err := doJSON[ExecutePaymentSuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return ExecutePaymentSuccess{}, raw, err
	}
	// execute-payment may return in `success` or `result`
	if env.Success != nil {
		return *env.Success, raw, nil
	}
	if env.Result != nil {
		return *env.Result, raw, nil
	}
	return ExecutePaymentSuccess{}, raw, fmt.Errorf("no success/result: resultType=%s", env.ResultType)
}
