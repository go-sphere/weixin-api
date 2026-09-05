package official

import (
	"context"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 自定义菜单 (custom menu) — buttons with click/view/miniprogram actions.
// ============================================================

// MenuButtonType is the action type of a menu button.
type MenuButtonType string

const (
	// MenuButtonClick sends a click event with the button key.
	MenuButtonClick MenuButtonType = "click"
	// MenuButtonView opens a URL.
	MenuButtonView MenuButtonType = "view"
	// MenuButtonMiniProgram opens a Mini Program page (服务号).
	MenuButtonMiniProgram MenuButtonType = "miniprogram"
	// MenuButtonScanCodePush opens the scanner.
	MenuButtonScanCodePush MenuButtonType = "scancode_push"
	// MenuButtonScanCodeWaitMsg scans and waits for a message.
	MenuButtonScanCodeWaitMsg MenuButtonType = "scancode_waitmsg"
	// MenuButtonPicSysPhoto opens the system camera.
	MenuButtonPicSysPhoto MenuButtonType = "pic_sysphoto"
	// MenuButtonPicPhotoOrAlbum opens photo or album.
	MenuButtonPicPhotoOrAlbum MenuButtonType = "pic_photo_or_album"
	// MenuButtonPicWeixin opens the WeChat album.
	MenuButtonPicWeixin MenuButtonType = "pic_weixin"
	// MenuButtonLocationSelect opens the location picker.
	MenuButtonLocationSelect MenuButtonType = "location_select"
	// MenuButtonMediaID opens a media (article) message.
	MenuButtonMediaID MenuButtonType = "media_id"
	// MenuButtonViewLimited opens a limited article URL.
	MenuButtonViewLimited MenuButtonType = "view_limited"
)

// MenuButton is one button of the custom menu.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Custom_Menus/Creating_Custom-Defined_Menu.html
type MenuButton struct {
	// Type of the button.
	Type MenuButtonType `json:"type,omitempty"`
	// Name of the button (up to 16 chars for a root button, 40 for a
	// sub-button).
	Name string `json:"name"`
	// Key of a click button.
	Key string `json:"key,omitempty"`
	// URL of a view button.
	URL string `json:"url,omitempty"`
	// MediaID of a media_id / view_limited button.
	MediaID string `json:"media_id,omitempty"`
	// AppID of a miniprogram button.
	AppID string `json:"appid,omitempty"`
	// PagePath of a miniprogram button.
	PagePath string `json:"pagepath,omitempty"`
	// SubButton of a parent button (up to 5).
	SubButton []MenuButton `json:"sub_button,omitempty"`
}

// CreateMenuRequest is the payload of the custom menu.
type CreateMenuRequest struct {
	// Button of the menu (1-3 root buttons).
	Button []MenuButton `json:"button"`
}

// CreateMenu creates or replaces the custom menu.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Custom_Menus/Creating_Custom-Defined_Menu.html
func (oa *OfficialAccount) CreateMenu(ctx context.Context, req *CreateMenuRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/menu/create", nil, core.DefaultRequestOptions(), req, nil)
}

// GetMenuResponse is returned by GetMenu.
type GetMenuResponse struct {
	ErrResponse
	// Menu of the current menu.
	Menu struct {
		// Button of the menu.
		Button []MenuButton `json:"button"`
	} `json:"menu"`
}

// GetMenu returns the current custom menu.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Custom_Menus/Querying_custom_menus.html
func (oa *OfficialAccount) GetMenu(ctx context.Context) (*GetMenuResponse, error) {
	var result GetMenuResponse
	if err := oa.withToken(ctx, "GET", "/cgi-bin/menu/get", nil, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMenu removes the custom menu.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Custom_Menus/Deleting_custom_menus.html
func (oa *OfficialAccount) DeleteMenu(ctx context.Context) error {
	return oa.withTokenPost(ctx, "/cgi-bin/menu/delete", nil, core.DefaultRequestOptions(), map[string]string{}, nil)
}

// GetCurrentSelfMenuResponse is returned by GetCurrentSelfMenu.
type GetCurrentSelfMenuResponse struct {
	ErrResponse
	// IsMenuOpen of the menu.
	IsMenuOpen int `json:"is_menu_open"`
	// SelfMenuInfo of the current menu.
	SelfMenuInfo struct {
		// Button of the menu.
		Button []MenuButton `json:"button"`
	} `json:"selfmenu_info"`
}

// GetCurrentSelfMenu returns the currently active menu (including conditional
// menus).
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Custom_Menus/Querying_custom_menus.html
func (oa *OfficialAccount) GetCurrentSelfMenu(ctx context.Context) (*GetCurrentSelfMenuResponse, error) {
	var result GetCurrentSelfMenuResponse
	if err := oa.withToken(ctx, "GET", "/cgi-bin/get_current_selfmenu_info", nil, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
