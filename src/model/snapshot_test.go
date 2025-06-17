package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroups(t *testing.T) {
	groups := []Group{
		{Name: "group1"},
		{Name: "group2"},
		{Name: "group3"},
	}

	t.Run("Should return group names", func(t *testing.T) {
		result := GroupNames(groups)

		expected := []string{"group1", "group2", "group3"}
		assert.Equal(t, expected, result, "Group names do not match")
	})

	t.Run("Should return Group by name", func(t *testing.T) {
		result := GetGroupByName(groups, "group2")

		expected := Group{Name: "group2"}
		assert.Equal(t, expected, result, "Expected group not found")
	})

	t.Run("Should return empty Group if not found", func(t *testing.T) {
		result := GetGroupByName(groups, "group4")

		assert.Empty(t, result.Name, "Expected empty group when not found")
	})
}

func TestConfigs(t *testing.T) {
	configs := []Config{
		{Key: "config1"},
		{Key: "config2"},
		{Key: "config3"},
	}

	t.Run("Should return config keys", func(t *testing.T) {
		result := ConfigKeys(configs)

		expected := []string{"config1", "config2", "config3"}
		assert.Equal(t, expected, result, "Config keys do not match")
	})

	t.Run("Should return Config by key", func(t *testing.T) {
		result := GetConfigByKey(configs, "config2")

		expected := Config{Key: "config2"}
		assert.Equal(t, expected, result, "Expected config not found")
	})

	t.Run("Should return empty Config if not found", func(t *testing.T) {
		result := GetConfigByKey(configs, "config4")

		assert.Empty(t, result.Key, "Expected empty config when not found")
	})
}

func TestStrategies(t *testing.T) {
	strategies := []Strategy{
		{Strategy: "strategy1"},
		{Strategy: "strategy2"},
		{Strategy: "strategy3"},
	}

	t.Run("Should return strategy names", func(t *testing.T) {
		result := StrategyNames(strategies)

		expected := []string{"strategy1", "strategy2", "strategy3"}
		assert.Equal(t, expected, result, "Strategy names do not match")
	})

	t.Run("Should return Strategy by name", func(t *testing.T) {
		result := GetStrategyByName(strategies, "strategy2")

		expected := Strategy{Strategy: "strategy2"}
		assert.Equal(t, expected, result, "Expected strategy not found")
	})

	t.Run("Should return empty Strategy if not found", func(t *testing.T) {
		result := GetStrategyByName(strategies, "strategy4")

		assert.Empty(t, result.Strategy, "Expected empty strategy when not found")
	})
}
