package miniprogram

import (
	"context"
	"net/http"
)

// PhoneNumberWatermark carries the appid and capture timestamp of a phone
// number record; compare the appid to the current app to detect replayed data.
type PhoneNumberWatermark struct {
	// Timestamp is the Unix timestamp (seconds) when the data was obtained.
	Timestamp int `json:"timestamp"`
	// AppID recorded in the watermark; must equal the calling Mini Program.
	AppID string `json:"appid"`
}

// PhoneNumberInfo describes a user's phone number retrieved via a headless
// phone-number button.
type PhoneNumberInfo struct {
	// PhoneNumber is the full phone number with country calling code prefix.
	PhoneNumber string `json:"phoneNumber"`
	// PurePhoneNumber is the phone number without the country calling code.
	PurePhoneNumber string `json:"purePhoneNumber"`
	// CountryCode is the country calling code, e.g. "86".
	CountryCode string `json:"countryCode"`
	// Watermark helps verifying the data belongs to this Mini Program.
	Watermark PhoneNumberWatermark `json:"watermark"`
}

// GetUserPhoneNumberResponse is returned by GetUserPhoneNumber.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/phone-number/api_getphonenumber.html
type GetUserPhoneNumberResponse struct {
	ErrResponse
	// PhoneInfo is the resolved phone number payload.
	PhoneInfo PhoneNumberInfo `json:"phone_info"`
}

// GetUserPhoneNumber resolves the user's phone number from the code obtained by
// a headless phone-number button in the Mini Program
// (button open-type="getPhoneNumber"). The code is single use and valid for
// five minutes.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/phone-number/api_getphonenumber.html
func (w *MiniProgram) GetUserPhoneNumber(ctx context.Context, code string) (*GetUserPhoneNumberResponse, error) {
	var result GetUserPhoneNumberResponse
	err := w.withAccessToken(ctx, http.MethodPost, "/wxa/business/getuserphonenumber", nil, defaultReqOptions(), map[string]string{"code": code}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
