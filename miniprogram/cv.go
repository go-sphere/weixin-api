package miniprogram

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// OCRImgRequest points the OCR/CV endpoints at the image to process.
type OCRImgRequest struct {
	// ImgURL of the image (publicly reachable, https preferred).
	ImgURL string
	// Type is an OCR-mode selector used by some endpoints (idcard:
	// photo/scan; bankcard: photo/scan; driving: photo/scan). Empty uses the
	// server default.
	Type string
}

func (r *OCRImgRequest) query() url.Values {
	q := url.Values{}
	q.Set("img_url", r.ImgURL)
	if r.Type != "" {
		q.Set("type", r.Type)
	}
	return q
}

// OCRCoordinate is one corner of an OCR bounding box.
type OCRCoordinate struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

// OCRPosition is the quadrilateral bounding box of a recognised region.
type OCRPosition struct {
	LeftTop     OCRCoordinate `json:"left_top,omitempty"`
	RightTop    OCRCoordinate `json:"right_top,omitempty"`
	RightBottom OCRCoordinate `json:"right_bottom,omitempty"`
	LeftBottom  OCRCoordinate `json:"left_bottom,omitempty"`
}

// OCRImageSize is the pixel size of the source image.
type OCRImageSize struct {
	Width  int `json:"w,omitempty"`
	Height int `json:"h,omitempty"`
}

// OCRIDCardResponse is the recognised content of an ID card.
type OCRIDCardResponse struct {
	ErrResponse
	// Type is "photo" or "scan" reflecting the request mode.
	Type string `json:"type,omitempty"`
	// Name on the ID card.
	Name string `json:"name,omitempty"`
	// ID is the ID card number.
	ID string `json:"id,omitempty"`
	// Address on the ID card.
	Address string `json:"addr,omitempty"`
	// Gender on the ID card.
	Gender string `json:"gender,omitempty"`
	// Nationality on the ID card.
	Nationality string `json:"nationality,omitempty"`
	// ValidDate of the ID card.
	ValidDate string `json:"valid_date,omitempty"`
}

// OCRIDCard recognises the front (type=photo) or back (type=scan) of a Chinese
// resident ID card.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_idcardocr.html
func (w *MiniProgram) OCRIDCard(ctx context.Context, req *OCRImgRequest) (*OCRIDCardResponse, error) {
	var result OCRIDCardResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/idcard", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OCRBankCardResponse carries the recognised bank card number.
type OCRBankCardResponse struct {
	ErrResponse
	// Number is the bank card number.
	Number string `json:"number,omitempty"`
}

// OCRBankCard recognises the number of a bank card.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_bankcardocr.html
func (w *MiniProgram) OCRBankCard(ctx context.Context, req *OCRImgRequest) (*OCRBankCardResponse, error) {
	var result OCRBankCardResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/bankcard", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OCRDrivingLicenseResponse is the recognised content of a driving licence
// (驾驶证).
type OCRDrivingLicenseResponse struct {
	ErrResponse
	// IDNumber of the licence holder.
	IDNumber string `json:"id_num,omitempty"`
	// Name of the licence holder.
	Name string `json:"name,omitempty"`
	// Sex of the licence holder.
	Sex string `json:"sex,omitempty"`
	// Nationality of the licence holder.
	Nationality string `json:"nationality,omitempty"`
	// Address of the licence holder.
	Address string `json:"address,omitempty"`
	// Birthday of the licence holder.
	Birthday string `json:"birth_date,omitempty"`
	// IssueDate of the licence.
	IssueDate string `json:"issue_date,omitempty"`
	// CarClass the licence allows driving.
	CarClass string `json:"car_class,omitempty"`
	// ValidFrom of the licence validity window.
	ValidFrom string `json:"valid_from,omitempty"`
	// ValidTo of the licence validity window.
	ValidTo string `json:"valid_to,omitempty"`
}

// OCRDrivingLicense recognises the content of a driving licence.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_drivinglicenseocr.html
func (w *MiniProgram) OCRDrivingLicense(ctx context.Context, req *OCRImgRequest) (*OCRDrivingLicenseResponse, error) {
	var result OCRDrivingLicenseResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/drivinglicense", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OCRVehicleLicenseResponse is the recognised content of a vehicle licence
// (行驶证).
type OCRVehicleLicenseResponse struct {
	ErrResponse
	// PlateNumber of the vehicle.
	PlateNumber string `json:"plate_num,omitempty"`
	// VehicleType of the vehicle.
	VehicleType string `json:"vehicle_type,omitempty"`
	// Owner of the vehicle.
	Owner string `json:"owner,omitempty"`
	// Address of the owner.
	Address string `json:"addr,omitempty"`
	// Model of the vehicle.
	Model string `json:"model,omitempty"`
	// Vin of the vehicle.
	Vin string `json:"vin,omitempty"`
	// EngineNumber of the vehicle.
	EngineNumber string `json:"engine_num,omitempty"`
	// RegisterDate of the vehicle.
	RegisterDate string `json:"register_date,omitempty"`
	// IssueDate of the licence.
	IssueDate string `json:"issue_date,omitempty"`
}

// OCRVehicleLicense recognises the content of a vehicle licence (行驶证).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_drivingocr.html
func (w *MiniProgram) OCRVehicleLicense(ctx context.Context, req *OCRImgRequest) (*OCRVehicleLicenseResponse, error) {
	var result OCRVehicleLicenseResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/driving", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OCRBusinessLicenseResponse is the recognised content of a business licence.
type OCRBusinessLicenseResponse struct {
	ErrResponse
	// RegisterNumber (统一社会信用代码) of the enterprise.
	RegisterNumber string `json:"reg_num,omitempty"`
	// Serial number of the licence.
	Serial string `json:"serial,omitempty"`
	// LegalRepresentative of the enterprise.
	LegalRepresentative string `json:"legal_representative,omitempty"`
	// EnterpriseName of the enterprise.
	EnterpriseName string `json:"enterprise_name,omitempty"`
	// TypeOfOrganization of the enterprise.
	TypeOfOrganization string `json:"type_of_organization,omitempty"`
	// Address of the enterprise.
	Address string `json:"address,omitempty"`
	// TypeOfEnterprise of the enterprise.
	TypeOfEnterprise string `json:"type_of_enterprise,omitempty"`
	// BusinessScope of the enterprise.
	BusinessScope string `json:"business_scope,omitempty"`
	// RegisteredCapital of the enterprise.
	RegisteredCapital string `json:"registered_capital,omitempty"`
	// PaidInCapital of the enterprise.
	PaidInCapital string `json:"paid_in_capital,omitempty"`
	// ValidPeriod of the licence.
	ValidPeriod string `json:"valid_period,omitempty"`
	// RegisterDate of the enterprise.
	RegisterDate string `json:"registered_date,omitempty"`
}

// OCRBusinessLicense recognises the content of an enterprise business licence.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_bizlicenseocr.html
func (w *MiniProgram) OCRBusinessLicense(ctx context.Context, req *OCRImgRequest) (*OCRBusinessLicenseResponse, error) {
	var result OCRBusinessLicenseResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/bizlicense", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OCRCommonItem is one recognised text region of a printed document.
type OCRCommonItem struct {
	// Position of the recognised text region.
	Position OCRPosition `json:"pos,omitempty"`
	// Text recognised in the region.
	Text string `json:"text,omitempty"`
}

// OCRCommonResponse is the recognised text of a generic printed document.
type OCRCommonResponse struct {
	ErrResponse
	// Items of recognised text.
	Items []OCRCommonItem `json:"items,omitempty"`
	// ImageSize of the source image.
	ImageSize OCRImageSize `json:"img_size,omitempty"`
}

// OCRCommon recognises generic printed text (通用印刷体识别).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/ocr/api_commocr.html
func (w *MiniProgram) OCRCommon(ctx context.Context, req *OCRImgRequest) (*OCRCommonResponse, error) {
	var result OCRCommonResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/ocr/comm", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CVAICropResponse reports the cropped portrait region of an image.
type CVAICropResponse struct {
	ErrResponse
	// Result is the raw AI crop payload.
	Result CVAICropResult `json:"result,omitempty"`
}

// CVAICropResult carries the crop coordinates.
type CVAICropResult struct {
	// LeftTop is the top-left corner of the crop box.
	LeftTop OCRCoordinate `json:"left_top,omitempty"`
	// RightTop is the top-right corner of the crop box.
	RightTop OCRCoordinate `json:"right_top,omitempty"`
	// RightBottom is the bottom-right corner of the crop box.
	RightBottom OCRCoordinate `json:"right_bottom,omitempty"`
	// LeftBottom is the bottom-left corner of the crop box.
	LeftBottom OCRCoordinate `json:"left_bottom,omitempty"`
}

// ImageAICrop performs an AI smart crop of the person/object region of an
// image.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/img/api_imgaicrop.html
func (w *MiniProgram) ImageAICrop(ctx context.Context, imgURL string) (*CVAICropResponse, error) {
	var result CVAICropResponse
	q := url.Values{}
	q.Set("img_url", imgURL)
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/img/aicrop", q, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImageScanQRCodeResponse reports the QR codes detected in an image.
type ImageScanQRCodeResponse struct {
	ErrResponse
	// ImgSize of the source image.
	ImgSize OCRImageSize `json:"img_size,omitempty"`
	// CodeResults of the detected codes.
	CodeResults []QRCodeResult `json:"code_results,omitempty"`
}

// QRCodeResult is one QR code detected by ImageScanQRCode.
type QRCodeResult struct {
	// TypeName of the code (e.g. "QR_CODE").
	TypeName string `json:"type_name,omitempty"`
	// Data encoded in the code.
	Data string `json:"data,omitempty"`
	// Position of the code within the image.
	Position OCRPosition `json:"pos,omitempty"`
}

// ImageScanQRCode detects and decodes QR codes inside an image.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/img/api_imgqrcode.html
func (w *MiniProgram) ImageScanQRCode(ctx context.Context, imgURL string) (*ImageScanQRCodeResponse, error) {
	var result ImageScanQRCodeResponse
	q := url.Values{}
	q.Set("img_url", imgURL)
	if err := w.withAccessToken(ctx, http.MethodPost, "/cv/img/qrcode", q, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImageSuperResolutionResponse carries the upscaled image content.
type ImageSuperResolutionResponse struct {
	ErrResponse
	// MediaID of the resulting high-resolution image (valid for 3 days via
	// the media get API) when the "返回 media_id" mode is used, otherwise the
	// raw image bytes are returned by the endpoint.
	MediaID string `json:"media_id,omitempty"`
}

// ImageSuperResolution upscales an image 4x using AI. It returns the processed
// image bytes directly.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/img-ocr/img/api_imgsuperresolution.html
func (w *MiniProgram) ImageSuperResolution(ctx context.Context, imgURL string) ([]byte, error) {
	q := url.Values{}
	q.Set("img_url", imgURL)
	return w.withAccessTokenRaw(ctx, http.MethodPost, "/cv/img/superresolution", q, defaultReqOptions(), nil)
}

// ImgSecCheckUpload synchronously checks one uploaded image for risky content
// by uploading its bytes (multipart). mediaType is ignored by the server; it
// exists for parity with the docs. Returns a suggestion verdict of "pass",
// "review" or "risky".
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_imgseccheck.html
func (w *MiniProgram) ImgSecCheckUpload(ctx context.Context, filename string, content []byte) (string, error) {
	data, err := w.withAccessTokenUpload(ctx, "/wxa/img_sec_check", nil, defaultReqOptions(), url.Values{}, "media", filename, "", bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	var probe struct {
		ErrResponse
		Result MsgSecCheckResult `json:"result"`
	}
	if err := decodeJSON(data, &probe); err != nil {
		return "", err
	}
	if probe.ErrCode != 0 {
		return "", core.ClassifyBusinessError(probe.ErrCode, probe.ErrMsg, "")
	}
	return probe.Result.Suggestion, nil
}
