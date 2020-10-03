package nfl

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNfl(t *testing.T) {
	testResponse := `{"2019010500":{"home":{"score":{"1":0,"2":0,"3":0,"4":0,"5":0,"T":0},"abbr":"HOU","to":3},"away":{"score":{"1":0,"2":0,"3":0,"4":0,"5":0,"T":0},"abbr":"IND","to":3},"bp":0,"down":0,"togo":0,"clock":"15:00","posteam":"IND","note":null,"redzone":false,"stadium":"NRG Stadium","media":{"radio":{"home":null,"away":null},"tv":"ESPN","sat":null,"sathd":null},"yl":"","qtr":"Pregame"},"2019010501":{"home":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"DAL","to":null},"away":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"SEA","to":null},"bp":0,"down":null,"togo":null,"clock":null,"posteam":null,"note":null,"redzone":null,"stadium":"AT&T Stadium","media":{"radio":{"home":null,"away":null},"tv":"FOX","sat":null,"sathd":null},"yl":null,"qtr":null},"2019010600":{"home":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"BAL","to":null},"away":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"LAC","to":null},"bp":0,"down":null,"togo":null,"clock":null,"posteam":null,"note":null,"redzone":null,"stadium":"M&T Bank Stadium","media":{"radio":{"home":null,"away":null},"tv":"CBS","sat":null,"sathd":null},"yl":null,"qtr":null},"2019010601":{"home":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"CHI","to":null},"away":{"score":{"1":null,"2":null,"3":null,"4":null,"5":null,"T":null},"abbr":"PHI","to":null},"bp":0,"down":null,"togo":null,"clock":null,"posteam":null,"note":null,"redzone":null,"stadium":"Soldier Field","media":{"radio":{"home":null,"away":null},"tv":"NBC","sat":null,"sathd":null},"yl":null,"qtr":null}}`
	m := map[string]game{}
	err := json.Unmarshal([]byte(testResponse), &m)
	if err != nil {
		log.Print(err)
	}
	var out []string
	for _, g := range m {
		if g.Qtr == nil {
			continue
		}
		out = append(out, createGameString(&g))
	}

	assert.Equal(t, 1, len(out))
	assert.Equal(t, `0 | - | Indianapolis Colts | @ | Houston Texans | - | 0 | 15:00 Pregame | ESPN`, out[0])
}
