package miniprogram

import (
	"context"
)

// ============================================================
// 云开发 (CloudBase / TCB) server APIs.
// Routes live under /tcb/* (database, functions, storage, SMS) plus /wxa
// for voip sign and open data.
// ============================================================

// TCBQuery is the shared {env, query} body of the database operation APIs.
type TCBQuery struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// Query is the JSON database operation statement.
	Query string `json:"query"`
}

// TCBDatabaseItemResponse is the common response of the database item APIs.
type TCBDatabaseItemResponse struct {
	ErrResponse
	// RawData of the response payload.
	RawData map[string]any `json:"-"`
}

// TCBAddDatabaseRecord executes a database "add" statement against a
// collection.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_adddatabaseitem.html
func (w *MiniProgram) TCBAddDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databaseadd", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBQueryDatabaseRecord queries database records.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_getdatabaserecord.html
func (w *MiniProgram) TCBQueryDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasequery", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBUpdateDatabaseRecord updates database records matching the query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_updatedatabaserecord.html
func (w *MiniProgram) TCBUpdateDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databaseupdate", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBDeleteDatabaseRecord deletes database records matching the query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_deletedatabaseitem.html
func (w *MiniProgram) TCBDeleteDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasedelete", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBCountDatabaseRecord counts records matching the query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_getdatabasecount.html
func (w *MiniProgram) TCBCountDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasecount", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBAggregateDatabaseRecord runs an aggregation pipeline.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_aggregatedatabase.html
func (w *MiniProgram) TCBAggregateDatabaseRecord(ctx context.Context, req *TCBQuery) (*TCBDatabaseItemResponse, error) {
	var result TCBDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databaseaggregate", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBCollectionRequest is the shared collection body.
type TCBCollectionRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// CollectionName of the collection.
	CollectionName string `json:"collection_name"`
}

// TCBAddCollection creates a database collection.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_adddatabasecollection.html
func (w *MiniProgram) TCBAddCollection(ctx context.Context, env, collectionName string) error {
	req := &TCBCollectionRequest{Env: env, CollectionName: collectionName}
	return w.withAccessTokenPost(ctx, "/tcb/databasecollectionadd", nil, defaultReqOptions(), req, nil)
}

// TCBDeleteCollection deletes a database collection.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_deletedatabasecollection.html
func (w *MiniProgram) TCBDeleteCollection(ctx context.Context, env, collectionName string) error {
	req := &TCBCollectionRequest{Env: env, CollectionName: collectionName}
	return w.withAccessTokenPost(ctx, "/tcb/databasecollectiondelete", nil, defaultReqOptions(), req, nil)
}

// TCBGetCollectionResponse is returned by TCBGetCollection.
type TCBGetCollectionResponse struct {
	ErrResponse
	// RawData of the collection list payload.
	RawData map[string]any `json:"-"`
}

// TCBGetCollectionRequest pages the collections of an environment.
type TCBGetCollectionRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// Limit of the result count.
	Limit int `json:"limit,omitempty"`
	// Offset of the result.
	Offset int `json:"offset,omitempty"`
}

// TCBGetCollection lists the collections of a cloud environment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_getdatabasecollection.html
func (w *MiniProgram) TCBGetCollection(ctx context.Context, req *TCBGetCollectionRequest) (*TCBGetCollectionResponse, error) {
	var result TCBGetCollectionResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasecollectionget", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBImportDatabaseItemRequest imports data into a collection from a cloud
// storage file.
type TCBImportDatabaseItemRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// CollectionName receiving the data.
	CollectionName string `json:"collection_name"`
	// FilePath of the source file in the environment storage.
	FilePath string `json:"file_path"`
	// FileType: 1 JSON, 2 CSV.
	FileType int `json:"file_type"`
	// StopOnError aborts the import on the first error when true.
	StopOnError bool `json:"stop_on_error"`
	// ConflictMode: 1 INSERT, 2 UPSERT.
	ConflictMode int `json:"conflict_mode"`
}

// TCBImportDatabaseItemResponse is returned by TCBImportDatabaseItem.
type TCBImportDatabaseItemResponse struct {
	ErrResponse
	// JobID of the import migration task.
	JobID int64 `json:"job_id,omitempty"`
}

// TCBImportDatabaseItem starts an async import of data into a collection.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_importdatabaseitem.html
func (w *MiniProgram) TCBImportDatabaseItem(ctx context.Context, req *TCBImportDatabaseItemRequest) (*TCBImportDatabaseItemResponse, error) {
	var result TCBImportDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasemigrateimport", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBExportDatabaseItemRequest exports a collection to a storage file.
type TCBExportDatabaseItemRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// FilePath of the exported file in the environment public storage.
	FilePath string `json:"file_path"`
	// FileType of the exported file.
	FileType int `json:"file_type"`
	// Query of the export condition.
	Query string `json:"query"`
}

// TCBExportDatabaseItemResponse is returned by TCBExportDatabaseItem.
type TCBExportDatabaseItemResponse struct {
	ErrResponse
	// JobID of the export migration task.
	JobID int64 `json:"job_id,omitempty"`
}

// TCBExportDatabaseItem starts an async export of collection data.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_exportdatabaseitem.html
func (w *MiniProgram) TCBExportDatabaseItem(ctx context.Context, req *TCBExportDatabaseItemRequest) (*TCBExportDatabaseItemResponse, error) {
	var result TCBExportDatabaseItemResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/databasemigrateexport", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBQueryMigrateStatusResponse is returned by TCBQueryMigrateStatus.
type TCBQueryMigrateStatusResponse struct {
	ErrResponse
	// Status of the migration task.
	Status string `json:"status,omitempty"`
	// RawData of the response payload.
	RawData map[string]any `json:"-"`
}

// TCBQueryMigrateStatus queries an import/export migration job.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_getdatabasemigratestatus.html
func (w *MiniProgram) TCBQueryMigrateStatus(ctx context.Context, env string, jobID int64) (*TCBQueryMigrateStatusResponse, error) {
	var result TCBQueryMigrateStatusResponse
	body := map[string]any{"env": env, "job_id": jobID}
	if err := w.withAccessTokenPost(ctx, "/tcb/databasemigratequeryinfo", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBUpdateIndexRequest updates the indexes of a collection.
type TCBUpdateIndexRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// CollectionName of the collection.
	CollectionName string `json:"collection_name"`
	// CreateIndexes to add.
	CreateIndexes []map[string]any `json:"create_indexes"`
	// DropIndexes to remove.
	DropIndexes []map[string]any `json:"drop_indexes"`
}

// TCBUpdateIndex updates the database indexes of a collection.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/database/api_updatedatabaseindexs.html
func (w *MiniProgram) TCBUpdateIndex(ctx context.Context, req *TCBUpdateIndexRequest) error {
	return w.withAccessTokenPost(ctx, "/tcb/updateindex", nil, defaultReqOptions(), req, nil)
}

// ============================================================
// Cloud functions.
// ============================================================

// TCBInvokeCloudFunctionRequest invokes a cloud function with a JSON payload.
type TCBInvokeCloudFunctionRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// Name of the cloud function.
	Name string `json:"name"`
	// ReqData is the JSON string passed to the function.
	ReqData string `json:"req_data"`
}

// TCBInvokeCloudFunctionResponse is returned by TCBInvokeCloudFunction.
type TCBInvokeCloudFunctionResponse struct {
	ErrResponse
	// Result of the invocation.
	Result string `json:"result,omitempty"`
	// RawData of the response payload.
	RawData map[string]any `json:"-"`
}

// TCBInvokeCloudFunction triggers a cloud function synchronously.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/functions/api_invokecloudfunction.html
func (w *MiniProgram) TCBInvokeCloudFunction(ctx context.Context, req *TCBInvokeCloudFunctionRequest) (*TCBInvokeCloudFunctionResponse, error) {
	var result TCBInvokeCloudFunctionResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/invokecloudfunction", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBAddDelayedFunctionTaskRequest schedules a delayed cloud function call.
type TCBAddDelayedFunctionTaskRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// FunctionName of the function.
	FunctionName string `json:"function_name"`
	// Data is the JSON string payload.
	Data string `json:"data"`
	// DelayTime in seconds (6 seconds to 30 days).
	DelayTime int64 `json:"delay_time"`
}

// TCBAddDelayedFunctionTask schedules a one-off delayed cloud function call.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/functions/api_adddelayedfunctiontask.html
func (w *MiniProgram) TCBAddDelayedFunctionTask(ctx context.Context, req *TCBAddDelayedFunctionTaskRequest) error {
	return w.withAccessTokenPost(ctx, "/tcb/adddelayedfunctiontask", nil, defaultReqOptions(), req, nil)
}

// ============================================================
// Cloud storage.
// ============================================================

// TCBGetUploadFileLinkRequest obtains an upload URL for a storage path.
type TCBGetUploadFileLinkRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// Path of the target file in the storage.
	Path string `json:"path"`
}

// TCBGetUploadFileLinkResponse is returned by TCBGetUploadFileLink.
type TCBGetUploadFileLinkResponse struct {
	ErrResponse
	// RawData of the upload link payload.
	RawData map[string]any `json:"-"`
}

// TCBGetUploadFileLink returns a signed upload URL for a storage path.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/storage/api_getuploadtcbfilelink.html
func (w *MiniProgram) TCBGetUploadFileLink(ctx context.Context, env, path string) (*TCBGetUploadFileLinkResponse, error) {
	req := &TCBGetUploadFileLinkRequest{Env: env, Path: path}
	var result TCBGetUploadFileLinkResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/uploadfile", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBFileDownload is one file of a batch download request.
type TCBFileDownload struct {
	// FileID of the file.
	FileID string `json:"fileid"`
	// MaxAge of the signed URL.
	MaxAge int `json:"max_age,omitempty"`
}

// TCBGetDownloadFileLinkRequest requests download links for files.
type TCBGetDownloadFileLinkRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// FileList of the files.
	FileList []TCBFileDownload `json:"file_list"`
}

// TCBGetDownloadFileLinkResponse is returned by TCBGetDownloadFileLink.
type TCBGetDownloadFileLinkResponse struct {
	ErrResponse
	// RawData of the download links payload.
	RawData map[string]any `json:"-"`
}

// TCBGetDownloadFileLink returns signed download URLs for storage files.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/storage/api_getdownloadtcbfilelink.html
func (w *MiniProgram) TCBGetDownloadFileLink(ctx context.Context, req *TCBGetDownloadFileLinkRequest) (*TCBGetDownloadFileLinkResponse, error) {
	var result TCBGetDownloadFileLinkResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/batchdownloadfile", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBDeleteCloudFile deletes storage files by id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/storage/api_deletetcbcloudfile.html
func (w *MiniProgram) TCBDeleteCloudFile(ctx context.Context, env string, fileIDs []string) error {
	body := map[string]any{"env": env, "fileid_list": fileIDs}
	return w.withAccessTokenPost(ctx, "/tcb/batchdeletefile", nil, defaultReqOptions(), body, nil)
}
