package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareVersionV1Invalid(t *testing.T) {
	result, err := CompareVersion("1.0.x", "1.0.0")
	require.True(t, err != nil)
	require.Equal(t, int64(0), result)
}

func TestCompareVersionV2Invalid(t *testing.T) {
	result, err := CompareVersion("1.0.0", "1.0.x")
	require.True(t, err != nil)
	require.Equal(t, int64(0), result)
}

func TestCompareVersionLess(t *testing.T) {
	result, err := CompareVersion("1.0.0", "1.0.1")
	require.True(t, err == nil)
	require.True(t, result < 0)

	result, err = CompareVersion("1.0", "1.0.1")
	require.True(t, err == nil)
	require.True(t, result < 0)

	result, err = CompareVersion("1.0.0.0", "1.0.1")
	require.True(t, err == nil)
	require.True(t, result < 0)

	result, err = CompareVersion("3.0.0.0", "4.0.0.0")
	require.True(t, err == nil)
	require.True(t, result < 0)
}

func TestCompareVersionEqual(t *testing.T) {
	result, err := CompareVersion("1.0.1", "1.0.1")
	require.True(t, err == nil)
	require.True(t, result == 0)

	result, err = CompareVersion("1.0.0", "1.0")
	require.True(t, err == nil)
	require.True(t, result == 0)
}

func TestCompareVersionGreater(t *testing.T) {
	result, err := CompareVersion("1.0.1", "1.0.0")
	require.True(t, err == nil)
	require.True(t, result > 0)

	result, err = CompareVersion("1.0.1", "1.0")
	require.True(t, err == nil)
	require.True(t, result > 0)

	result, err = CompareVersion("1.0.1", "1.0.0.0")
	require.True(t, err == nil)
	require.True(t, result > 0)
}

func TestWithBuild(t *testing.T) {
	result, err := CompareVersion("1.0.0-12333", "1.0.0")
	require.True(t, err == nil)
	require.True(t, result == 0)
}

func TestParseVersionCommentNormal(t *testing.T) {
	result1, err := ParseVersionComment("OceanBase 4.0.0.0 (r20220718201918-4a01f78810d731b12067f6624b73e010c36e4bf4) (Built Jul 18 2022 20:27:11)")
	require.True(t, err == nil)
	require.True(t, len(result1) > 0)

	result2, err := ParseVersionComment("OceanBase 3.2.3.1 (r20220718201918-4a01f78810d731b12067f6624b73e010c36e4bf4) (Built Jul 18 2022 20:27:11)")
	require.True(t, err == nil)
	require.True(t, len(result2) > 0)

	result3, err := ParseVersionComment("OceanBase 2.2.77 (r20220718201918-4a01f78810d731b12067f6624b73e010c36e4bf4) (Built Jul 18 2022 20:27:11)")
	require.True(t, err == nil)
	require.True(t, len(result3) > 0)
}

func TestParseVersionCommentEmpty(t *testing.T) {
	result, err := ParseVersionComment("")
	require.True(t, err != nil)
	require.True(t, len(result) == 0)
}
