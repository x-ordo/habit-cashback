package toss

import (
	"context"
	"fmt"
)

type GenerateTokenReq struct {
	AuthorizationCode string `json:"authorizationCode"`
	Referrer          string `json:"referrer"`
}

type GenerateTokenSuccess struct {
	TokenType    string `json:"tokenType"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	Scope        string `json:"scope"`
}

func (c *Client) GenerateToken(ctx context.Context, req GenerateTokenReq) (GenerateTokenSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/user/oauth2/generate-token", c.APIBaseURL)
	env, raw, err := doJSON[GenerateTokenSuccess](ctx, c.http, "POST", url, nil, req)
	if err != nil {
		return GenerateTokenSuccess{}, raw, err
	}
	if env.Success == nil {
		return GenerateTokenSuccess{}, raw, fmt.Errorf("no success: resultType=%s", env.ResultType)
	}
	return *env.Success, raw, nil
}

type LoginMeSuccess struct {
	UserKey int64  `json:"userKey"`
	Scope   string `json:"scope"`
	// other fields are encrypted/pii; we intentionally ignore them in MVP.
}

func (c *Client) LoginMe(ctx context.Context, accessToken string) (LoginMeSuccess, []byte, error) {
	url := fmt.Sprintf("%s/api-partner/v1/apps-in-toss/user/oauth2/login-me", c.APIBaseURL)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	env, raw, err := doJSON[LoginMeSuccess](ctx, c.http, "GET", url, headers, nil)
	if err != nil {
		return LoginMeSuccess{}, raw, err
	}
	if env.Success == nil {
		return LoginMeSuccess{}, raw, fmt.Errorf("no success: resultType=%s", env.ResultType)
	}
	return *env.Success, raw, nil
}
