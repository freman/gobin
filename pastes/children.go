package pastes

import (
	"encoding/json"
)

type Children map[string]bool

func (c *Children) UnmarshalJSON(b []byte) error {
	list := make([]string,0)
	err := json.Unmarshal(b, &list)
	if err != nil {
		return err
	}

	if len(list) > 0 {
		*c = Children{}
		for _, id := range list {
			(*c)[id] = true
		}
	}

	return nil
}

func (c *Children) MarshalJSON() ([]byte, error) {
	list := make([]string, len(*c))
	i := 0;
	for id := range *c {
		list[i] = id
		i += 1
	}
	return json.Marshal(list)
}
