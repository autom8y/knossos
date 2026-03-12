package hook

import (
	"os"
	"testing"
)

func TestVerify(t *testing.T) {
	secret := "test-secret"
	os.Setenv("KNOSSOS_HOOK_SECRET", secret)
	defer os.Unsetenv("KNOSSOS_HOOK_SECRET")

	payload := []byte(`{"event":"test"}`)
	signature := Sign(payload)

	tests := []struct {
		name      string
		payload   []byte
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			payload:   payload,
			signature: signature,
			want:      true,
		},
		{
			name:      "invalid signature",
			payload:   payload,
			signature: "invalid",
			want:      false,
		},
		{
			name:      "missing signature",
			payload:   payload,
			signature: "",
			want:      false,
		},
		{
			name:      "wrong payload",
			payload:   []byte(`{"event":"wrong"}`),
			signature: signature,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Verify(tt.payload, tt.signature); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerify_NoSecret(t *testing.T) {
	os.Unsetenv("KNOSSOS_HOOK_SECRET")
	payload := []byte(`{"event":"test"}`)
	
	if !Verify(payload, "any") {
		t.Error("Verify() should return true when KNOSSOS_HOOK_SECRET is unset (opt-in)")
	}
}
