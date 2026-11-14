package config

import "testing"

func TestRuntimeIsClaudeOrAgy(t *testing.T) {
	for data, want := range map[string]string{`{}`: "", `{"runtime":"agy"}`: "agy", `{"runtime":"claude"}`: "claude"} {
		s, err := ParseScope([]byte(data), "/home/x")
		if err != nil || s.Runtime != want {
			t.Errorf("%s: runtime %q, %v; want %q", data, s.Runtime, err, want)
		}
	}
	for _, data := range []string{`{"runtime":"pi"}`, `{"runtime":""}`, `{"runtime":null}`} {
		if _, err := ParseScope([]byte(data), "/home/x"); err == nil {
			t.Errorf("%s is accepted", data)
		}
	}
}
