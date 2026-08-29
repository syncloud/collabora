package backend

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := []byte("secret")
	token, err := SignAccessToken(secret, "abc", "user", time.Minute)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	parsed, err := VerifyAccessToken(secret, token, "abc")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if parsed.User != "user" || parsed.FileID != "abc" {
		t.Fatalf("unexpected claims: %+v", parsed)
	}
}

func TestAccessTokenRejectsOtherFile(t *testing.T) {
	secret := []byte("secret")
	token, _ := SignAccessToken(secret, "abc", "user", time.Minute)

	if _, err := VerifyAccessToken(secret, token, "other"); err == nil {
		t.Fatal("expected a file mismatch error")
	}
}

func TestAccessTokenRejectsForgedSignature(t *testing.T) {
	token, _ := SignAccessToken([]byte("secret"), "abc", "user", time.Minute)

	if _, err := VerifyAccessToken([]byte("another"), token, "abc"); err == nil {
		t.Fatal("expected a signature error")
	}
}

func TestAccessTokenRejectsExpired(t *testing.T) {
	secret := []byte("secret")
	token, _ := SignAccessToken(secret, "abc", "user", -time.Minute)

	if _, err := VerifyAccessToken(secret, token, "abc"); err == nil {
		t.Fatal("expected an expiry error")
	}
}

func TestFileIDRoundTrip(t *testing.T) {
	name, err := FileName(FileID("my report.docx"))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if name != "my report.docx" {
		t.Fatalf("got %q", name)
	}
}

func TestAppendWopiSrc(t *testing.T) {
	cases := map[string]string{
		"https://host/browser/h/cool.html?":    "https://host/browser/h/cool.html?WOPISrc=http%3A%2F%2Fx%2F1",
		"https://host/browser/h/cool.html":     "https://host/browser/h/cool.html?WOPISrc=http%3A%2F%2Fx%2F1",
		"https://host/browser/h/cool.html?a=b": "https://host/browser/h/cool.html?a=b&WOPISrc=http%3A%2F%2Fx%2F1",
	}
	for urlsrc, expected := range cases {
		if got := appendWopiSrc(urlsrc, "http://x/1"); got != expected {
			t.Fatalf("for %q got %q, want %q", urlsrc, got, expected)
		}
	}
}

func TestRebaseURL(t *testing.T) {
	got := rebaseURL("http://127.0.0.1:9980/browser/hash/cool.html?", "https://collabora.example.com")
	if got != "https://collabora.example.com/browser/hash/cool.html?" {
		t.Fatalf("got %q", got)
	}
}

func TestLockStoreAllowsPutWithoutLockHeader(t *testing.T) {
	locks := NewLockStore()
	locks.Set("doc", "LOCK123")

	if held := locks.Get("doc"); held != "LOCK123" {
		t.Fatalf("got %q", held)
	}
	locks.Clear("doc")
	if held := locks.Get("doc"); held != "" {
		t.Fatalf("expected cleared, got %q", held)
	}
}

func TestClaimsUserPrefersPreferredUsername(t *testing.T) {
	if got := (Claims{PreferredUsername: "boris", Subject: "sub-1"}).User(); got != "boris" {
		t.Fatalf("got %q", got)
	}
	if got := (Claims{Subject: "sub-1"}).User(); got != "sub-1" {
		t.Fatalf("got %q", got)
	}
}

func TestNewDocumentFileName(t *testing.T) {
	if got := (NewDocument{Name: "report", Kind: "docx"}).FileName(); got != "report.docx" {
		t.Fatalf("got %q", got)
	}
	if got := (NewDocument{Name: "report.docx", Kind: "docx"}).FileName(); got != "report.docx" {
		t.Fatalf("got %q", got)
	}
}

func TestDiscoveryDocumentActionsPrefersEdit(t *testing.T) {
	document := discoveryDocument{NetZone: discoveryNetZone{Apps: []discoveryApp{{
		Name: "writer",
		Actions: []discoveryAction{
			{Name: "view", Ext: "docx", URLSrc: "https://host/view?<ui=UI_LLCC&>"},
			{Name: "edit", Ext: "docx", URLSrc: "https://host/edit?<ui=UI_LLCC&>"},
			{Name: "edit", Ext: "", URLSrc: "https://host/ignored?"},
		},
	}}}}

	actions := document.Actions()
	if len(actions) != 1 {
		t.Fatalf("expected one action, got %v", actions)
	}
	if actions["docx"] != "https://host/edit?" {
		t.Fatalf("got %q", actions["docx"])
	}
}

func TestKindClassifiesByExtension(t *testing.T) {
	for name, expected := range map[string]string{
		"a.docx": "document", "b.xlsx": "spreadsheet",
		"c.pptx": "presentation", "d.pdf": "pdf", "e.bin": "unknown",
	} {
		if got := Kind(name); got != expected {
			t.Fatalf("%s: got %q want %q", name, got, expected)
		}
	}
}

func TestNewFileInfoDerivesIdAndKind(t *testing.T) {
	modified := time.Unix(1700000000, 0)
	info := NewFileInfo("report.xlsx", 42, modified)

	if info.ID != FileID("report.xlsx") {
		t.Fatalf("id: got %q", info.ID)
	}
	if info.Kind != "spreadsheet" {
		t.Fatalf("kind: got %q", info.Kind)
	}
	if info.Name != "report.xlsx" || info.Size != 42 || !info.ModTime.Equal(modified) {
		t.Fatalf("unexpected: %+v", info)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	secret := []byte("secret")
	signed, err := SignSession(secret, Session{Username: "boris", Email: "b@example.com", Name: "Boris"}, time.Hour)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	session, err := VerifySession(secret, signed)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if session.Username != "boris" || session.Email != "b@example.com" {
		t.Fatalf("unexpected: %+v", session)
	}
	if _, err := VerifySession([]byte("other"), signed); err == nil {
		t.Fatal("expected a signature failure")
	}
}

func TestSessionRejectsExpired(t *testing.T) {
	secret := []byte("secret")
	signed, _ := SignSession(secret, Session{Username: "boris"}, -time.Minute)
	if _, err := VerifySession(secret, signed); err == nil {
		t.Fatal("expected an expiry failure")
	}
}

func TestStateRoundTripRejectsTampering(t *testing.T) {
	secret := []byte("secret")
	signed, err := encodeState(secret, stateBlob{
		Nonce:            "n",
		Verifier:         "v",
		Return:           "/",
		RegisteredClaims: jwt.RegisteredClaims{ID: "state-1"},
	})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	blob, err := decodeState(secret, signed)
	if err != nil || blob.ID != "state-1" || blob.Verifier != "v" {
		t.Fatalf("decode failed: %v %+v", err, blob)
	}
	if _, err := decodeState([]byte("other"), signed); err == nil {
		t.Fatal("expected a signature failure")
	}
}
