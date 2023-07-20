package secret

import "testing"

func TestArgon2PasswordSecret(t *testing.T) {
	secret := &Argon2PasswordSecret{}
	salt, err := secret.RandomSalt()
	if err != nil {
		t.Fatal(err)
	}
	secretedPwd, _ := secret.Secret([]byte("Aa12345!"), salt)
	if b, err := secret.Matches([]byte("Aa12345!"), salt, secretedPwd); err != nil {
		t.Fatal(err)
	} else if b {
		t.Log("success")
	} else {
		t.Fatal("fail")
	}
}
