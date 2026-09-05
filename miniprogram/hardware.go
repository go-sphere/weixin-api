package miniprogram

import (
	"context"
)

// HardwareDevice (微信硬件设备) groups a set of IoT device management APIs built
// around device model ids and serial numbers. All paths below live under the
// Mini Program wxa/business namespace.

// DeviceID uniquely identifies a hardware device by its model and serial.
type DeviceID struct {
	// ModelID of the device model.
	ModelID string `json:"model_id"`
	// SN is the serial number of the device.
	SN string `json:"sn"`
}

// HardwareLicenseDeviceRequest activates licensed devices (pkg_type selects the
// license package type purchased).
type HardwareLicenseDeviceRequest struct {
	// DeviceList of the devices to activate.
	DeviceList []LicenseActivationDevice `json:"device_list"`
	// PkgType of the license package to consume.
	PkgType int `json:"pkg_type,omitempty"`
}

// LicenseActivationDevice is one device to activate, with the number of
// licenses it consumes.
type LicenseActivationDevice struct {
	DeviceID
	// ActiveNumber of licenses to activate for the device.
	ActiveNumber int `json:"active_number,omitempty"`
}

// ActivateLicenseDevice activates one or more hardware devices with the
// purchased license packages.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_activelicensedevice.html
func (w *MiniProgram) ActivateLicenseDevice(ctx context.Context, req *HardwareLicenseDeviceRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/license/activedevice", nil, defaultReqOptions(), req, nil)
}

// GetLicenseDeviceInfoRequest selects the devices to query.
type GetLicenseDeviceInfoRequest struct {
	// DeviceList of the devices to query.
	DeviceList []DeviceID `json:"device_list"`
}

// GetLicenseDeviceInfoResponse is returned by GetLicenseDeviceInfo.
type GetLicenseDeviceInfoResponse struct {
	ErrResponse
	// DeviceList of the queried devices with their activation state.
	DeviceList []LicenseDeviceStatus `json:"device_list,omitempty"`
}

// LicenseDeviceStatus is the activation state of one device.
type LicenseDeviceStatus struct {
	DeviceID
	// ActiveNumber of licenses consumed by the device.
	ActiveNumber int `json:"active_number,omitempty"`
	// DeviceStatus of the device.
	DeviceStatus int `json:"device_status,omitempty"`
}

// GetLicenseDeviceInfo queries the activation state of licensed devices.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_getlicensedeviceinfo.html
func (w *MiniProgram) GetLicenseDeviceInfo(ctx context.Context, req *GetLicenseDeviceInfoRequest) (*GetLicenseDeviceInfoResponse, error) {
	var result GetLicenseDeviceInfoResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/license/getdeviceinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LicensePackage describes one license package type.
type LicensePackage struct {
	// PkgType of the license package.
	PkgType int `json:"pkg_type"`
	// PkgName of the package.
	PkgName string `json:"pkg_name,omitempty"`
	// TotalNum of licenses purchased.
	TotalNum int `json:"total_num,omitempty"`
	// RemainNum of unused licenses.
	RemainNum int `json:"remain_num,omitempty"`
	// UsedNum of consumed licenses.
	UsedNum int `json:"used_num,omitempty"`
}

// GetLicensePkgListResponse is returned by GetLicensePkgList.
type GetLicensePkgListResponse struct {
	ErrResponse
	// PackageList of the purchased license packages.
	PackageList []LicensePackage `json:"pkg_list,omitempty"`
}

// GetLicensePkgList lists the hardware-device license packages of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_getlicensepkglist.html
func (w *MiniProgram) GetLicensePkgList(ctx context.Context) (*GetLicensePkgListResponse, error) {
	var result GetLicensePkgListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/license/getpkglist", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSNTicketRequest selects the device whose SN sticker ticket is needed.
type GetSNTicketRequest struct {
	// SN of the device.
	SN string `json:"sn,omitempty"`
	// ModelID of the device.
	ModelID string `json:"model_id,omitempty"`
}

// GetSNTicketResponse is returned by GetSNTicket.
type GetSNTicketResponse struct {
	ErrResponse
	// Ticket for printing the device SN sticker.
	Ticket string `json:"ticket,omitempty"`
}

// GetSNTicket requests a ticket for printing a hardware device SN sticker.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_getsnticket.html
func (w *MiniProgram) GetSNTicket(ctx context.Context, req *GetSNTicketRequest) (*GetSNTicketResponse, error) {
	var result GetSNTicketResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/getsnticket", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateIoTGroupRequest creates an IoT device group.
type CreateIoTGroupRequest struct {
	// ModelID of the device model the group manages.
	ModelID string `json:"model_id"`
	// GroupName of the new group.
	GroupName string `json:"group_name"`
}

// CreateIoTGroupResponse is returned by CreateIoTGroup.
type CreateIoTGroupResponse struct {
	ErrResponse
	// GroupID of the created group.
	GroupID string `json:"group_id,omitempty"`
}

// CreateIoTGroup creates an IoT device group used to organise devices of a
// model.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_createiotgroupid.html
func (w *MiniProgram) CreateIoTGroup(ctx context.Context, req *CreateIoTGroupRequest) (*CreateIoTGroupResponse, error) {
	var result CreateIoTGroupResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/group/createid", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IoTGroupDeviceRequest adds/removes devices of an IoT group.
type IoTGroupDeviceRequest struct {
	// GroupID of the IoT group.
	GroupID string `json:"group_id"`
	// DeviceList of the devices to add/remove.
	DeviceList []DeviceID `json:"device_list"`
}

// AddIoTGroupDevice adds devices into an IoT group.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_addiotgroupdevice.html
func (w *MiniProgram) AddIoTGroupDevice(ctx context.Context, req *IoTGroupDeviceRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/group/adddevice", nil, defaultReqOptions(), req, nil)
}

// RemoveIoTGroupDevice removes devices from an IoT group.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_removeiotgroupdevice.html
func (w *MiniProgram) RemoveIoTGroupDevice(ctx context.Context, req *IoTGroupDeviceRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/group/removedevice", nil, defaultReqOptions(), req, nil)
}

// GetIoTGroupInfoRequest selects an IoT group to query.
type GetIoTGroupInfoRequest struct {
	// GroupID of the IoT group.
	GroupID string `json:"group_id"`
}

// GetIoTGroupInfoResponse describes an IoT group.
type GetIoTGroupInfoResponse struct {
	ErrResponse
	// GroupInfo of the queried group.
	GroupInfo IoTGroupInfo `json:"group_info,omitempty"`
}

// IoTGroupInfo is the detailed state of an IoT group.
type IoTGroupInfo struct {
	// GroupID of the group.
	GroupID string `json:"group_id,omitempty"`
	// ModelID of the device model.
	ModelID string `json:"model_id,omitempty"`
	// GroupName of the group.
	GroupName string `json:"group_name,omitempty"`
	// GroupDesc of the group.
	GroupDesc string `json:"group_desc,omitempty"`
	// DeviceList of the group members.
	DeviceList []DeviceID `json:"device_list,omitempty"`
}

// GetIoTGroupInfo queries one IoT group and its devices.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_getiotgroupinfo.html
func (w *MiniProgram) GetIoTGroupInfo(ctx context.Context, groupID string) (*GetIoTGroupInfoResponse, error) {
	var result GetIoTGroupInfoResponse
	req := &GetIoTGroupInfoRequest{GroupID: groupID}
	if err := w.withAccessTokenPost(ctx, "/wxa/business/group/getinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// HardwareDeviceMessageRequest sends a device-subscribe template message to
// the owners of a device.
type HardwareDeviceMessageRequest struct {
	// TemplateID of the device subscription template.
	TemplateID string `json:"template_id"`
	// SN of the device (unique serial assigned by the vendor, up to 128
	// bytes of digits/letters/_/-).
	SN string `json:"sn"`
	// Page to jump to inside the Mini Program when the message is tapped.
	Page string `json:"page"`
	// ToOpenIDList of the receiving users.
	ToOpenIDList []string `json:"to_openid_list"`
	// MiniprogramState: developer / trial / formal.
	MiniprogramState string `json:"miniprogram_state,omitempty"`
	// ModelID of the device model (from device registration).
	ModelID string `json:"modelId"`
	// Data of the template values.
	Data map[string]any `json:"data"`
	// Lang of the "view in Mini Program" entry.
	Lang string `json:"lang,omitempty"`
}

// SendHardwareDeviceMessage delivers a device subscription message (设备订阅
// 消息) to the device owners.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/hardware-device/api_sendhardwaredevicemessage.html
func (w *MiniProgram) SendHardwareDeviceMessage(ctx context.Context, req *HardwareDeviceMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/device/subscribe/send", nil, defaultReqOptions(), req, nil)
}
