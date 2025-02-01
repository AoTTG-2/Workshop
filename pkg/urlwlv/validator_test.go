package urlwlv

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Validator_Validate(t *testing.T) {
	t.Run("Invalid URL", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("://invalid-url")
		require.Equal(t, ErrInvalidURL, err)
	})

	t.Run("Invalid Protocol", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("ftp://example.com/file.txt")
		require.Equal(t, ErrInvalidProtocol, err)
	})

	t.Run("Invalid Domain", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://invalidexample.com/file.txt")
		require.Equal(t, ErrInvalidDomain, err)
	})

	t.Run("Invalid Extension: missing extension", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://example.com/file")
		require.Equal(t, ErrInvalidExtension, err)
	})

	t.Run("Invalid Extension: not allowed", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://example.com/file.pdf")
		require.Equal(t, ErrInvalidExtension, err)
	})

	t.Run("Valid URL", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://example.com/file.txt")
		require.NoError(t, err)
	})

	t.Run("No Rules", func(t *testing.T) {
		validator := NewValidator([]string{}, []string{}, []string{})
		err := validator.Validate("git://any-value")
		require.NoError(t, err)
	})

	t.Run("TXT Dropbox", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"www.dropbox.com"},
			[]string{".txt"},
		)
		urlStr := "https://www.dropbox.com/scl/fi/xuxpwqvy5t10fp73mh56b/ts-challenge.txt?rlkey=uihxcpyf7860rfrwllix5ncvf&st=kbfzcy9j&dl=0"
		err := validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("TXT Discord", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"cdn.discordapp.com"},
			[]string{".txt"},
		)
		urlStr := "https://cdn.discordapp.com/attachments/1290856272274653254/1290856336313024545/ShiganshinaLARGE_MAP.txt?ex=678e5765&is=678d05e5&hm=f7b781f87351668d15c678c962b6dad447037aa04c55eee2f87ef38332621f80&"
		err := validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("TXT Github", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"raw.githubusercontent.com"},
			[]string{".txt"},
		)
		urlStr := "https://raw.githubusercontent.com/Jagerente/aottg2-awesome-cl/refs/heads/main/modes/portal/maps/portal2.txt"
		err := validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("TXT Pastebin", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"pastebin.com"},
			[]string{},
		)
		urlStr1 := "https://pastebin.com/raw/GGTCtmpC"
		urlStr2 := "https://pastebin.com/dl/GGTCtmpC"
		err := validator.Validate(urlStr1)
		require.NoError(t, err)
		err = validator.Validate(urlStr2)
		require.NoError(t, err)
	})

	t.Run("IMG Valid", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"i.imgur.com"},
			[]string{".jpg", ".png", ".jpeg"},
		)
		urlStr := "https://i.imgur.com/abc123.JPG"
		err := validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("Binary Dropbox", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"www.dropbox.com"},
			[]string{".exe", ".bin", ".dll"},
		)
		urlStr := "https://www.dropbox.com/scl/fi/abc/filename.exe?dl=0"
		err := validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("Binary Dropbox No Ext", func(t *testing.T) {
		validator := NewValidator(
			[]string{"http", "https"},
			[]string{"www.dropbox.com"},
			[]string{""},
		)
		urlStr := "https://www.dropbox.com/scl/fi/abc/filename.exe?dl=0"
		err := validator.Validate(urlStr)
		require.Equal(t, ErrInvalidExtension, err)

		urlStr = "https://www.dropbox.com/scl/fi/abc/filename?dl=0"
		err = validator.Validate(urlStr)
		require.NoError(t, err)
	})

	t.Run("Valid URL with port", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://example.com:8080/file.txt")
		require.NoError(t, err)
	})

	t.Run("Extension Normalization", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{"txt"})
		err := validator.Validate("https://example.com/file.txt")
		require.NoError(t, err)
	})

	t.Run("Domain Normalization with www", func(t *testing.T) {
		validator := NewValidator([]string{"http", "https"}, []string{"example.com"}, []string{".txt"})
		err := validator.Validate("https://www.example.com/file.txt")
		require.NoError(t, err)

		validator = NewValidator([]string{"http", "https"}, []string{"www.example.com"}, []string{".txt"})
		err = validator.Validate("https://example.com/file.txt")
		require.NoError(t, err)
	})
}
