package secret

import (
	"context"
	"reflect"
	"testing"

	"github.com/zalando/go-keyring"
)

type stubSecret string

func (s stubSecret) Secret() string { return string(s) }

func TestResolveLiteral(t *testing.T) {
	cases := map[string]string{
		"hunter2":         "hunter2",
		"literal:hunter2": "hunter2",
		"literal:a:b":     "a:b",
		"pg://weird:pass": "pg://weird:pass",
		"":                "",
	}

	r := Default()
	for raw, want := range cases {
		got, err := r.Resolve(context.Background(), stubSecret(raw))
		if err != nil {
			t.Fatalf("Resolve(%q): %v", raw, err)
		}
		if got != want {
			t.Errorf("Resolve(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestStoreAndResolveKeyring(t *testing.T) {
	keyring.MockInit()

	r := Default()
	s := stubSecret("keyring:prod-db")
	if err := r.Store(context.Background(), s, "s3cret"); err != nil {
		t.Fatal(err)
	}

	got, err := r.Resolve(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if got != "s3cret" {
		t.Errorf("got %q, want %q", got, "s3cret")
	}

	if stored, err := keyring.Get(keyringService, "prod-db"); err != nil || stored != "s3cret" {
		t.Errorf("keyring account must not carry the scheme prefix: got %q, err %v", stored, err)
	}
}

func TestResolveKeyringExplicitService(t *testing.T) {
	keyring.MockInit()

	if err := keyring.Set("other-app", "prod-db", "foreign"); err != nil {
		t.Fatal(err)
	}

	got, err := Default().Resolve(context.Background(), stubSecret("keyring:other-app/prod-db"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "foreign" {
		t.Errorf("got %q, want %q", got, "foreign")
	}
}

func TestResolveKeyringMissing(t *testing.T) {
	keyring.MockInit()

	if _, err := Default().Resolve(context.Background(), stubSecret("keyring:absent")); err == nil {
		t.Fatal("expected an error for a secret that is not stored")
	}
}

func TestResolveKeyringEmptyAccount(t *testing.T) {
	keyring.MockInit()

	if _, err := Default().Resolve(context.Background(), stubSecret("keyring:")); err == nil {
		t.Fatal("expected an error when the account is empty")
	}
}

func TestParseKeyringSpec(t *testing.T) {
	cases := []struct {
		spec    string
		service string
		account string
	}{
		{"prod-db", keyringService, "prod-db"},
		{"other-app/prod-db", "other-app", "prod-db"},
		{"", keyringService, ""},
	}

	for _, c := range cases {
		service, account := parseKeyringSpec(c.spec)
		if service != c.service || account != c.account {
			t.Errorf("parseKeyringSpec(%q) = (%q, %q), want (%q, %q)", c.spec, service, account, c.service, c.account)
		}
	}
}

func TestUsagesCountReferences(t *testing.T) {
	keyring.MockInit()

	r := Default()
	r.Track(stubSecret("keyring:prod-cluster"))
	r.Track(stubSecret("keyring:prod-cluster"))
	r.Track(stubSecret("keyring:staging"))
	r.Track(stubSecret("hunter2"))

	want := []Usage{{Spec: "prod-cluster", Refs: 2}, {Spec: "staging", Refs: 1}}
	if got := r.Usages(Keyring); !reflect.DeepEqual(got, want) {
		t.Errorf("Usages(Keyring) = %v, want %v", got, want)
	}

	if got := r.Usages(Literal); len(got) != 0 {
		t.Errorf("Usages(Literal) = %v, want none", got)
	}
}

func TestUntrackDropsSpec(t *testing.T) {
	keyring.MockInit()

	r := Default()
	r.Track(stubSecret("keyring:only"))
	r.Untrack(stubSecret("keyring:only"))

	if got := r.Usages(Keyring); len(got) != 0 {
		t.Errorf("Usages(Keyring) = %v, want none", got)
	}
}

func TestRemoveKeepsSharedSecret(t *testing.T) {
	keyring.MockInit()

	r := Default()
	s := stubSecret("keyring:shared")
	if err := r.Store(context.Background(), s, "pw"); err != nil {
		t.Fatal(err)
	}
	r.Track(s)

	if err := r.Remove(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(context.Background(), s); err != nil {
		t.Fatalf("secret still referenced elsewhere must survive removal: %v", err)
	}

	if err := r.Remove(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(context.Background(), s); err == nil {
		t.Fatal("last reference removed, secret must be gone")
	}
}

func TestStoreLiteralIsNoop(t *testing.T) {
	keyring.MockInit()

	r := Default()
	s := stubSecret("literal:hunter2")
	if err := r.Store(context.Background(), s, "ignored"); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(context.Background(), s); err != nil {
		t.Fatal(err)
	}

	got, err := r.Resolve(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q", got, "hunter2")
	}
}
