package applier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripManagedBlock(t *testing.T) {
	t.Run("removes managed block from content", func(t *testing.T) {
		input := "ssh-rsa AAAA user@host\n" +
			"# BEGIN GLUON MANAGED KEYS\n" +
			"ssh-ed25519 BBBB managed@host\n" +
			"# END GLUON MANAGED KEYS\n"

		result := stripManagedBlock(input)

		assert.Contains(t, result, "ssh-rsa AAAA user@host")
		assert.NotContains(t, result, "GLUON MANAGED")
		assert.NotContains(t, result, "ssh-ed25519 BBBB managed@host")
	})

	t.Run("content without block is unchanged", func(t *testing.T) {
		input := "ssh-rsa AAAA user@host\nssh-rsa BBBB other@host\n"

		result := stripManagedBlock(input)

		assert.Contains(t, result, "ssh-rsa AAAA user@host")
		assert.Contains(t, result, "ssh-rsa BBBB other@host")
	})

	t.Run("only managed block produces just a trailing newline", func(t *testing.T) {
		input := "# BEGIN GLUON MANAGED KEYS\n" +
			"ssh-ed25519 BBBB managed@host\n" +
			"# END GLUON MANAGED KEYS\n"

		result := stripManagedBlock(input)

		assert.Equal(t, "\n", result)
	})

	t.Run("content before and after block is preserved", func(t *testing.T) {
		input := "ssh-rsa BEFORE before@host\n" +
			"# BEGIN GLUON MANAGED KEYS\n" +
			"ssh-ed25519 MANAGED managed@host\n" +
			"# END GLUON MANAGED KEYS\n" +
			"ssh-rsa AFTER after@host\n"

		result := stripManagedBlock(input)

		assert.Contains(t, result, "ssh-rsa BEFORE before@host")
		assert.Contains(t, result, "ssh-rsa AFTER after@host")
		assert.NotContains(t, result, "MANAGED")
		assert.NotContains(t, result, "GLUON MANAGED")
	})
}

func TestRenderManagedBlock(t *testing.T) {
	t.Run("single key wrapped with markers", func(t *testing.T) {
		keys := []string{"ssh-rsa AAAA user@host"}

		result := renderManagedBlock(keys)

		expected := "# BEGIN GLUON MANAGED KEYS\n" +
			"ssh-rsa AAAA user@host\n" +
			"# END GLUON MANAGED KEYS\n"
		assert.Equal(t, expected, result)
	})

	t.Run("multiple keys each on own line", func(t *testing.T) {
		keys := []string{
			"ssh-rsa AAAA user1@host",
			"ssh-ed25519 BBBB user2@host",
			"ssh-rsa CCCC user3@host",
		}

		result := renderManagedBlock(keys)

		expected := "# BEGIN GLUON MANAGED KEYS\n" +
			"ssh-rsa AAAA user1@host\n" +
			"ssh-ed25519 BBBB user2@host\n" +
			"ssh-rsa CCCC user3@host\n" +
			"# END GLUON MANAGED KEYS\n"
		assert.Equal(t, expected, result)
	})

	t.Run("empty slice returns empty string", func(t *testing.T) {
		result := renderManagedBlock([]string{})

		assert.Equal(t, "", result)
	})
}

func TestNormalizeKeyLines(t *testing.T) {
	t.Run("duplicate keys removed", func(t *testing.T) {
		keys := []string{
			"ssh-rsa AAAA user@host",
			"ssh-rsa AAAA user@host",
			"ssh-ed25519 BBBB other@host",
		}

		result := normalizeKeyLines(keys)

		assert.Equal(t, []string{
			"ssh-rsa AAAA user@host",
			"ssh-ed25519 BBBB other@host",
		}, result)
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		keys := []string{
			"  ssh-rsa AAAA user@host  ",
			"\tssh-ed25519 BBBB other@host\t",
		}

		result := normalizeKeyLines(keys)

		assert.Equal(t, []string{
			"ssh-rsa AAAA user@host",
			"ssh-ed25519 BBBB other@host",
		}, result)
	})

	t.Run("empty strings filtered", func(t *testing.T) {
		keys := []string{
			"ssh-rsa AAAA user@host",
			"",
			"   ",
			"ssh-ed25519 BBBB other@host",
		}

		result := normalizeKeyLines(keys)

		assert.Equal(t, []string{
			"ssh-rsa AAAA user@host",
			"ssh-ed25519 BBBB other@host",
		}, result)
	})

	t.Run("order preserved for unique keys", func(t *testing.T) {
		keys := []string{
			"ssh-rsa CCCC third@host",
			"ssh-rsa AAAA first@host",
			"ssh-ed25519 BBBB second@host",
		}

		result := normalizeKeyLines(keys)

		assert.Equal(t, []string{
			"ssh-rsa CCCC third@host",
			"ssh-rsa AAAA first@host",
			"ssh-ed25519 BBBB second@host",
		}, result)
	})
}
