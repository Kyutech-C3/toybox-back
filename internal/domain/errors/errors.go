package errors

import "errors"

// 共通のエラー定義
var (
	ErrInvalidRequestBody = errors.New("invalid request body")
)

// 認証関連のエラー定義
var (
	ErrUserNotAllowedGuild   = errors.New("user is not in an allowed discord guild")
	ErrRefreshTokenExpired   = errors.New("refresh token is expired")
	ErrRefreshTokenInvalid   = errors.New("refresh token is invalid")
	ErrFaileRequestToDiscord = errors.New("failed to request to discord")
	ErrClientIDNotSet        = errors.New("client ID is not set")
	ErrRedirectURLNotSet     = errors.New("redirect URL is not set")
)

// DB関連のエラー定義
var (
	ErrFailedToBeginTransaction    = errors.New("failed to begin transaction")
	ErrFailedToCommitTransaction   = errors.New("failed to commit transaction")
	ErrFailedToRollbackTransaction = errors.New("failed to rollback transaction")
)

// ユーザー関連のエラー定義
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrFailedToCreateUser = errors.New("failed to create user")
	ErrFailedToUpdateUser = errors.New("failed to update user")
)

// 作品関連のエラー定義
var (
	ErrInvalidTitle                        = errors.New("invalid title")
	ErrInvalidDescription                  = errors.New("invalid description")
	ErrInvalidVisibility                   = errors.New("invalid visibility")
	ErrInvalidUserID                       = errors.New("invalid user id")
	ErrInvalidTagIDs                       = errors.New("invalid tag ids")
	ErrOwnerCannotBeCollaborator           = errors.New("work owner cannot be added as a collaborator")
	ErrFailedToCreateWork                  = errors.New("failed to create work")
	ErrFailedToUpdateWork                  = errors.New("failed to update work")
	ErrFailedToGetAllWorksByLimitAndOffset = errors.New("failed to get all works by limit and offset")
	ErrFailedToGetWorkById                 = errors.New("failed to get work by id")
	ErrWorkNotFound                        = errors.New("work not found")
	ErrFailedToCreateThumbnail             = errors.New("failed to create thumbnail")
	ErrFailedToCreateURL                   = errors.New("failed to create url")
	ErrFailedToCreateTagging               = errors.New("failed to create tagging")
	ErrFailedToGetWorksByUserID            = errors.New("failed to get works by user id")
	ErrWorkNotOwnedByUser                  = errors.New("work not owned by user")
	ErrFailedToDeleteWork                  = errors.New("failed to delete work")
)

// コメント関連のエラー定義
var (
	ErrFailedToGetCommentsByWorkID = errors.New("failed to get comments by work id")
	ErrFailedToGetCommentById      = errors.New("failed to get comment by id")
	ErrCommentNotFound             = errors.New("comment not found")
	ErrFailedToCreateComment       = errors.New("failed to create comment")
)

// アセット関連のエラー定義
var (
	ErrFailedToOpenFile    = errors.New("failed to open file")
	ErrFailedToUploadFile  = errors.New("failed to upload file")
	ErrFailedToCreateAsset = errors.New("failed to create asset")
	ErrFailedToDeleteAsset = errors.New("failed to delete asset")
)

// いいね関連のエラー定義
var (
	ErrFailedToCreateFavorite         = errors.New("failed to create favorite")
	ErrFailedToDeleteFavorite         = errors.New("failed to delete favorite")
	ErrFailedToCountFavoritesByWorkID = errors.New("failed to count favorites by work id")
	ErrFavoriteAlreadyExists          = errors.New("favorite already exists")
	ErrFavoriteNotFound               = errors.New("favorite not found")
)

// タグ関連のエラー定義
var (
	ErrTagNotFound = errors.New("tag not found")
)
