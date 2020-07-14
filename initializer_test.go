// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_logger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigFileType_Indexing(t *testing.T) {
	assert.Equal(t, FileType(0), JSON)
	assert.Equal(t, FileType(1), YAML)
	assert.Equal(t, FileType(2), TOML)
	assert.Equal(t, FileType(3), HCL)
}

func TestConfigFileType_String_HappyCase(t *testing.T) {
	assert.Equal(t, "JSON", JSON.String())
	assert.Equal(t, "YAML", YAML.String())
	assert.Equal(t, "TOML", TOML.String())
	assert.Equal(t, "HCL", HCL.String())
}

func TestConfigFileType_String_Overflow_LeftBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(-1).String())
}

func TestConfigFileType_String_Overflow_RightBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(4).String())
}
