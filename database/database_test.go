package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValues(t *testing.T) {
	d := NewDatabase(t.TempDir())
	defer d.db.Close()

	var seen []string
	found, err := d.GetValue("autotrader:seen", &seen)
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, d.SetValue("autotrader:seen", []string{"1", "2"}))
	found, err = d.GetValue("autotrader:seen", &seen)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, []string{"1", "2"}, seen)

	// Values aren't mistaken for users.
	require.NoError(t, d.NewCunt("123", &CuntInfo{Location: "Bristol"}))
	cunts, err := d.GetCunts()
	require.NoError(t, err)
	require.Len(t, cunts, 1)
	assert.Equal(t, "123", cunts[0].ID)
}
