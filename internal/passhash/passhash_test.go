package passhash

import (
	"testing"
)

func TestVerifyPassHash(t *testing.T) {
	passhash := "ag2id01$O7oUDJXQOpBLMWcOZZy2vw$EFRbA_sWRqwLf-G20ktiofDz4YnMqJvwYEdmw-mfn2s"
	ok, upgrade_required := VerifyPassHash("password", passhash)
	if !ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}

	ok, upgrade_required = VerifyPassHash("wrong_password", passhash)
	if ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}

	// invalid passhash
	passhash = "ag2id01$O7oUDJXQOpBLMWcOZZy2vwEFRbA_sWRqwL0ktiofDz4YnMqJvwYEdm*-mfn2s"
	ok, upgrade_required = VerifyPassHash("password", passhash)
	if ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}

	passhash = "ag2id01$O7oUDJXQOpBLMWcOZZ^2vw$EFRbA_sWRqwLf-G20ktiofDz4YnMqJvwYEdmw-mfn2s"
	ok, upgrade_required = VerifyPassHash("password", passhash)
	if ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}

	passhash = "ag2id01$O7oUDJXQOpBLMWcOZZy2vw$EFRbA_sWRqwLf-G20ktiofDz4Yn(qJvwYEdmw-mfn2s"
	ok, upgrade_required = VerifyPassHash("password", passhash)
	if ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}

	// unknown alg
	passhash = "unk01$O7oUDJXQOpBLMWcOZZy2vw$EFRbA_sWRqwLf-G20ktiofDz4YnMqJvwYEdmw-mfn2s"
	ok, upgrade_required = VerifyPassHash("password", passhash)
	if ok {
		t.Errorf("VerifyPassHash() failed")
	}
	if !upgrade_required {
		t.Errorf("VerifyPassHash() upgrade_required")
	}
}

func TestNewPassHash(t *testing.T) {
	passhash := NewPassHash("password")
	if len(passhash) == 0 {
		t.Errorf("NewPassHash() failed")
	}
}
