package secret

import (
	"testing"
)

func TestCommandConstants(t *testing.T) {
	if commandGetSecrets != "getsecret" {
		t.Errorf("commandGetSecrets = %s, want 'getsecret'", commandGetSecrets)
	}
	if commandGetSecretsAndDecrypt != "getsecret-decrypt" {
		t.Errorf("commandGetSecretsAndDecrypt = %s, want 'getsecret-decrypt'", commandGetSecretsAndDecrypt)
	}
}

func TestFlagNames(t *testing.T) {
	// Verify flag variable bindings are declared
	// These would panic at init if the flag registration had issues
	_ = kmsEndpoint
	_ = authMethod
	_ = aapArn
	_ = tokenPath
	_ = clientKeyConfigPath
	_ = clientKeyPasswordPath
}
