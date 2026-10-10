// Package slash exposes the existing ! commands as Discord slash commands.
//
// Each slash command is translated back into the equivalent ! message and fed
// to the command's existing message handler, so both forms behave identically.
package slash

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// Options gives access to a slash command's options by name.
type Options map[string]*discordgo.ApplicationCommandInteractionDataOption

func (o Options) String(name string) string {
	if v, ok := o[name]; ok {
		return v.StringValue()
	}
	return ""
}

func (o Options) Int(name string) int64 {
	if v, ok := o[name]; ok {
		return v.IntValue()
	}
	return 0
}

func (o Options) Bool(name string) bool {
	if v, ok := o[name]; ok {
		return v.BoolValue()
	}
	return false
}

// Command maps a slash command onto an existing ! command handler.
type Command struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	// Trigger is the ! command this maps to; defaults to "!" + Name.
	Trigger string
	// Text builds the ! message from the options. Defaults to Trigger followed
	// by each non-empty option value in declared order.
	Text func(Options) string
	// Handler is the existing message handler for the ! command.
	Handler func(*discordgo.Session, *discordgo.MessageCreate)
	// Allowed optionally restricts which channels the command may be used in.
	Allowed func(channelID string) bool
}

// StringOpt is a shorthand for a string option.
func StringOpt(name, description string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

func (c *Command) trigger() string {
	if c.Trigger != "" {
		return c.Trigger
	}
	return "!" + c.Name
}

func (c *Command) text(o Options) string {
	if c.Text != nil {
		return c.Text(o)
	}
	parts := []string{c.trigger()}
	for _, opt := range c.Options {
		if v, ok := o[opt.Name]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v.Value)); s != "" {
				parts = append(parts, s)
			}
		}
	}
	return strings.Join(parts, " ")
}

// Router holds the slash commands and dispatches interactions to them.
type Router struct {
	commands map[string]*Command
	logger   func(*discordgo.Session, *discordgo.MessageCreate)
}

// NewRouter creates a Router. logger, if set, is called with each translated
// message so slash commands appear in the message log.
func NewRouter(logger func(*discordgo.Session, *discordgo.MessageCreate)) *Router {
	return &Router{commands: map[string]*Command{}, logger: logger}
}

// Add adds commands to the router. All commands must be added before Register.
func (r *Router) Add(cmds ...*Command) {
	for _, c := range cmds {
		r.commands[c.Name] = c
	}
}

// Register publishes the commands to Discord, replacing any previously
// registered set, and starts handling interactions.
func (r *Router) Register(s *discordgo.Session) error {
	app, err := s.Application("@me")
	if err != nil {
		return fmt.Errorf("fetching application: %w", err)
	}

	defs := make([]*discordgo.ApplicationCommand, 0, len(r.commands))
	for _, c := range r.commands {
		defs = append(defs, &discordgo.ApplicationCommand{
			Name:        c.Name,
			Description: c.Description,
			Options:     c.Options,
		})
	}
	if _, err := s.ApplicationCommandBulkOverwrite(app.ID, "", defs); err != nil {
		return fmt.Errorf("registering slash commands: %w", err)
	}

	s.AddHandler(r.handle)
	return nil
}

func (r *Router) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()
	c, ok := r.commands[data.Name]
	if !ok {
		return
	}

	if c.Allowed != nil && !c.Allowed(i.ChannelID) {
		r.respond(s, i, "That command isn't available in this channel.", true)
		return
	}

	opts := Options{}
	for _, o := range data.Options {
		opts[o.Name] = o
	}
	text := c.text(opts)

	user := i.User
	if i.Member != nil {
		user = i.Member.User
	}

	// Discord requires a reply within 3 seconds, so show "thinking" while the
	// command runs; its first message then replaces that.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Errorf("deferring /%s: %s", c.Name, err)
		return
	}

	m := &discordgo.MessageCreate{Message: &discordgo.Message{
		ID:        i.ID,
		ChannelID: i.ChannelID,
		GuildID:   i.GuildID,
		Author:    user,
		Content:   text,
		Timestamp: time.Now(),
	}}
	if r.logger != nil {
		r.logger(s, m)
	}

	reply := newReplySession(s, i.Interaction)
	c.Handler(reply.session, m)

	// Some commands send their first message from a goroutine, so give them
	// a moment before giving up on output.
	if !reply.finish(10 * time.Second) {
		content := "Done."
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}); err != nil {
			log.Errorf("finishing /%s: %s", c.Name, err)
		}
	}
}

func (r *Router) respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	data := &discordgo.InteractionResponseData{
		Content:         content,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}
	if ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
	if err != nil {
		log.Errorf("responding to /%s: %s", i.ApplicationCommandData().Name, err)
	}
}
