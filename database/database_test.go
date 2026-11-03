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

func TestClearWatchlist(t *testing.T) {
	d := NewDatabase(t.TempDir())
	defer d.db.Close()

	n, err := d.ClearWatchlist("nobody")
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	require.NoError(t, d.NewCunt("123", &CuntInfo{Location: "Bristol", Watchlist: []string{"NVDA", "AMD"}}))
	n, err = d.ClearWatchlist("123")
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	c, err := d.GetCunt("123")
	require.NoError(t, err)
	assert.Empty(t, c.Info.Watchlist)
	assert.Equal(t, "Bristol", c.Info.Location)

	n, err = d.ClearWatchlist("123")
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
