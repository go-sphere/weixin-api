package official

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 用户管理 (user management) — profile, tag groups, blacklist.
// ============================================================

// GetUserInfoRequest selects the user whose profile is requested.
type GetUserInfoRequest struct {
	// OpenID of the user.
	OpenID string
	// Lang of the returned profile (zh_CN, zh_TW, en, ...).
	Lang string
}

// UserInfo is the profile of a subscribed user.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
type UserInfo struct {
	ErrResponse
	// Subscribe is 1 when the user is subscribed, 0 otherwise.
	Subscribe int `json:"subscribe"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// Language of the user.
	Language string `json:"language"`
	// SubscribeTime of the subscription (unix seconds).
	SubscribeTime int64 `json:"subscribe_time"`
	// UnionID of the user (when bound to an Open Platform account).
	UnionID string `json:"unionid"`
	// Remark assigned by the account.
	Remark string `json:"remark"`
	// GroupID of the user group.
	GroupID int `json:"groupid"`
	// TagIDList of the user tags.
	TagIDList []int `json:"tagid_list"`
	// SubscribeScene of the subscription (e.g. "ADD_SCENE_QR_CODE").
	SubscribeScene string `json:"subscribe_scene"`
	// QrScene of the QR scene used to subscribe.
	QrScene int64 `json:"qr_scene"`
	// QrSceneStr of the QR scene string.
	QrSceneStr string `json:"qr_scene_str"`
	// Nickname of the user.
	Nickname string `json:"nickname,omitempty"`
	// Sex: 1 male, 2 female, 0 unknown.
	Sex int `json:"sex,omitempty"`
	// Province of the user.
	Province string `json:"province,omitempty"`
	// City of the user.
	City string `json:"city,omitempty"`
	// Country of the user.
	Country string `json:"country,omitempty"`
	// HeadImgURL of the user avatar.
	HeadImgURL string `json:"headimgurl,omitempty"`
	// Privilege list of the user.
	Privilege []string `json:"privilege,omitempty"`
}

// GetUserInfo returns the basic profile of a subscribed user.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
func (oa *OfficialAccount) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*UserInfo, error) {
	query := url.Values{}
	query.Set("openid", req.OpenID)
	if req.Lang != "" {
		query.Set("lang", req.Lang)
	}
	var result UserInfo
	if err := oa.withToken(ctx, http.MethodGet, "/cgi-bin/user/info", query, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchGetUserInfoRequest selects the users whose profiles are requested.
type BatchGetUserInfoRequest struct {
	// UserList of the requested users.
	UserList []UserInfoSelector `json:"user_list"`
}

// UserInfoSelector identifies one user of a batch profile request.
type UserInfoSelector struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// Lang of the profile.
	Lang string `json:"lang,omitempty"`
}

// BatchGetUserInfoResponse is returned by BatchGetUserInfo.
type BatchGetUserInfoResponse struct {
	ErrResponse
	// UserInfoList of the requested profiles.
	UserInfoList []UserInfo `json:"user_info_list,omitempty"`
}

// BatchGetUserInfo returns up to 100 user profiles in one call.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
func (oa *OfficialAccount) BatchGetUserInfo(ctx context.Context, req *BatchGetUserInfoRequest) (*BatchGetUserInfoResponse, error) {
	var result BatchGetUserInfoResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/user/info/batchget", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFollowerOpenIDListResponse is returned by GetFollowerOpenIDList.
type GetFollowerOpenIDListResponse struct {
	ErrResponse
	// Total number of followers.
	Total int `json:"total"`
	// Count of the returned openids.
	Count int `json:"count"`
	// Data of the follower openids.
	Data struct {
		// OpenID of the followers of the page.
		OpenID []string `json:"openid"`
	} `json:"data"`
	// NextOpenID cursor for the next page ("" when done).
	NextOpenID string `json:"next_openid"`
}

// GetFollowerOpenIDList lists the openids of the followers, paged with a
// next_openid cursor.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_the_List_of_Fans.html
func (oa *OfficialAccount) GetFollowerOpenIDList(ctx context.Context, nextOpenID string) (*GetFollowerOpenIDListResponse, error) {
	query := url.Values{}
	query.Set("next_openid", nextOpenID)
	var result GetFollowerOpenIDListResponse
	if err := oa.withToken(ctx, http.MethodGet, "/cgi-bin/user/get", query, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateUserRemark updates the remark of a subscribed user.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Configuring_user_notes.html
func (oa *OfficialAccount) UpdateUserRemark(ctx context.Context, openid, remark string) error {
	body := map[string]string{"openid": openid, "remark": remark}
	return oa.withTokenPost(ctx, "/cgi-bin/user/info/updateremark", nil, core.DefaultRequestOptions(), body, nil)
}

// ============================================================
// 用户标签 (user tag management).
// ============================================================

// UserTag is one user tag.
type UserTag struct {
	// ID of the tag.
	ID int `json:"id"`
	// Name of the tag.
	Name string `json:"name"`
	// Count of users holding the tag.
	Count int `json:"count,omitempty"`
}

// CreateTagRequest names the new tag.
type CreateTagRequest struct {
	// Tag of the new tag.
	Tag struct {
		// Name of the tag (up to 30 chars).
		Name string `json:"name"`
	} `json:"tag"`
}

// CreateTagResponse is returned by CreateTag.
type CreateTagResponse struct {
	ErrResponse
	// Tag of the created tag.
	Tag UserTag `json:"tag"`
}

// CreateTag creates a user tag.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) CreateTag(ctx context.Context, name string) (*CreateTagResponse, error) {
	req := &CreateTagRequest{}
	req.Tag.Name = name
	var result CreateTagResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/tags/create", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTagListResponse is returned by GetTagList.
type GetTagListResponse struct {
	ErrResponse
	// Tags of the account.
	Tags []UserTag `json:"tags"`
}

// GetTagList lists all user tags of the account.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) GetTagList(ctx context.Context) (*GetTagListResponse, error) {
	var result GetTagListResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/tags/get", nil, core.DefaultRequestOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTagRequest renames a tag.
type UpdateTagRequest struct {
	// Tag of the update.
	Tag struct {
		// ID of the tag.
		ID int `json:"id"`
		// Name of the tag.
		Name string `json:"name"`
	} `json:"tag"`
}

// UpdateTag renames a user tag.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) UpdateTag(ctx context.Context, id int, name string) error {
	req := &UpdateTagRequest{}
	req.Tag.ID = id
	req.Tag.Name = name
	return oa.withTokenPost(ctx, "/cgi-bin/tags/update", nil, core.DefaultRequestOptions(), req, nil)
}

// DeleteTagRequest removes a tag.
type DeleteTagRequest struct {
	// Tag of the delete.
	Tag struct {
		// ID of the tag.
		ID int `json:"id"`
	} `json:"tag"`
}

// DeleteTag deletes a user tag and removes it from all users.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) DeleteTag(ctx context.Context, id int) error {
	req := &DeleteTagRequest{}
	req.Tag.ID = id
	return oa.withTokenPost(ctx, "/cgi-bin/tags/delete", nil, core.DefaultRequestOptions(), req, nil)
}

// BatchTagUsersRequest tags a batch of users.
type BatchTagUsersRequest struct {
	// TagID of the tag.
	TagID int `json:"tagid"`
	// OpenIDList of the users (up to 50 per call).
	OpenIDList []string `json:"openid_list"`
}

// BatchTagUsers assigns a tag to up to 50 users.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) BatchTagUsers(ctx context.Context, req *BatchTagUsersRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/tags/members/batchtagging", nil, core.DefaultRequestOptions(), req, nil)
}

// BatchUntagUsers removes a tag from up to 50 users.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) BatchUntagUsers(ctx context.Context, req *BatchTagUsersRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/tags/members/batchuntagging", nil, core.DefaultRequestOptions(), req, nil)
}

// GetTagUserListRequest pages the users of a tag.
type GetTagUserListRequest struct {
	// TagID of the tag.
	TagID int `json:"tagid"`
	// NextOpenID cursor.
	NextOpenID string `json:"next_openid,omitempty"`
}

// GetTagUserListResponse is returned by GetTagUserList.
type GetTagUserListResponse struct {
	ErrResponse
	// Count of the returned openids.
	Count int `json:"count"`
	// Data of the follower openids.
	Data struct {
		OpenID []string `json:"openid"`
	} `json:"data"`
	// NextOpenID cursor.
	NextOpenID string `json:"next_openid"`
}

// GetTagUserList lists the openids holding a tag, paged.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) GetTagUserList(ctx context.Context, req *GetTagUserListRequest) (*GetTagUserListResponse, error) {
	var result GetTagUserListResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/user/tag/get", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserTagsResponse is returned by GetUserTags.
type GetUserTagsResponse struct {
	ErrResponse
	// TagIDList of the user's tags.
	TagIDList []int `json:"tagid_list"`
}

// GetUserTags returns the tags held by one user.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (oa *OfficialAccount) GetUserTags(ctx context.Context, openid string) (*GetUserTagsResponse, error) {
	var result GetUserTagsResponse
	body := map[string]string{"openid": openid}
	if err := oa.withTokenPost(ctx, "/cgi-bin/tags/getidlist", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// 黑名单 (blacklist).
// ============================================================

// GetBlackListResponse is returned by GetBlackList.
type GetBlackListResponse struct {
	ErrResponse
	// Total number of blacklisted users.
	Total int `json:"total"`
	// Count of the returned openids.
	Count int `json:"count"`
	// Data of the blacklisted openids.
	Data struct {
		OpenID []string `json:"openid"`
	} `json:"data"`
	// NextOpenID cursor.
	NextOpenID string `json:"next_openid"`
}

// GetBlackList pages the blacklisted users.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Manage_blacklist.html
func (oa *OfficialAccount) GetBlackList(ctx context.Context, beginOpenID string) (*GetBlackListResponse, error) {
	var result GetBlackListResponse
	body := map[string]string{"begin_openid": beginOpenID}
	if err := oa.withTokenPost(ctx, "/cgi-bin/tags/members/getblacklist", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchBlacklistUsersRequest adds/removes users to/from the blacklist.
type BatchBlacklistUsersRequest struct {
	// OpenIDList of the users.
	OpenIDList []string `json:"openid_list"`
}

// BatchBlacklistUsers adds users to the blacklist.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Manage_blacklist.html
func (oa *OfficialAccount) BatchBlacklistUsers(ctx context.Context, openids []string) error {
	req := &BatchBlacklistUsersRequest{OpenIDList: openids}
	return oa.withTokenPost(ctx, "/cgi-bin/tags/members/batchblacklist", nil, core.DefaultRequestOptions(), req, nil)
}

// BatchUnblacklistUsers removes users from the blacklist.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/User_Management/Manage_blacklist.html
func (oa *OfficialAccount) BatchUnblacklistUsers(ctx context.Context, openids []string) error {
	req := &BatchBlacklistUsersRequest{OpenIDList: openids}
	return oa.withTokenPost(ctx, "/cgi-bin/tags/members/batchunblacklist", nil, core.DefaultRequestOptions(), req, nil)
}
