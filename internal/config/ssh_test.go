package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSSHCredential(t *testing.T) {
	tests := map[string]struct {
		auth SSHAuth
		ref  string
		ok   bool
	}{
		"key with a filepath":       {auth: SSHAuthKey, ref: "filepath:~/.ssh/id_ed25519", ok: true},
		"key with a relative path":  {auth: SSHAuthKey, ref: "filepath:id_ed25519"},
		"key with literal material": {auth: SSHAuthKey, ref: "-----BEGIN OPENSSH PRIVATE KEY-----", ok: true},
		"key from a keyring":        {auth: SSHAuthKey, ref: "keyring:bastion-key", ok: true},
		"key with none":             {auth: SSHAuthKey, ref: "none:"},
		"key with nothing":          {auth: SSHAuthKey},
		"agent with nothing":        {auth: SSHAuthAgent, ok: true},
		"agent with none":           {auth: SSHAuthAgent, ref: "none:", ok: true},
		"agent with a keyring":      {auth: SSHAuthAgent, ref: "keyring:bastion"},
		"password with a literal":   {auth: SSHAuthPassword, ref: "s3cret", ok: true},
		"password with a keyring":   {auth: SSHAuthPassword, ref: "keyring:bastion", ok: true},
		"password with a filepath":  {auth: SSHAuthPassword, ref: "filepath:/etc/passwd"},
		"password with none":        {auth: SSHAuthPassword, ref: "none:", ok: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateSSHCredential(tt.auth, tt.ref)
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}
