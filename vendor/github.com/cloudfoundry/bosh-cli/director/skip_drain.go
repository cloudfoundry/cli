package director

import "strings"

type SkipDrains []SkipDrain

type SkipDrain struct {
	All  bool
	Slug InstanceGroupOrInstanceSlug
}

func (s SkipDrains) AsQueryValue() string {
	skips := []string{}

	for _, skipDrain := range s {
		if skipDrain.All {
			return "*"
		}

		skips = append(skips, skipDrain.Slug.String())
	}

	return strings.Join(skips, ",")
}

func (s *SkipDrain) UnmarshalFlag(data string) error {
	if data == "*" || data == "" {
		s.All = true
	} else {
		var err error

		s.Slug, err = NewInstanceGroupOrInstanceSlugFromString(data)
		if err != nil {
			return err
		}
	}

	return nil
}
