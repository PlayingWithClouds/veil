package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// accountBundle wants the session "fresh": with it, it renews the session
// and answers with it; without, it says the site wants a login.
const accountBundle = `
const { writeSync } = require("fs");
async function main() {
  await Bun.stdin.text();
  const account = process.env.VEIL_ACCOUNT ? JSON.parse(process.env.VEIL_ACCOUNT) : null;
  if (!account || account.session !== "fresh") {
    writeSync(1, JSON.stringify({ error: "log in to read this", needs_login: true }) + "\n");
    process.exit(0);
  }
  writeSync(1, JSON.stringify({ session: "rotated" }) + "\n");
  writeSync(1, JSON.stringify({ result: { session: account.session } }) + "\n");
  process.exit(0);
}
main();
`

// fakeAccounts hands out one session and counts logins and renewals.
type fakeAccounts struct {
	session  string
	canLogin bool
	logins   int
	renewed  []string
}

// AccountSession returns the current session.
func (accounts *fakeAccounts) AccountSession(string) *AccountArgs {
	return &AccountArgs{Session: accounts.session}
}

// SessionRenewed records the renewal.
func (accounts *fakeAccounts) SessionRenewed(_ string, session string) {
	accounts.renewed = append(accounts.renewed, session)
}

// LoginAgain swaps in a fresh session when it can.
func (accounts *fakeAccounts) LoginAgain(context.Context, string) bool {
	accounts.logins++
	if accounts.canLogin {
		accounts.session = "fresh"
	}
	return accounts.canLogin
}

// TestRunnerLogsInAgainAndKeepsRenewedSessions runs a call the site refuses,
// expects one login and a retry that renews the session, and a refusal
// that can't be fixed to come back as ErrNeedsLogin.
func TestRunnerLogsInAgainAndKeepsRenewedSessions(t *testing.T) {
	ctx := context.Background()
	plugin := writeBundle(t, accountBundle)
	accounts := &fakeAccounts{session: "stale", canLogin: true}
	runner := NewEmbeddedRunner()
	runner.SetAccountProvider(accounts)
	raw, err := runner.invoke(ctx, plugin, "story:chapter", nil)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Session string `json:"session"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || result.Session != "fresh" {
		t.Errorf("result = %s, err %v", raw, err)
	}
	if accounts.logins != 1 || len(accounts.renewed) != 1 || accounts.renewed[0] != "rotated" {
		t.Errorf("logins %d, renewed %v", accounts.logins, accounts.renewed)
	}

	stuck := &fakeAccounts{session: "stale"}
	runner.SetAccountProvider(stuck)
	if _, err := runner.invoke(ctx, plugin, "story:chapter", nil); !errors.Is(err, ErrNeedsLogin) {
		t.Errorf("err = %v, want ErrNeedsLogin", err)
	}
	if stuck.logins != 1 {
		t.Errorf("tried to log in %d times, want once", stuck.logins)
	}
}
