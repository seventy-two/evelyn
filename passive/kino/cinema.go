package kino

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	cinemaCursorKey = "kino:cinema"
	cinemaMsgsKey   = "kino:cinema:msgs"
)

// news is one of kino's cinema announcements.
type news struct {
	ID       int64  `json:"id"`
	Kind     string `json:"kind"` // booked, starting or removed
	Reason   string `json:"reason"`
	Markdown string `json:"markdown"`
	Showing  struct {
		ID int64 `json:"id"`
	} `json:"showing"`
}

// checkCinema posts showings as they're booked and as they start, and
// deletes those posts again when a showing is ended early, cancelled or
// fails. The first check only records where kino's news is.
func (w *watcher) checkCinema() error {
	var cursor int64
	known, err := w.config.Store.GetValue(cinemaCursorKey, &cursor)
	if err != nil {
		return fmt.Errorf("loading cursor: %w", err)
	}
	var page struct {
		Latest int64  `json:"latest"`
		Items  []news `json:"items"`
	}
	if !known {
		if err := w.get("cinema/announcements", &page); err != nil {
			return err
		}
		log.Infof("kino: watching the cinema from news %d", page.Latest)
		return w.config.Store.SetValue(cinemaCursorKey, page.Latest)
	}
	if err := w.get(fmt.Sprintf("cinema/announcements?since=%d", cursor), &page); err != nil {
		return err
	}
	if len(page.Items) == 0 {
		return nil
	}
	msgs := map[int64][]string{} // showing → message IDs posted about it
	if _, err := w.config.Store.GetValue(cinemaMsgsKey, &msgs); err != nil {
		return fmt.Errorf("loading cinema posts: %w", err)
	}
	for _, n := range page.Items {
		switch n.Kind {
		case "booked", "starting":
			id, err := w.post(n.Markdown)
			if err != nil {
				return fmt.Errorf("posting: %w", err) // cursor kept: tried again next time
			}
			msgs[n.Showing.ID] = append(msgs[n.Showing.ID], id)
		case "removed":
			for _, id := range msgs[n.Showing.ID] {
				if err := w.remove(id); err != nil {
					log.Errorf("kino: deleting cinema post %s: %s", id, err)
				}
			}
			delete(msgs, n.Showing.ID)
		}
		cursor = n.ID
		if err := w.config.Store.SetValue(cinemaMsgsKey, msgs); err != nil {
			return err
		}
		if err := w.config.Store.SetValue(cinemaCursorKey, cursor); err != nil {
			return err
		}
	}
	return nil
}
