package i18n_test

import (
	_ "embed"
	"github.com/amanbolat/pkg/i18n"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"testing"
)

//go:embed test_files/strings.zh_cn.toml
var messagesZH []byte

//go:embed test_files/strings.en.toml
var messagesEN []byte

var messagesMap = map[string][]byte{
	"strings.en.toml":    messagesEN,
	"strings.zh_cn.toml": messagesZH,
}

func TestLocalizer(t *testing.T) {
	chinese, err := language.Parse("zh_cn")
	require.NoError(t, err)
	english, err := language.Parse("en")
	require.NoError(t, err)

	l, err := i18n.NewLocalizer(chinese, messagesMap)
	require.NoError(t, err)

	t.Parallel()
	t.Run("localize with no params - default localizer", func(t *testing.T) {
		res := l.Localize("hello")
		require.Equal(t, "你好", res)
	})

	t.Run("localize with params - default localizer", func(t *testing.T) {
		res := l.Localize("hello_name", map[string]interface{}{
			"Name": "John",
		})
		require.Equal(t, "你好John", res)
	})

	t.Run("localize with no params - English localizer", func(t *testing.T) {
		res := l.LocalizeFor(english, "hello")
		require.Equal(t, "Hello", res)
	})

	t.Run("localize with params - English  localizer", func(t *testing.T) {
		res := l.LocalizeFor(english, "hello_name", map[string]interface{}{
			"Name": "John",
		})
		require.Equal(t, "Hello John", res)
	})
}
