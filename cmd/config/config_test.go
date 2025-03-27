package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// newPCCViper function for creating a mocked up Viper with YAML configuration
func newPCCViper(t *testing.T, configData string) *viper.Viper {
	t.Helper()

	v := viper.New()
	v.SetConfigType("yaml")

	err := v.ReadConfig(strings.NewReader(configData))
	if err != nil {
		t.Fatalf("Error load mock config: %v", err)
	}

	return v
}

func TestMain(m *testing.M) {
	initValidator()

	m.Run()
}

// TestPCCRulesUnmarshal test of loading and deserialization of PCC-rules without file
func TestPCCRulesUnmarshal(t *testing.T) {
	mockConfig := `
pcc_rules:
  - pcc_name: "block_icmp"
    sdf_filter: "permit out ip from 192.168.1.100/32 to any"
    far:
      action: 2
      outer_header_creation: 0
      teid: 0
      remote_ip: 0
      transport_level_marking: 0
    qer:
      qfi: 9
      max_bitrate_ul: 100000
      max_bitrate_dl: 100000
`
	v := newPCCViper(t, mockConfig)

	var config PCCRulesConfig
	err := v.UnmarshalExact(&config)
	assert.NoError(t, err)

	assert.Len(t, config.PccRules, 1)
	assert.Equal(t, "block_icmp", config.PccRules[0].PccName)
	assert.Equal(t, "permit out ip from 192.168.1.100/32 to any", config.PccRules[0].SdfFilter)
	assert.Equal(t, uint8(2), config.PccRules[0].Far.Action)
	assert.Equal(t, uint8(9), config.PccRules[0].Qer.Qfi)
}

// TestPCCRulesValidation validation test of a valid PCC configuration
func TestPCCRulesValidation(t *testing.T) {
	mockConfig := `
pcc_rules:
  - pcc_name: "allow_http"
    sdf_filter: "permit out ip from 192.168.1.100/32 to any"
    far:
      action: 1
    qer:
      qfi: 9
      max_bitrate_ul: 100000
      max_bitrate_dl: 100000
`
	v := newPCCViper(t, mockConfig)

	var config PCCRulesConfig
	err := v.UnmarshalExact(&config)
	assert.NoError(t, err)

	err = validate.Struct(config)
	assert.NoError(t, err)
}

// TestPCCRulesIncorrectSPDFFilter check regexp validation
func TestPCCRulesIncorrectSPDFFilter(t *testing.T) {
	invalidConfigs := []string{
		`pcc_rules:
  - pcc_name: "allow_http"
    sdf_filter: "pert out ip from 192.168.1.100/32 to any"
    far:
      action: 1
    qer:
      qfi: 9
      max_bitrate_ul: 100000
      max_bitrate_dl: 100000
`,
		`pcc_rules:
  - pcc_name: "allow_http"
    sdf_filter: "permit out ip from 192.168.1.100/32 to anyy"
    far:
      action: 1
    qer:
      qfi: 9
      max_bitrate_ul: 100000
      max_bitrate_dl: 100000
`,
	}

	for _, cfg := range invalidConfigs {
		v := newPCCViper(t, cfg)

		var config PCCRulesConfig
		err := v.UnmarshalExact(&config)
		assert.NoError(t, err)

		err = validate.Struct(config)
		assert.Error(t, err)
	}

}

// TestPCCRulesMissingFields checks that there are no mandatory fields
func TestPCCRulesMissingFields(t *testing.T) {
	invalidConfig := `
pcc_rules:
  - pcc_name: ""
    sdf_filter: ""
    far:
      action: 0
    qer:
      qfi: 0
`
	v := newPCCViper(t, invalidConfig)

	var config PCCRulesConfig
	err := v.UnmarshalExact(&config)
	assert.NoError(t, err)

	err = validate.Struct(config)
	assert.Error(t, err)
}
