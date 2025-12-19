package toss

import (
	"context"
	"fmt"
)

type GetKeyReq struct {
	PromotionCode string `json:"promotionCode"`
}

type GetKeySuccess struct {
	Key string `json:"key"`
}

func (c *Client) PromotionGetKey(ctx context.Context, tossUserKey string, req GetKeyReq) (GetKeySuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/promotion/get-key", c.APIBaseURL)
	headers := map[string]string{"X-Toss-User-Key": tossUserKey}
	env, raw, err := doJSON[GetKeySuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return GetKeySuccess{}, raw, err
	}
	if env.Success == nil {
		return GetKeySuccess{}, raw, fmt.Errorf("no success: resultType=%s", env.ResultType)
	}
	return *env.Success, raw, nil
}

type ExecutePromotionReq struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type ExecutePromotionSuccess struct {
	ResultType string `json:"resultType,omitempty"`
}

func (c *Client) PromotionExecute(ctx context.Context, tossUserKey string, idemKey string, req ExecutePromotionReq) (ExecutePromotionSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/promotion/execute-promotion", c.APIBaseURL)
	headers := map[string]string{"X-Toss-User-Key": tossUserKey}
	if idemKey != "" {
		headers["Idempotency-Key"] = idemKey
	}
	env, raw, err := doJSON[ExecutePromotionSuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return ExecutePromotionSuccess{}, raw, err
	}
	// This endpoint uses `success` envelope but may just return resultType.
	if env.Success != nil {
		return *env.Success, raw, nil
	}
	return ExecutePromotionSuccess{ResultType: env.ResultType}, raw, nil
}

type ExecutionResultReq struct {
	Key string `json:"key"`
}

type ExecutionResultSuccess struct {
	ResultType string `json:"resultType,omitempty"`
}

func (c *Client) PromotionExecutionResult(ctx context.Context, tossUserKey string, req ExecutionResultReq) (ExecutionResultSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/promotion/execution-result", c.APIBaseURL)
	headers := map[string]string{"X-Toss-User-Key": tossUserKey}
	env, raw, err := doJSON[ExecutionResultSuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return ExecutionResultSuccess{}, raw, err
	}
	if env.Success != nil {
		return *env.Success, raw, nil
	}
	return ExecutionResultSuccess{ResultType: env.ResultType}, raw, nil
}
