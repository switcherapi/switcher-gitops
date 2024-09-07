package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const DEFAULT_JSON = "../../resources/fixtures/comparator/default.json"

func TestCheckGroupSnapshot(t *testing.T) {
	c := NewComparatorService()

	t.Run("Should return changes in group", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/changed_group.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "CHANGED",
				"diff": "GROUP",
				"path": [
					"Release 1"
				],
				"content": {
					"activated": false,
					"description": "New description"
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return new group", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/new_group.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "NEW",
				"diff": "GROUP",
				"path": [],
				"content": {
					"name": "Release 2",
					"description": "Showcase configuration 2",
					"activated": true,
					"config": [
						{
							"key": "MY_SWITCHER_4",
							"activated": false,
							"components": [
								"switcher-playground"
							]
						}
					]
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return deleted group", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/deleted_group.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "DELETED",
				"diff": "GROUP",
				"path": [
					"Release 1"
				],
				"content": null
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return new group from empty group", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile("../../resources/fixtures/comparator/default_empty.json")
		jsonRight := utils.ReadJsonFromFile(DEFAULT_JSON)
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.Equal(t, "NEW", actual.Changes[0].Action)
		assert.Equal(t, "GROUP", actual.Changes[0].Diff)
	})

	t.Run("Should return new group from empty config", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile("../../resources/fixtures/comparator/default_empty_config.json")
		jsonRight := utils.ReadJsonFromFile(DEFAULT_JSON)
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.Equal(t, "NEW", actual.Changes[0].Action)
		assert.Equal(t, "CONFIG", actual.Changes[0].Diff)
	})
}

func TestCheckConfigSnapshot(t *testing.T) {
	c := NewComparatorService()

	t.Run("Should return changes in config", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/changed_config.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "CHANGED",
				"diff": "CONFIG",
				"path": [
					"Release 1",
					"MY_SWITCHER_2"
				],
				"content": {
					"activated": true,
					"description": "New description"
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return new config", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/new_config.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "NEW",
				"diff": "CONFIG",
				"path": [
					"Release 1"
				],
				"content": {
					"key": "MY_SWITCHER_4",
					"activated": true,
					"components": [
						"benchmark"
					]
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return deleted config", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/deleted_config.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "DELETED",
				"diff": "CONFIG",
				"path": [
					"Release 1",
					"MY_SWITCHER_3"
				],
				"content": null
			}
		]}`, utils.ToJsonFromObject(actual))
	})
}

func TestCheckStrategySnapshot(t *testing.T) {
	c := NewComparatorService()

	t.Run("Should return changes in strategy", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/changed_strategy.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "CHANGED",
				"diff": "STRATEGY",
				"path": [
					"Release 1",
					"MY_SWITCHER_1",
					"VALUE_VALIDATION"
				],
				"content": {
					"activated": true
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return new strategy", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/new_strategy.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "NEW",
				"diff": "STRATEGY",
				"path": [
					"Release 1",
					"MY_SWITCHER_2"
				],
				"content": {
					"strategy": "VALUE_VALIDATION",
					"activated": true,
					"operation": "EXIST",
					"values": [
						"user_2"
					]
				}
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return deleted strategy", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/deleted_strategy.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "DELETED",
				"diff": "STRATEGY",
				"path": [
					"Release 1",
					"MY_SWITCHER_1",
					"VALUE_VALIDATION"
				],
				"content": null
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return new strategy value", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/new_strategy_value.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "NEW",
				"diff": "STRATEGY_VALUE",
				"path": [
					"Release 1",
					"MY_SWITCHER_1",
					"VALUE_VALIDATION"
				],
				"content": [
					"user_2"
				]
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return deleted strategy value", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/deleted_strategy_value.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "DELETED",
				"diff": "STRATEGY_VALUE",
				"path": [
					"Release 1",
					"MY_SWITCHER_1",
					"VALUE_VALIDATION"
				],
				"content": [
					"user_1"
				]
			}
		]}`, utils.ToJsonFromObject(actual))
	})
}

func TestCheckComponentSnapshot(t *testing.T) {
	c := NewComparatorService()

	t.Run("Should return new component", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/new_component.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "NEW",
				"diff": "COMPONENT",
				"path": [
					"Release 1",
					"MY_SWITCHER_3"
				],
				"content": [
					"new_component"
				]
			}
		]}`, utils.ToJsonFromObject(actual))
	})

	t.Run("Should return deleted component", func(t *testing.T) {
		// Given
		jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
		jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/comparator/deleted_component.json")
		snapshotLeft := c.NewSnapshotFromJson([]byte(jsonLeft))
		snapshotRight := c.NewSnapshotFromJson([]byte(jsonRight))

		// Test Check/Merge changes
		diffChanged := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
		diffNew := c.CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
		diffDeleted := c.CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
		actual := c.MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

		assert.NotNil(t, actual)
		assert.JSONEq(t, `{
		"changes": [
			{
				"action": "DELETED",
				"diff": "COMPONENT",
				"path": [
					"Release 1",
					"MY_SWITCHER_3"
				],
				"content": [
					"benchmark"
				]
			}
		]}`, utils.ToJsonFromObject(actual))
	})
}
