package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IAuthUsecase interface {
	GetDiscordAuthURL(ctx context.Context) (string, error)
	AuthenticateUser(ctx context.Context, code string) (string, *entity.Token, error)
	RegenerateToken(ctx context.Context, refreshToken uuid.UUID) (string, *entity.Token, error)
	Logout(ctx context.Context, refreshToken uuid.UUID) error
}

type authUsecase struct {
	discordRepository repository.DiscordRepository
	userRepository    repository.UserRepository
	tokenProvider     TokenProvider
	tokenRepository   repository.TokenRepository
	assetRepository   repository.AssetRepository
}

func NewAuthUsecase(discordRepository repository.DiscordRepository, userRepository repository.UserRepository, tokenProvider TokenProvider, tokenRepository repository.TokenRepository, assetRepository repository.AssetRepository) IAuthUsecase {
	return &authUsecase{
		discordRepository: discordRepository,
		userRepository:    userRepository,
		tokenProvider:     tokenProvider,
		tokenRepository:   tokenRepository,
		assetRepository:   assetRepository,
	}
}

func (uc *authUsecase) GetDiscordAuthURL(ctx context.Context) (string, error) {
	if _, err := uc.discordRepository.GetDiscordClientID(ctx); err != nil {
		return "", err
	}
	if _, err := uc.discordRepository.GetRedirectURL(ctx); err != nil {
		return "", err
	}
	return uc.discordRepository.GetDiscordAuthURL(ctx)
}

func (uc *authUsecase) AuthenticateUser(ctx context.Context, code string) (string, *entity.Token, error) {
	token, err := uc.discordRepository.GetDiscordToken(ctx, code)
	if err != nil {
		return "", nil, err
	}

	discordUser, err := uc.discordRepository.FetchDiscordUser(ctx, token)
	if err != nil {
		return "", nil, err
	}
	guildIDs, err := uc.discordRepository.GetDiscordGuilds(ctx, token)
	if err != nil {
		return "", nil, err
	}
	allowedGuildIDs, err := uc.discordRepository.GetAllowedDiscordGuilds(ctx)
	if err != nil {
		return "", nil, err
	}
	if !userBelongsToAllowedGuild(guildIDs, allowedGuildIDs) {
		return "", nil, domainerrors.ErrUserNotAllowedGuild
	}

	user, err := uc.userRepository.GetUserByDiscordUserID(ctx, discordUser.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			avatarURL, err := uc.assetRepository.UploadAvatar(ctx, discordUser.ID, discordUser.AvatarHash)
			if err != nil {
				return "", nil, err
			}
			user = entity.NewUser(discordUser.Username, discordUser.Email, discordUser.Username, discordUser.ID, *avatarURL)
			user, err = uc.userRepository.Create(ctx, user)
			if err != nil {
				return "", nil, err
			}
		} else {
			return "", nil, err
		}
	}

	appToken, err := uc.tokenProvider.GenerateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	newRefreshToken := entity.NewToken(user.ID)
	refreshToken, err := uc.tokenRepository.Create(ctx, newRefreshToken)
	if err != nil {
		return "", nil, err
	}

	return appToken, refreshToken, nil
}

func userBelongsToAllowedGuild(guildIDs []string, allowedGuildIDs []string) bool {
	for _, guildID := range guildIDs {
		if slices.Contains(allowedGuildIDs, guildID) {
			return true
		}
	}
	return false
}

func (uc *authUsecase) RegenerateToken(ctx context.Context, refreshToken uuid.UUID) (string, *entity.Token, error) {
	userID, err := uc.tokenRepository.CheckRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", nil, err
	}

	appToken, err := uc.tokenProvider.GenerateToken(userID)
	if err != nil {
		return "", nil, err
	}
	updatedRefreshToken, err := uc.tokenRepository.UpdateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", nil, err
	}
	return appToken, updatedRefreshToken, nil
}

func (uc *authUsecase) Logout(ctx context.Context, refreshToken uuid.UUID) error {
	return uc.tokenRepository.DeleteRefreshToken(ctx, refreshToken)
}
