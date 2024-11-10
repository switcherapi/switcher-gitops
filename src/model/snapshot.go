package model

type Domain struct {
	Group   []Group `json:"group,omitempty"`
	Version int     `json:"version,omitempty"`
}

type Group struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Activated   *bool    `json:"activated,omitempty"`
	Config      []Config `json:"config,omitempty"`
}

type Config struct {
	Key         string     `json:"key,omitempty"`
	Description string     `json:"description,omitempty"`
	Activated   *bool      `json:"activated,omitempty"`
	Strategies  []Strategy `json:"strategies,omitempty"`
	Relay       *Relay     `json:"relay,omitempty"`
	Components  []string   `json:"components,omitempty"`
}

type Strategy struct {
	Strategy  string   `json:"strategy,omitempty"`
	Activated *bool    `json:"activated,omitempty"`
	Operation string   `json:"operation,omitempty"`
	Values    []string `json:"values,omitempty"`
}

type Relay struct {
	Type        string `json:"type,omitempty"`
	Method      string `json:"method,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Activated   *bool  `json:"activated,omitempty"`
	Description string `json:"description,omitempty"`
}

type Snapshot struct {
	Domain Domain `json:"domain,omitempty"`
}

type Data struct {
	Snapshot Snapshot `json:"data,omitempty"`
}

func GroupNames(groups []Group) []string {
	names := make([]string, len(groups))
	for i, group := range groups {
		names[i] = group.Name
	}
	return names
}

func ConfigKeys(configs []Config) []string {
	keys := make([]string, len(configs))
	for i, config := range configs {
		keys[i] = config.Key
	}
	return keys
}

func StrategyNames(strategies []Strategy) []string {
	names := make([]string, len(strategies))
	for i, strategy := range strategies {
		names[i] = strategy.Strategy
	}
	return names
}

func GetStrategyByName(strategies []Strategy, name string) Strategy {
	for _, s := range strategies {
		if s.Strategy == name {
			return s
		}
	}
	return Strategy{}
}

func GetConfigByKey(configs []Config, key string) Config {
	for _, c := range configs {
		if c.Key == key {
			return c
		}
	}
	return Config{}
}

func GetGroupByName(groups []Group, name string) Group {
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	return Group{}
}
