package proxy

import (
	"os"
	"path/filepath"
	"testing"

	"kiro-go/config"
)

func TestReadOfficialKiroProfileARNValidatesAndNormalizesARN(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "profile.json")
	const profileARN = "arn:aws:codewhisperer:us-east-1:123456789012:profile/example_1"
	if err := os.WriteFile(profilePath, []byte(`{"arn":" `+profileARN+` "}`), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	got, found, err := readOfficialKiroProfileARN(profilePath)
	if err != nil || !found {
		t.Fatalf("readOfficialKiroProfileARN() found=%v err=%v, want valid profile", found, err)
	}
	if got != profileARN {
		t.Fatalf("profile ARN = %q, want %q", got, profileARN)
	}
}

func TestReadOfficialKiroProfileARNRejectsInvalidARN(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(profilePath, []byte(`{"arn":"not-an-arn"}`), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	_, found, err := readOfficialKiroProfileARN(profilePath)
	if err == nil || found {
		t.Fatalf("readOfficialKiroProfileARN() found=%v err=%v, want validation error", found, err)
	}
}

func TestSynchronizeOfficialKiroProfilePersistsProfileForIDCAccount(t *testing.T) {
	if err := config.Init(filepath.Join(t.TempDir(), "config.json")); err != nil {
		t.Fatalf("config.Init: %v", err)
	}
	account := config.Account{ID: "account-1", AuthMethod: "idc", Enabled: true}
	if err := config.AddAccount(account); err != nil {
		t.Fatalf("config.AddAccount: %v", err)
	}

	profilePath := filepath.Join(t.TempDir(), "profile.json")
	const profileARN = "arn:aws:codewhisperer:us-east-1:123456789012:profile/example_1"
	if err := os.WriteFile(profilePath, []byte(`{"arn":"`+profileARN+`"}`), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	synced, err := synchronizeOfficialKiroProfileFromPath(&account, profilePath)
	if err != nil || !synced {
		t.Fatalf("synchronizeOfficialKiroProfileFromPath() synced=%v err=%v", synced, err)
	}
	stored := persistedAccount(account.ID)
	if stored == nil || stored.ProfileArn != profileARN {
		t.Fatalf("stored profile = %#v, want %q", stored, profileARN)
	}
}
