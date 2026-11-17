package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// animationsJSON is the clip catalog from Big Blue Fish, the same file that ships in resources, so the policy and its tests read the rules the clips ship with.
//
//go:embed resources/fish/animations.json
var animationsJSON []byte

// Rule is one clip and when it plays, as Big Blue Fish's animations.json describes it.
type Rule struct {
	ID             string   `json:"Id"`
	File           string   `json:"File"`
	Pool           string   `json:"Pool"`
	Weight         float64  `json:"Weight"`
	NightOnly      bool     `json:"NightOnly"`
	SeasonOnly     string   `json:"SeasonOnly"`
	SummerWeight   *float64 `json:"SummerWeight"`
	WinterWeight   *float64 `json:"WinterWeight"`
	MorningWeight  *float64 `json:"MorningWeight"`
	Meal           string   `json:"Meal"`
	MealBoost      float64  `json:"MealBoost"`
	Festival       string   `json:"Festival"`
	RadiusDays     float64  `json:"RadiusDays"`
	FestivalMode   string   `json:"FestivalMode"`
	FestivalWeight float64  `json:"FestivalWeight"`
	WorkKind       string   `json:"WorkKind"`
	WorkWeight     float64  `json:"WorkWeight"`
	MoveKind       string   `json:"MoveKind"`
	HourParity     int      `json:"HourParity"`
}

// Context is what weights depend on besides the rules: the local time and what the user is doing.
type Context struct {
	Now     time.Time
	Coding  bool
	Working bool
}

// Policy chooses clips by pool and weight, a port of Big Blue Fish's CompanionAnimationPolicy.
type Policy struct {
	Rules []Rule
	byID  map[string]Rule
}

// LoadPolicy reads the embedded catalog.
func LoadPolicy() (*Policy, error) {
	var rules []Rule
	if err := json.Unmarshal(animationsJSON, &rules); err != nil {
		return nil, fmt.Errorf("animations.json: %w", err)
	}
	p := &Policy{Rules: rules, byID: make(map[string]Rule, len(rules))}
	for _, r := range rules {
		if _, dup := p.byID[r.ID]; dup {
			return nil, fmt.Errorf("animations.json: %q twice", r.ID)
		}
		p.byID[r.ID] = r
	}
	return p, nil
}

// Find returns the rule named id; a test checks that every name the pet plays by name is in the catalog.
func (p *Policy) Find(id string) Rule { return p.byID[id] }

// Season is the northern season of t's month.
func Season(t time.Time) string {
	switch m := t.Month(); {
	case m >= 3 && m <= 5:
		return "Spring"
	case m >= 6 && m <= 8:
		return "Summer"
	case m >= 9 && m <= 11:
		return "Autumn"
	}
	return "Winter"
}

// Meal hours, Big Blue Fish's defaults.
const (
	breakfastHour = 8
	lunchHour     = 12
	dinnerHour    = 18
)

// nearHour reports whether t is within an hour of hour o'clock, across midnight too.
func nearHour(t time.Time, hour int) bool {
	minutes := math.Abs(float64(t.Hour()*60+t.Minute()) + float64(t.Second())/60 - float64(hour*60))
	return math.Min(minutes, 1440-minutes) <= 60
}

func atMeal(r Rule, t time.Time) bool {
	switch r.Meal {
	case "Breakfast":
		return nearHour(t, breakfastHour)
	case "Lunch":
		return nearHour(t, lunchHour)
	case "Dinner":
		return nearHour(t, dinnerHour)
	case "Any":
		return nearHour(t, breakfastHour) || nearHour(t, lunchHour) || nearHour(t, dinnerHour)
	}
	return false
}

// Weight is how likely r is against the other clips of its pool at c, 0 when it may not play.
func (p *Policy) Weight(r Rule, c Context) float64 {
	t := c.Now
	season := Season(t)
	if r.Pool == "Disabled" || r.NightOnly && t.Hour() >= 8 && t.Hour() < 23 ||
		r.SeasonOnly != "" && r.SeasonOnly != season ||
		r.MoveKind != "" && (t.Minute() < 5 || t.Minute() >= 55 || t.Hour()%2 != r.HourParity) {
		return 0
	}
	w := r.Weight
	if season == "Summer" && r.SummerWeight != nil {
		w = *r.SummerWeight
	}
	if season == "Winter" && r.WinterWeight != nil {
		w = *r.WinterWeight
	}
	if t.Hour() >= 8 && t.Hour() < 11 && r.MorningWeight != nil {
		w = *r.MorningWeight
	}
	mealEligible := true
	if r.Festival != "" {
		d := festivalDistance(t, r.Festival)
		if r.FestivalMode == "Triangle" {
			w += (r.FestivalWeight - w) * math.Max(0, 1-d/r.RadiusDays)
		} else {
			mealEligible = d <= r.RadiusDays
			if mealEligible {
				w = r.FestivalWeight
			}
		}
	}
	if mealEligible && atMeal(r, t) {
		w += r.MealBoost
	}
	if r.WorkKind == "Working" && c.Working || r.WorkKind == "Coding" && c.Coding {
		w = r.WorkWeight
	}
	return math.Max(0, w)
}

// Choose picks a clip of pool by weight; ok is false when none may play.
func (p *Policy) Choose(pool string, c Context, rnd *rand.Rand) (r Rule, ok bool) {
	type entry struct {
		r Rule
		w float64
	}
	var all []entry
	sum := 0.0
	for _, r := range p.Rules {
		if r.Pool != pool {
			continue
		}
		if w := p.Weight(r, c); w > 0 {
			all = append(all, entry{r, w})
			sum += w
		}
	}
	if len(all) == 0 {
		return Rule{}, false
	}
	point := rnd.Float64() * sum
	for _, e := range all {
		if point -= e.w; point < 0 {
			return e.r, true
		}
	}
	return all[len(all)-1].r, true
}

// monthDay is a festival's day in the Gregorian calendar.
type monthDay struct {
	month time.Month
	day   int
}

// gregorianFestivals are the festivals on fixed Gregorian days.
var gregorianFestivals = map[string][]monthDay{
	"Christmas":        {{time.December, 25}},
	"Halloween":        {{time.October, 31}},
	"ValentineOrMay20": {{time.February, 14}, {time.May, 20}},
}

// festivalDistance is how many days t is from the nearest day of festival, from last year's to next year's. The lunar festivals (Spring Festival, Lantern, Dragon Boat, Qixi, Mid-Autumn, Laba) need a Chinese lunisolar calendar this port does not have yet, so they are infinitely far and their clips keep their plain weight.
func festivalDistance(t time.Time, festival string) float64 {
	days, ok := gregorianFestivals[festival]
	if !ok {
		return math.Inf(1)
	}
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	best := math.Inf(1)
	for y := t.Year() - 1; y <= t.Year()+1; y++ {
		for _, f := range days {
			day := time.Date(y, f.month, f.day, 0, 0, 0, 0, time.UTC)
			best = math.Min(best, math.Abs(today.Sub(day).Hours()/24))
		}
	}
	return best
}
