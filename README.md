# hayley
A simple discord assistant

Prefixes are currently hardcoded with trigger `!`


#### APIs that require keys

* [Google Geocode](https://developers.google.com/maps/documentation/geocoding/intro): [commands/weather](commands/weather/weather.go)
* [Forecast.io](https://developer.forecast.io/docs/v2): [commands/weather](commands/weather/weather.go)
* [Twitter](https://dev.twitter.com/rest/public): [commands/twitter](commands/twitter/twitter.go)
* [Wolfram Alpha](http://products.wolframalpha.com/api/): [commands/wolfram](commands/wolfram/wolfram.go)
* [Youtube](https://developers.google.com/youtube/v3/): [commands/youtube](commands/youtube/youtube.go)
* [Dota2](https://steamcommunity.com/dev/apikey): [commands/dota](commands/dota/dota.go)
* [Omdb](http://omdbapi.com/): [commands/omdb](commands/omdb/omdb.go)
* [Football/Soccer](http://api.football-data.org/): [commands/divegrass](commands/divegrass/divegrass.go)
* [Wordnik](http://api.wordnik.com): [commands/dictionary](commands/dictionary/dictionary.go)

***

### Functions

* [Dictionary](#dictionary)
* [DotA2](#DotA2)
* [Football/Soccer](#divegrass)
* [Omdb](#omdb)
* [Stocks](#stocks)
* [TVMaze](#tvmaze)
* [Urban Dictionary](#urban-dictionary)
* [Weather](#weather)
* [WolframAlpha](#wolframalpha)
* [Youtube](#youtube)

***


### Dictionary 
Returns the word of the day from Wordnik

**!word/!wotd**

Returns the Wordnik dictionary results (up to 3) for the given query

**!dict** *search query*


## Divegrass
Returns the upcoming games for the given number of days

**!f** *n1-9*

Returns the scores of the games from the past number of days

**!f** *p1-9*

Simulates behaviour of n1

**!f**

## DotA2 
Returns information on the current games being played. For tier 3 (Premium) games, games with more than 200 viewers are returned. For tier 2 (Professional) games, games with more than 1000 viewers are returned.

**!d2/dota**

Returns heroes picked, along with the above

**!d2h** 

Returns scores along with the game data

**!d2s**

Returns all information 

**!d2hs**


### Omdb
Returns tags, imdb + rt ratings, and short descriptions of the given query

**!m** *search query* 

## Stocks
Returns the current ask price, and the current change in % and USD from the NYSE of the given query. Query format must be a NYSE Symbol.

**!stocks** *Query*

### TVMaze
Info for *tv show* with episode airtime if available **-tv** *tv show*

	-tv Better call saul
	TVmaze | Better Call Saul | Airtime: Monday 22:00 on AMC | Status: Running | Next Ep: S2E6 at 22:00 2016-03-21
	
	-tv Mr Robot
	TVmaze | Mr. Robot | Airtime: Wednesday 22:00 on USA Network | Status: Running



### Urban Dictionary
Gets the first definition of *query* at [UrbanDictionary](http://www.urbandictionary.com/)

**!ud** *query*

	.urban 4chan
	Urban Dictionary | 4chan | http://mnn.im/upucr | you have just entered the very heart, soul, and life force of the internet. this is a place beyond sanity, wild and untamed. there is nothing new here. "new" content on 4chan is not found; it is created from old material. every interesting, offensive, shoc…



### Weather
[Yahoo Weather](http://weather.yahoo.com/) for *location*
**!w** *location*

	.weather Washington, DC
	Weather | Washington | Cloudy 15°C. Wind chill: 15°C. Humidity: 72%

[Yahoo Weather Forecast](http://weather.yahoo.com/) for *location*
**!f** *location*

	.forecast Washington, DC
	Forecast | Washington | Sun: Clouds Early/Clearing Late 16°C/10°C | Mon: Mostly Sunny 19°C/8°C | Tue: Mostly Sunny 23°C/11°C | Wed: Partly Cloudy 24°C/11°C
	

### WolframAlpha
Finds the answer of *question* using [WolfarmAlpha](http://www.wolframalpha.com/)

**!wa** *question*

	.wa time in Bosnia
	Wolfram | current time in Bosnia and Herzegovina >>> 12:55:38 pm CEST | Tuesday, October 6, 2015


### Youtube
Gets the first result from [Youtube](https://www.youtube.com) for *search query* 

**.yt** *search query*

	.yt Richard Stallman interject
	YouTube | I'd just like to interject... | 3m1s | https://youtu.be/QlD9UBTcSW4

***
