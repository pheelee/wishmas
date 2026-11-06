package app

import (
	"errors"
	"math"
	"net/url"
	"strconv"
)

type appearance struct {
	BulbSize   int `yaml:"bulb_size"`
	SnowSize   int `yaml:"snow_size"`
	SnowAmount int `yaml:"snow_amount"`
}
type bulbDecoration struct {
	Left, Drop float64
	Color      string
}
type snowDecoration struct {
	X, Size, Rest   float64
	Duration, Delay int
	Opacity         float64
}

func defaultAppearance() *appearance { return &appearance{120, 100, 16} }

// @impl CON-003-AC6 CON-003-AC7 CON-003-AC8
func (p appearance) valid() bool {
	return p.BulbSize >= 80 && p.BulbSize <= 200 && p.SnowSize >= 50 && p.SnowSize <= 200 && p.SnowAmount >= 8 && p.SnowAmount <= 48
}
func parseAppearance(form url.Values) (*appearance, error) {
	p := defaultAppearance()
	var parseErr error
	for _, f := range []struct {
		key   string
		value *int
	}{{"bulb_size", &p.BulbSize}, {"snow_size", &p.SnowSize}, {"snow_amount", &p.SnowAmount}} {
		if raw := form.Get(f.key); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				parseErr = errors.New("E15")
				continue
			}
			*f.value = value
		}
	}
	if parseErr != nil {
		return p, parseErr
	}
	if !p.valid() {
		return p, errors.New("E16")
	}
	return p, nil
}

// @impl CON-003-AC3 CON-003-AC4 CON-003-AC6 CON-003-AC7 CON-003-AC8
func decorateView(data *view) {
	if data.Appearance == nil || !data.Appearance.valid() {
		data.Appearance = defaultAppearance()
	}
	count := int(math.Round(1800 / float64(data.Appearance.BulbSize)))
	colors := []string{"coral", "mint", "gold", "blue"}
	for i := 0; i < count; i++ {
		left := 2.5 + 95*float64(i)/float64(count-1)
		x := left / 100
		data.Bulbs = append(data.Bulbs, bulbDecoration{left, 5 + 240*x*(1-x), colors[i%4]})
	}
	for i := 0; i < data.Appearance.SnowAmount; i++ {
		sideCount := (data.Appearance.SnowAmount + 1) / 2
		flake := snowDecoration{5 + 90*float64(i/2)/float64(sideCount-1), float64(1+i%3) * float64(data.Appearance.SnowSize) / 100, float64(5 + i*17%90), 20 + (i*7+2)%21, -(i * 11 % 40), .14 + float64(i%5)*.04}
		if i%2 == 0 {
			data.LeftSnow = append(data.LeftSnow, flake)
		} else {
			data.RightSnow = append(data.RightSnow, flake)
		}
	}
}
