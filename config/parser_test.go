package config

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestParseConfigFromYamlBytes(t *testing.T) {
	RegisterTestingT(t)

	type BigConfig struct {
		Simple    string
		Str       string `json:"name"`
		Integer16 int16
		Timeout   time.Duration
	}

	var testData = map[string]struct {
		input string
		want  BigConfig
		fail  bool
	}{
		"simple":          {"simple: test", BigConfig{Simple: "test"}, false},
		"bad simple":      {"simple test", BigConfig{}, true},
		"json tag":        {"name: BigName", BigConfig{Str: "BigName"}, false},
		"int16":           {"integer16: 25", BigConfig{Integer16: 25}, false},
		"duration number": {"timeout: 5000000000", BigConfig{Timeout: 5 * time.Second}, false},
		"duration string": {"timeout: 5s", BigConfig{Timeout: 5 * time.Second}, false},
		"bad duration":    {"timeout: s5s", BigConfig{}, true},
	}

	for name, tt := range testData {
		t.Run(name, func(t *testing.T) {
			RegisterTestingT(t)

			out := BigConfig{}
			err := parseConfigFromYamlBytes([]byte(tt.input), &out)

			if tt.fail {
				Expect(err).To(HaveOccurred())
				return
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(tt.want))
		})
	}
}
