package geetest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	validateURL = "https://gcaptcha4.geetest.com/validate"
	httpTimeout = 10 * time.Second
)

var (
	ErrVerifyFailed = errors.New("geetest: verification failed")
	ErrVerifyError  = errors.New("geetest: request error")

	DefaultVerifier *Verifier

	httpClient = &http.Client{Timeout: httpTimeout}
)

type Verifier struct {
	captchaID  string
	captchaKey string
}

func InitVerifier(cfg *Config) {
	DefaultVerifier = &Verifier{
		captchaID:  cfg.CaptchaID,
		captchaKey: cfg.CaptchaKey,
	}
}

func (v *Verifier) Verify(ctx context.Context, lotNumber, captchaOutput, passToken, genTime string) error {
	signToken := v.sign(lotNumber)

	form := url.Values{}
	form.Set("captcha_id", v.captchaID)
	form.Set("lot_number", lotNumber)
	form.Set("captcha_output", captchaOutput)
	form.Set("pass_token", passToken)
	form.Set("gen_time", genTime)
	form.Set("sign_token", signToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, validateURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ErrVerifyError
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ErrVerifyError
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrVerifyError
	}

	var result struct {
		Result string `json:"result"`
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return ErrVerifyError
	}
	if result.Result != "success" {
		return ErrVerifyFailed
	}
	return nil
}

func (v *Verifier) sign(lotNumber string) string {
	mac := hmac.New(sha256.New, []byte(v.captchaKey))
	mac.Write([]byte(lotNumber))
	return hex.EncodeToString(mac.Sum(nil))
}
