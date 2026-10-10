package slash

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// replySession is a copy of the bot's session whose messages to the
// interaction's channel become the interaction's reply: the first replaces the
// "thinking" placeholder and later ones are sent as follow-ups. This lets the
// existing ! handlers answer slash commands without knowing about them.
type replySession struct {
	session *discordgo.Session
	t       *replyTransport
}

func newReplySession(s *discordgo.Session, i *discordgo.Interaction) *replySession {
	base := s.Client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	t := &replyTransport{
		base:         base,
		messagesPath: "/channels/" + i.ChannelID + "/messages",
		original:     discordgo.EndpointInteractionResponseActions(i.AppID, i.Token),
		followup:     discordgo.EndpointFollowupMessage(i.AppID, i.Token),
		active:       true,
		replied:      make(chan struct{}),
	}

	rs, _ := discordgo.New(s.Token)
	rs.State = s.State
	rs.Ratelimiter = s.Ratelimiter
	rs.UserAgent = s.UserAgent
	rs.MaxRestRetries = s.MaxRestRetries
	rs.Client = &http.Client{Timeout: s.Client.Timeout, Transport: t}
	return &replySession{session: rs, t: t}
}

// finish waits up to timeout for a reply to have been sent, then stops
// redirecting messages so anything sent later goes to the channel as normal.
// It reports whether a reply was sent.
func (r *replySession) finish(timeout time.Duration) bool {
	select {
	case <-r.t.replied:
	case <-time.After(timeout):
	}
	r.t.mu.Lock()
	defer r.t.mu.Unlock()
	r.t.active = false
	return r.t.sent > 0
}

type replyTransport struct {
	base         http.RoundTripper
	messagesPath string
	original     string
	followup     string

	mu      sync.Mutex
	active  bool
	sent    int
	replied chan struct{}
}

func (t *replyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodPost || !strings.HasSuffix(req.URL.Path, t.messagesPath) {
		return t.base.RoundTrip(req)
	}

	// Held for the whole request so messages stay in order and only the
	// first successful one replaces the placeholder.
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return t.base.RoundTrip(req)
	}

	target, method := t.followup, http.MethodPost
	if t.sent == 0 {
		target, method = t.original, http.MethodPatch
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	r := req.Clone(req.Context())
	r.Method, r.URL, r.Host = method, u, u.Host

	resp, err := t.base.RoundTrip(r)
	if err == nil && resp.StatusCode < 300 {
		if t.sent == 0 {
			close(t.replied)
		}
		t.sent++
	}
	return resp, err
}
