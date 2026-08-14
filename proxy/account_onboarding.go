package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"kiro-go/config"
	"kiro-go/logger"
)

const officialKiroProfileRelativePath = "Library/Application Support/Kiro/User/globalStorage/kiro.kiroagent/profile.json"

type officialKiroProfile struct {
	ARN string `json:"arn"`
}

// finishAccountAddition completes the local initialization that must happen
// before a newly persisted account is used by the model and usage refreshers.
func (h *Handler) finishAccountAddition(account *config.Account) {
	if account == nil {
		return
	}
	if synced, err := synchronizeOfficialKiroProfile(account); err != nil {
		logger.Warnf("[ProfileArn] Failed to synchronize official Kiro profile for %s: %v", accountEmailForLog(account), err)
	} else if synced {
		logger.Infof("[ProfileArn] Synchronized official Kiro profile for %s", accountEmailForLog(account))
	}

	// Reload only after profileArn is persisted, so every subsequent request
	// obtains the initialized account from the pool.
	h.pool.Reload()
	if !account.Enabled || accountBearerToken(account) == "" {
		return
	}
	go h.refreshNewAccountMetadata(account.ID)
}

func (h *Handler) refreshNewAccountMetadata(accountID string) {
	account := persistedAccount(accountID)
	if account == nil || !account.Enabled {
		return
	}
	if err := h.fetchAndCacheAccountModels(account); err != nil {
		logger.Warnf("[ModelsCache] Auto-refresh failed for new account %s: %v", accountEmailForLog(account), err)
	}
	if info, err := RefreshAccountInfo(account); err != nil {
		logger.Warnf("[AccountRefresh] Auto-refresh failed for new account %s: %v", accountEmailForLog(account), err)
	} else if err := config.UpdateAccountInfo(account.ID, *info); err != nil {
		logger.Warnf("[AccountRefresh] Failed to persist refreshed account %s: %v", accountEmailForLog(account), err)
	}
}

func persistedAccount(id string) *config.Account {
	for _, account := range config.GetAccounts() {
		if account.ID == id {
			return &account
		}
	}
	return nil
}

func synchronizeOfficialKiroProfile(account *config.Account) (bool, error) {
	profilePath, ok := officialKiroProfilePath()
	if !ok {
		return false, nil
	}
	return synchronizeOfficialKiroProfileFromPath(account, profilePath)
}

func synchronizeOfficialKiroProfileFromPath(account *config.Account, profilePath string) (bool, error) {
	if account == nil || config.IsAPIKeyAccount(account) || strings.TrimSpace(account.ProfileArn) != "" || !strings.EqualFold(strings.TrimSpace(account.AuthMethod), "idc") {
		return false, nil
	}
	profileARN, found, err := readOfficialKiroProfileARN(profilePath)
	if err != nil || !found {
		return false, err
	}
	if err := config.UpdateAccountProfileArn(account.ID, profileARN); err != nil {
		return false, fmt.Errorf("persist official Kiro profile ARN: %w", err)
	}
	account.ProfileArn = profileARN
	return true, nil
}

func officialKiroProfilePath() (string, bool) {
	if runtime.GOOS != "darwin" {
		return "", false
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", false
	}
	return filepath.Join(home, officialKiroProfileRelativePath), true
}

func readOfficialKiroProfileARN(path string) (string, bool, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read official Kiro profile: %w", err)
	}
	var profile officialKiroProfile
	if err := json.Unmarshal(contents, &profile); err != nil {
		return "", false, fmt.Errorf("decode official Kiro profile: %w", err)
	}
	profileARN, _, ok := parseKiroProfileArn(profile.ARN)
	if !ok {
		return "", false, fmt.Errorf("official Kiro profile ARN is invalid")
	}
	return profileARN, true, nil
}
