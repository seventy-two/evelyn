package upsetter

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Upset(s *discordgo.Session) {
	t := NewRandomTicker(12*time.Hour, 30*time.Hour)

	for {
		select {
		case <-t.C:
			for i := range s.State.Guilds {
				for {
					r := rand.Intn(s.State.Guilds[i].MemberCount)
					members, _ := s.GuildMembers(s.State.Guilds[i].ID, "0", 1000)
					p, err := s.State.Presence(s.State.Guilds[i].ID, members[r].User.ID)
					if err != nil {
						fmt.Println(err.Error())
						break
					}
					if p.Status != discordgo.StatusInvisible && p.Status != discordgo.StatusOffline {
						s.ChannelMessageSend(s.State.Guilds[i].Channels[0].ID, fmt.Sprintf("<@%s> has been chosen. Please bully them accordingly.", members[r].User.ID))
						break
					}
				}
			}
		}
	}
}

// RandomTicker is similar to time.Ticker but ticks at random intervals between
// the min and max duration values (stored internally as int64 nanosecond
// counts).
type RandomTicker struct {
	C     chan time.Time
	stopc chan chan struct{}
	min   int64
	max   int64
}

// NewRandomTicker returns a pointer to an initialized instance of the
// RandomTicker. Min and max are durations of the shortest and longest allowed
// ticks. Ticker will run in a goroutine until explicitly stopped.
func NewRandomTicker(min, max time.Duration) *RandomTicker {
	rt := &RandomTicker{
		C:     make(chan time.Time),
		stopc: make(chan chan struct{}),
		min:   min.Nanoseconds(),
		max:   max.Nanoseconds(),
	}
	go rt.loop()
	return rt
}

// Stop terminates the ticker goroutine and closes the C channel.
func (rt *RandomTicker) Stop() {
	c := make(chan struct{})
	rt.stopc <- c
	<-c
}

func (rt *RandomTicker) loop() {
	defer close(rt.C)
	t := time.NewTimer(rt.nextInterval())
	for {
		// either a stop signal or a timeout
		select {
		case c := <-rt.stopc:
			t.Stop()
			close(c)
			return
		case <-t.C:
			select {
			case rt.C <- time.Now():
				t.Stop()
				t = time.NewTimer(rt.nextInterval())
			default:
				// there could be noone receiving...
			}
		}
	}
}

func (rt *RandomTicker) nextInterval() time.Duration {
	interval := rand.Int63n(rt.max-rt.min) + rt.min
	return time.Duration(interval) * time.Nanosecond
}
