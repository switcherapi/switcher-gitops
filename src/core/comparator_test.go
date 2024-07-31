package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const DEFAULT_JSON = "../../resources/fixtures/default.json"

func TestCheckSnapshotDiffGroupChange(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/changed_group.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffNewGroup(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/new_group.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffNewGroupFromEmptyGroup(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile("../../resources/fixtures/default_empty.json")
	jsonRight := utils.ReadJsonFromFile(DEFAULT_JSON)
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
	assert.Equal(t, "NEW", actual.Changes[0].Action)
	assert.Equal(t, "GROUP", actual.Changes[0].Diff)
}

func TestCheckSnapshotDiffNewGroupFromEmptyConfig(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile("../../resources/fixtures/default_empty_config.json")
	jsonRight := utils.ReadJsonFromFile(DEFAULT_JSON)
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
	assert.Equal(t, "NEW", actual.Changes[0].Action)
	assert.Equal(t, "CONFIG", actual.Changes[0].Diff)
}

func TestCheckSnapshotDiffDeletedGroup(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/deleted_group.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffConfigChange(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/changed_config.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffNewConfig(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/new_config.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffDeletedConfig(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/deleted_config.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffStrategyChange(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/changed_strategy.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffNewStrategy(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/new_strategy.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffDeletedStrategy(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/deleted_strategy.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffStrategyValueChange(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/new_strategy_value.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffDeletedStrategyValue(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/deleted_strategy_value.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffNewComponent(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/new_component.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

func TestCheckSnapshotDiffDeletedComponent(t *testing.T) {
	// Given
	jsonLeft := utils.ReadJsonFromFile(DEFAULT_JSON)
	jsonRight := utils.ReadJsonFromFile("../../resources/fixtures/deleted_component.json")
	snapshotLeft := NewSnapshotFromJson([]byte(jsonLeft))
	snapshotRight := NewSnapshotFromJson([]byte(jsonRight))

	// Test Check/Merge changes
	diffChanged := CheckSnapshotDiff(snapshotLeft, snapshotRight, CHANGED)
	diffNew := CheckSnapshotDiff(snapshotRight, snapshotLeft, NEW)
	diffDeleted := CheckSnapshotDiff(snapshotLeft, snapshotRight, DELETED)
	actual := MergeResults([]model.DiffResult{diffChanged, diffNew, diffDeleted})

	AssertNotNil(t, actual)
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
		]
	}`, utils.ToJsonFromObject(actual))
}

// Helpers

func AssertNotNil(t *testing.T, object interface{}) {
	if object == nil {
		t.Errorf("Object is nil")
	}
}

func AssertContains(t *testing.T, actual string, expected string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected %v to contain %v", actual, expected)
	}
}
