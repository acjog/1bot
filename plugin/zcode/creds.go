package zcode

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// ZCode's coding plan is served as an Anthropic-compatible Messages endpoint.
// The key is read, never written, from ZCode's own credential store.
const (
	ZaiBase      = "https://api.z.ai/api/anthropic"
	BigModelBase = "https://open.bigmodel.cn/api/anthropic"

	// DefaultModel is used when the gateway passes no --model: ZCode keeps
	// the model it picked in its window, not in a file, so there is nothing
	// on disk to read a default from.
	DefaultModel = "glm-5"
)

// cred is a coding plan's key and the Anthropic-compatible base it is served
// at (api.z.ai or BigModel).
type cred struct {
	Key  string
	Base string
}

// zcodeDir is the folder ZCode keeps its data in: ~/.zcode, or the folder the
// user moved it to in ZCode's settings (数据存储路径 / dataBaseDir), which the
// choice in the home folder's .zcode/v2/setting.json names.
func zcodeDir() string {
	home, _ := os.UserHomeDir()
	base := home
	if b, err := os.ReadFile(filepath.Join(home, ".zcode", "v2", "setting.json")); err == nil {
		var s struct {
			DataBaseDir string `json:"dataBaseDir"`
		}
		if json.Unmarshal(b, &s) == nil {
			if d := strings.TrimSpace(s.DataBaseDir); d != "" && filepath.IsAbs(d) {
				base = d
			}
		}
	}
	return filepath.Join(base, ".zcode")
}

// zcodeSecret is the key ZCode encrypts its credentials with: SHA-256 of
// $ZCODE_CREDENTIAL_SECRET, or else of a platform-and-user fallback.
func zcodeSecret() []byte {
	seed := os.Getenv("ZCODE_CREDENTIAL_SECRET")
	if seed == "" {
		home, _ := os.UserHomeDir()
		name := ""
		if u, err := user.Current(); err == nil {
			name = u.Username
			if i := strings.LastIndexByte(name, '\\'); i >= 0 {
				name = name[i+1:] // DOMAIN\user; node's userInfo() has the user alone
			}
		}
		platform := runtime.GOOS
		if platform == "windows" {
			platform = "win32"
		}
		seed = "zcode-credential-fallback:" + platform + ":" + home + ":" + name
	}
	sum := sha256.Sum256([]byte(seed))
	return sum[:]
}

// zcodeDecrypt opens a value ZCode stored as "enc:v1:" + iv.tag.ciphertext
// (base64url, AES-256-GCM under zcodeSecret).
func zcodeDecrypt(secret []byte, v string) (string, bool) {
	rest, ok := strings.CutPrefix(v, "enc:v1:")
	if !ok {
		return "", false
	}
	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return "", false
	}
	var raw [3][]byte
	for i, p := range parts {
		b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(p, "="))
		if err != nil {
			return "", false
		}
		raw[i] = b
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", false
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(raw[0]))
	if err != nil {
		return "", false
	}
	out, err := gcm.Open(nil, raw[0], append(raw[2], raw[1]...), nil)
	return string(out), err == nil
}

// readCred finds ZCode's coding plan key. The plan's key in credentials.json
// comes first (ZCode 3.14 and earlier); ZCode 3.15 keeps no key there, and
// the plan's key then sits in config.json's built-in coding plan provider.
func readCred() (cred, error) {
	v2 := filepath.Join(zcodeDir(), "v2")

	if store, ok := readJSONMap(filepath.Join(v2, "credentials.json")); ok {
		secret := zcodeSecret()
		for name, v := range store {
			// account-provider:coding-plan:account:zai-individual-coding-plan:account:<uuid>:api-key
			if !strings.Contains(name, ":coding-plan:") || !strings.HasSuffix(name, ":api-key") {
				continue
			}
			key, ok := zcodeDecrypt(secret, v)
			if !ok || !strings.Contains(key, ".") {
				continue
			}
			base := ZaiBase
			if strings.Contains(name, ":bigmodel-") {
				base = BigModelBase
			}
			return cred{Key: key, Base: base}, nil
		}
	}

	if c, ok := readConfigKey(filepath.Join(v2, "config.json")); ok {
		return c, nil
	}
	return cred{}, fmt.Errorf("no ZCode coding plan key found in %s", v2)
}

// readConfigKey reads the coding plan key from config.json, where ZCode 3.15
// keeps it as the built-in coding plan provider's options.apiKey (the Z.ai
// one first, as ZCode lists it). A key written encrypted reads too.
func readConfigKey(path string) (cred, bool) {
	var cfg struct {
		Provider map[string]struct {
			Options struct {
				APIKey string `json:"apiKey"`
			} `json:"options"`
		} `json:"provider"`
	}
	if !readJSON(path, &cfg) {
		return cred{}, false
	}
	secret := zcodeSecret()
	for _, p := range []struct {
		ids  []string
		base string
	}{
		{[]string{"builtin:zai-coding-plan", "account:zai-individual-coding-plan"}, ZaiBase},
		{[]string{"builtin:bigmodel-coding-plan", "account:bigmodel-individual-coding-plan"}, BigModelBase},
	} {
		for _, id := range p.ids {
			key := strings.TrimSpace(cfg.Provider[id].Options.APIKey)
			if strings.HasPrefix(key, "enc:") {
				key, _ = zcodeDecrypt(secret, key)
				key = strings.TrimSpace(key)
			}
			if strings.Contains(key, ".") && !strings.ContainsAny(key, " {$") {
				return cred{Key: key, Base: p.base}, true
			}
		}
	}
	return cred{}, false
}

func readJSONMap(path string) (map[string]string, bool) {
	var m map[string]string
	if !readJSON(path, &m) {
		return nil, false
	}
	return m, true
}

func readJSON(path string, dst any) bool {
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, dst) != nil {
		return false
	}
	return true
}
