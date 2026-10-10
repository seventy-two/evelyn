package generation

import "sync"

// threads remembers which Gemini interaction produced each of the bot's
// Discord messages, so that replying to one of them continues that exact
// conversation rather than whatever the bot said most recently.
type threads struct {
	mu        sync.Mutex
	byMessage map[string]string
	order     []string // message IDs, oldest first, for eviction
	latest    string   // interaction a plain mention continues from
	max       int
}

func newThreads(max int) *threads {
	return &threads{byMessage: map[string]string{}, max: max}
}

// lookup returns the interaction that produced the given Discord message.
func (t *threads) lookup(messageID string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id, ok := t.byMessage[messageID]
	return id, ok
}

// latestID returns the interaction a plain mention should continue from.
func (t *threads) latestID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.latest
}

// record notes that the given Discord messages came from an interaction.
// setLatest makes plain mentions continue from it.
func (t *threads) record(interactionID string, messageIDs []string, setLatest bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, m := range messageIDs {
		if _, ok := t.byMessage[m]; !ok {
			t.order = append(t.order, m)
		}
		t.byMessage[m] = interactionID
	}
	for len(t.order) > t.max {
		delete(t.byMessage, t.order[0])
		t.order = t.order[1:]
	}
	if setLatest {
		t.latest = interactionID
	}
}

// reset forgets every conversation.
func (t *threads) reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.byMessage = map[string]string{}
	t.order = nil
	t.latest = ""
}
