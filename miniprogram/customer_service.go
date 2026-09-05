package miniprogram

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// KfAccountItem describes one customer-service account.
type KfAccountItem struct {
	// ID (kf_account) is the account name like "test@kftest".
	ID string `json:"kf_account"`
	// Nickname of the account.
	Nickname string `json:"kf_nick"`
	// AvatarURL of the account avatar (absolute URL).
	AvatarURL string `json:"kf_headimgurl"`
	// Weixin bound WeChat id of the account.
	Weixin string `json:"kf_wx,omitempty"`
}

// AddKfAccountRequest is the payload of AddKfAccount.
type AddKfAccountRequest struct {
	// ID is the kf account id, format "prefix@alias", 1-30 chars.
	ID string `json:"kf_account"`
	// Nickname of the account (required).
	Nickname string `json:"nickname"`
}

// AddKfAccount adds a new customer-service account bound to a WeChat user who
// will staff the account. The WeChat user must accept the binding in the
// customer-service console.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_addkfaccount.html
func (w *MiniProgram) AddKfAccount(ctx context.Context, req *AddKfAccountRequest) error {
	return w.withAccessTokenPost(ctx, "/customservice/kfaccount/add", nil, defaultReqOptions(), req, nil)
}

// UpdateKfAccountRequest is the payload of UpdateKfAccount.
type UpdateKfAccountRequest struct {
	// ID of the account to update.
	ID string `json:"kf_account"`
	// Nickname to assign.
	Nickname string `json:"nickname"`
}

// UpdateKfAccount changes the nickname or rebinds a customer-service account to
// another WeChat user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_addkfaccount.html
func (w *MiniProgram) UpdateKfAccount(ctx context.Context, req *UpdateKfAccountRequest) error {
	return w.withAccessTokenPost(ctx, "/customservice/kfaccount/update", nil, defaultReqOptions(), req, nil)
}

// DeleteKfAccount removes a customer-service account (identified by its kf
// account id and the WeChat user bound to it).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_delkfaccount.html
func (w *MiniProgram) DeleteKfAccount(ctx context.Context, kfAccount, weixin string) error {
	body := map[string]string{
		"kf_account": kfAccount,
		"kf_wx":      weixin,
	}
	return w.withAccessTokenPost(ctx, "/customservice/kfaccount/del", nil, defaultReqOptions(), body, nil)
}

// GetKfAccountListResponse is returned by GetKfAccountList.
type GetKfAccountListResponse struct {
	ErrResponse
	// KfList is the list of all configured customer-service accounts.
	KfList []KfAccountItem `json:"kf_list"`
}

// GetKfAccountList returns all customer-service accounts of the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_getkflist.html
func (w *MiniProgram) GetKfAccountList(ctx context.Context) (*GetKfAccountListResponse, error) {
	var result GetKfAccountListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/customservice/getkflist", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// KfAccountOnlineItem is one currently-online customer-service account.
type KfAccountOnlineItem struct {
	// ID of the account.
	ID string `json:"kf_account"`
	// Status: 1 means the WeChat staff is online, 2 means offline.
	Status int `json:"status"`
	// AutoAccept is 1 when the account auto-accepts new sessions.
	AutoAccept int `json:"auto_accept"`
	// AcceptedCase counts the sessions being handled.
	AcceptedCase int `json:"accepted_case"`
}

// GetOnlineKfListResponse is returned by GetOnlineKfList.
type GetOnlineKfListResponse struct {
	ErrResponse
	// KfOnlineList lists the online accounts.
	KfOnlineList []KfAccountOnlineItem `json:"kf_online_list"`
}

// GetOnlineKfList returns the customer-service accounts whose staff are
// currently online.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_getonlinekflist.html
func (w *MiniProgram) GetOnlineKfList(ctx context.Context) (*GetOnlineKfListResponse, error) {
	var result GetOnlineKfListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/customservice/getonlinekflist", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InviteKfWorkerRequest invites a WeChat user to staff a customer-service
// account.
type InviteKfWorkerRequest struct {
	// ID of the customer-service account.
	ID string `json:"kf_account"`
	// InviteWeixin is the WeChat id (wxid) of the user to invite.
	InviteWeixin string `json:"invite_wx"`
}

// InviteKfWorker invites a WeChat user to become the staff of an existing
// customer-service account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_addkfaccount.html
func (w *MiniProgram) InviteKfWorker(ctx context.Context, req *InviteKfWorkerRequest) error {
	return w.withAccessTokenPost(ctx, "/customservice/kfaccount/inviteworker", nil, defaultReqOptions(), req, nil)
}

// SetKfAdmin promotes a customer-service account to admin. kfOpenID is the
// openid of the WeChat user staffing the account (the admin-openid contract).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_setkfadmin.html
func (w *MiniProgram) SetKfAdmin(ctx context.Context, kfOpenID string) error {
	query := url.Values{}
	query.Set("kf_openid", kfOpenID)
	return w.withAccessToken(ctx, http.MethodGet, "/customservice/kfaccount/setadmin", query, defaultReqOptions(), nil, nil)
}

// CancelKfAdmin revokes the admin role of a customer-service account. kfOpenID
// is the openid of the WeChat user staffing the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_cancelkfadmin.html
func (w *MiniProgram) CancelKfAdmin(ctx context.Context, kfOpenID string) error {
	query := url.Values{}
	query.Set("kf_openid", kfOpenID)
	return w.withAccessToken(ctx, http.MethodGet, "/customservice/kfaccount/canceladmin", query, defaultReqOptions(), nil, nil)
}

// KfMsgText is a plain text customer-service message.
type KfMsgText struct {
	// Content of the message.
	Content string `json:"content"`
}

// KfMsgImage is an image customer-service message referencing a media_id.
type KfMsgImage struct {
	// MediaID of the uploaded image media.
	MediaID string `json:"media_id"`
}

// KfMsgLink is a hyperlink customer-service message.
type KfMsgLink struct {
	// Title of the link.
	Title string `json:"title"`
	// Description of the link.
	Description string `json:"description"`
	// URL of the link.
	URL string `json:"url"`
	// ThumbURL of the link thumbnail.
	ThumbURL string `json:"thumb_url"`
}

// KfMsgMiniprogramPage is a Mini Program page card customer-service message.
type KfMsgMiniprogramPage struct {
	// Title of the card.
	Title string `json:"title"`
	// PagePath of the Mini Program page to open.
	PagePath string `json:"pagepath"`
	// ThumbMediaID of the cover image (must be an uploaded media).
	ThumbMediaID string `json:"thumb_media_id"`
}

// KfMsgMenu is a menu customer-service message.
type KfMsgMenu struct {
	// HeadContent of the menu message.
	HeadContent string `json:"head_content"`
	// List of the menu items.
	List []KfMsgMenuItem `json:"list"`
	// TailContent of the menu message.
	TailContent string `json:"tail_content"`
}

// KfMsgMenuItem is one option of a menu message.
type KfMsgMenuItem struct {
	// ID returned by the client menu callback to identify the choice.
	ID string `json:"id"`
	// Content of the menu item.
	Content string `json:"content"`
}

// SendKfMessageRequest is the payload of SendKfMessage. Only one of the Msg
// variants must be set; Text, Image, Link, MiniprogramPage and Menu are
// mutually exclusive.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_sendcustommessage.html
type SendKfMessageRequest struct {
	// ToUser is the openid of the user.
	ToUser string `json:"touser"`
	// MsgType is "text", "image", "link", "miniprogrampage" or "msgmenu".
	MsgType string `json:"msgtype"`
	// Text payload, set when MsgType is "text".
	Text *KfMsgText `json:"text,omitempty"`
	// Image payload, set when MsgType is "image".
	Image *KfMsgImage `json:"image,omitempty"`
	// Link payload, set when MsgType is "link".
	Link *KfMsgLink `json:"link,omitempty"`
	// MiniprogramPage payload, set when MsgType is "miniprogrampage".
	MiniprogramPage *KfMsgMiniprogramPage `json:"miniprogrampage,omitempty"`
	// Menu payload, set when MsgType is "msgmenu".
	Menu *KfMsgMenu `json:"msgmenu,omitempty"`
}

// SendKfMessage delivers a customer-service message to a user. The user must
// have talked to the Mini Program within 48 hours.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_sendcustommessage.html
func (w *MiniProgram) SendKfMessage(ctx context.Context, req *SendKfMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/custom/send", nil, defaultReqOptions(), req, nil)
}

// SetKfTyping tells the user that a customer-service staff is typing. Send it
// before a real message and cancel it after sending by passing
// command="CancelTyping".
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_typing.html
func (w *MiniProgram) SetKfTyping(ctx context.Context, toUser, command string) error {
	body := map[string]string{
		"touser":  toUser,
		"command": command,
	}
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/custom/business/typing", nil, defaultReqOptions(), body, nil)
}

// UploadKfMedia uploads a temporary media for customer-service image messages
// and returns its media_id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_uploadtempmedia.html
func (w *MiniProgram) UploadKfMedia(ctx context.Context, mediaType string, filename string, content []byte) (string, error) {
	form := url.Values{}
	form.Set("type", mediaType)
	data, err := w.withAccessTokenUpload(ctx, "/cgi-bin/media/upload", nil, defaultReqOptions(), form, "media", filename, "", bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	var result struct {
		ErrResponse
		MediaID string `json:"media_id"`
	}
	if err := decodeJSON(data, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", core.ClassifyBusinessError(result.ErrCode, result.ErrMsg, "")
	}
	return result.MediaID, nil
}

// DownloadKfMedia fetches a previously uploaded customer-service media by its
// media_id and returns the raw bytes.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-message/api_getmedia.html
func (w *MiniProgram) DownloadKfMedia(ctx context.Context, mediaID string) ([]byte, error) {
	query := url.Values{}
	query.Set("media_id", mediaID)
	data, err := w.withAccessTokenRaw(ctx, http.MethodGet, "/cgi-bin/media/get", query, defaultReqOptions(), nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// KfWorkInfo describes the customer-service staff members bound to the current
// account group (微信客服, "WeCom KF" — distinct from the classic Mini Program
// customer service).
type KfWorkInfo struct {
	// CorpName of the bound WeCom enterprise.
	CorpName string `json:"corp_name"`
	// UserID is the WeCom user id of the staff member.
	UserID string `json:"userid"`
	// Name of the staff member.
	Name string `json:"name"`
	// Avatar of the staff member.
	Avatar string `json:"avatar,omitempty"`
	// QrCode is used when the staff replies in a Mini Program session.
	QrCode string `json:"qr_code,omitempty"`
}

// GetKfWorkBoundResponse is returned by GetKfWorkBound.
type GetKfWorkBoundResponse struct {
	ErrResponse
	// KfList is the list of bound WeCom customer-service staff.
	KfList []KfWorkInfo `json:"kf_list"`
}

// GetKfWorkBound returns the WeCom (企业微信) customer-service account bound to
// the Mini Program. The bound enterprise must share the Mini Program's entity.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-work/api_getkfworkbound.html
func (w *MiniProgram) GetKfWorkBound(ctx context.Context) (*GetKfWorkBoundResponse, error) {
	var result GetKfWorkBoundResponse
	if err := w.withAccessTokenPost(ctx, "/customservice/work/get", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BindKfWork binds a WeCom (企业微信) enterprise to the Mini Program so that its
// 微信客服 accounts can serve the Mini Program users. The enterprise id must
// have the same legal entity as the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-work/api_bindkfwork.html
func (w *MiniProgram) BindKfWork(ctx context.Context, corpID string) error {
	body := map[string]string{"corpid": corpID}
	return w.withAccessTokenPost(ctx, "/customservice/work/bind", nil, defaultReqOptions(), body, nil)
}

// UnbindKfWork unbinds a WeCom (企业微信) enterprise from the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-work/api_unbindkfwork.html
func (w *MiniProgram) UnbindKfWork(ctx context.Context, corpID string) error {
	body := map[string]string{"corpid": corpID}
	return w.withAccessTokenPost(ctx, "/customservice/work/unbind", nil, defaultReqOptions(), body, nil)
}
