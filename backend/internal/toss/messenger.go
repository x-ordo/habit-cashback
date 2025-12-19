package toss

import (
	"context"
	"fmt"
)

type SendMessageReq struct {
	TemplateSetCode string         `json:"templateSetCode"`
	Context         map[string]any `json:"context"`
}

type SendMessageSuccess struct {
	MsgCount      int `json:"msgCount"`
	SentPushCount int `json:"sentPushCount"`
	SentInboxCount int `json:"sentInboxCount"`
}

func (c *Client) SendMessage(ctx context.Context, tossUserKey string, req SendMessageReq) (SendMessageSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/messenger/send-message", c.APIBaseURL)
	headers := map[string]string{"X-Toss-User-Key": tossUserKey}
	env, raw, err := doJSON[SendMessageSuccess](ctx, c.http, "POST", url, headers, req)
	if err != nil {
		return SendMessageSuccess{}, raw, err
	}
	// docs use `result` field here sometimes; our envelope maps both success/result.
	if env.Success != nil {
		return *env.Success, raw, nil
	}
	if env.Result != nil {
		return *env.Result, raw, nil
	}
	return SendMessageSuccess{}, raw, fmt.Errorf("no success/result: resultType=%s", env.ResultType)
}
