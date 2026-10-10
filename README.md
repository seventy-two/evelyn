# evelyn

A Discord assistant bot. Every command works with the `!` prefix or as a
slash command (`!w london` or `/w location:london`). `!help` lists them in
Discord.

## Commands

### Weather

| Command | |
| --- | --- |
| `!w [location]` | Current weather. Uses your saved location if you leave it out. |
| `!f [location]` | Forecast. Uses your saved location if you leave it out. |
| `!set location <location>` | Save your default location. |

Locations are looked up with Google Geocoding and weather comes from
[Pirate Weather](https://pirateweather.net/).

### Words

| Command | |
| --- | --- |
| `!dict <word>` | Definitions (up to 3), from [Wordnik](https://developer.wordnik.com/). |
| `!syn <word>` | Synonyms, from Wordnik. |
| `!wotd` | Wordnik's word of the day. |
| `!ud <term>` | Top definition from [Urban Dictionary](https://www.urbandictionary.com/). |

### Movies and TV

| Command | |
| --- | --- |
| `!m <title> [(year)]` | Plot, ratings and poster from [OMDb](https://www.omdbapi.com/), e.g. `!m heat (1995)`. |
| `!tv <show>` | Show info, airtime and next episode from [TVmaze](https://www.tvmaze.com/api). |

### Stocks

Stock commands only work in the stocks channels.

| Command | |
| --- | --- |
| `!q <stock>` | Quick quote, by symbol or company name. |
| `!quote <stock>` | Detailed quote. |
| `!wl` / `!watchlist` | Your watchlist. `!wl add <stock>`, `!wl del <symbol>`, `!wl help`. |
| `!earnings [text]` | Show the earnings note, or replace it with `text`. |
| `!ex <from> <to>` | Currency exchange rate, e.g. `!ex GBP USD`. |
| `!nasdaq`, `!sp` / `!s&p`, `!dow` | Top and bottom movers for the NASDAQ 100, S&P 500 or Dow 30. |

Quotes and movers come from CNBC, symbol search and exchange rates from
[Alpha Vantage](https://www.alphavantage.co/), and logos from
[logo.dev](https://www.logo.dev/).

### Other

| Command | |
| --- | --- |
| `!wa <question>` | Answer from [Wolfram Alpha](https://products.wolframalpha.com/api/). |
| `!nfl` | This week's NFL scores and upcoming games, from ESPN. |
| `!beer <name>` | Beer details from [Untappd](https://untappd.com/api/docs). |
| `!qotd` | Quote of the day from [ZenQuotes](https://zenquotes.io/). |
| `!help` | List commands. |

## Passive features

These run without a command:

- **AI chat** (main server only): mention the bot to chat with Google Gemini.
  It can see images in or replied to by your message, and replying to one of
  its messages continues that conversation. `!reset` (bot owner only) clears
  the chat history.
- **YouTube summaries** (main server only): posts a summary of linked YouTube
  videos up to 20 minutes long, or any length if you mention the bot.
- **Voice note transcription** (main server only): transcribes `.ogg` voice
  notes with Gemini's speech-to-text model.
- **Vertical media "fix"**: reposts tall images stretched to landscape
  (occasionally deep-fried) and tall videos rotated 90°.

## Running

```sh
go build -o evelyn .
./evelyn --help   # lists every option and its environment variable
```

All configuration is through environment variables. The bot needs:

| Variable | For |
| --- | --- |
| `EVELYN_DISCORD_API_KEY` | Discord bot token |
| `GEOCODE_API_KEY` | Google Geocoding (weather) |
| `WEATHER_API_KEY` | Pirate Weather |
| `WORDNIK_API_KEY` | Wordnik (dictionary) |
| `MOVIE_API_KEY` | OMDb |
| `MATH_API_KEY` | Wolfram Alpha app ID |
| `STOCKS_QUOTE_KEY` | Alpha Vantage |
| `STOCKS_LOGO_API_KEY` | logo.dev |
| `BEER_API_KEY` | Untappd credentials query string (`client_id=…&client_secret=…&access_token=…`) |
| `GENERATION_API_KEY` | Google Gemini |
| `GENERATION_YOUTUBE_API_KEY` | YouTube Data API (video length) |

TVmaze, Urban Dictionary, ESPN, CNBC and ZenQuotes need no key. Data is stored
in a Badger database at `DB_PATH` (default `/root/evelynDB`). Messages are
logged to `LOG_PATH` and errors to `ERR_PATH`.

Slash commands are registered with Discord each time the bot starts.
