package core

import (
	"encoding/json"
	"slices"

	"github.com/switcherapi/switcher-gitops/src/model"
)

type DiffType string
type DiffResult string

const (
	NEW     DiffType = "NEW"
	CHANGED DiffType = "CHANGED"
	DELETED DiffType = "DELETED"

	GROUP          DiffResult = "GROUP"
	CONFIG         DiffResult = "CONFIG"
	STRATEGY       DiffResult = "STRATEGY"
	STRATEGY_VALUE DiffResult = "STRATEGY_VALUE"
	COMPONENT      DiffResult = "COMPONENT"
)

func NewSnapshotFromJson(jsonData []byte) model.Snapshot {
	var snapshot model.Snapshot
	json.Unmarshal(jsonData, &snapshot)
	return snapshot
}

func CheckSnapshotDiff(left model.Snapshot, right model.Snapshot, diffType DiffType) model.DiffResult {
	diffResult := model.DiffResult{}
	return checkGroupDiff(left, right, diffType, diffResult)
}

func MergeResults(diffResults []model.DiffResult) model.DiffResult {
	var result model.DiffResult

	for _, diffResult := range diffResults {
		result.Changes = append(result.Changes, diffResult.Changes...)
	}

	return result
}

func checkGroupDiff(left model.Snapshot, right model.Snapshot, diffType DiffType, diffResult model.DiffResult) model.DiffResult {
	for _, leftGroup := range left.Domain.Group {
		if !slices.Contains(model.GroupNames(right.Domain.Group), leftGroup.Name) {
			if diffType == NEW {
				appendDiffResults(string(diffType), string(GROUP), []string{}, leftGroup, &diffResult)
			} else if diffType == DELETED {
				appendDiffResults(string(diffType), string(GROUP), []string{leftGroup.Name}, nil, &diffResult)
			}
		} else {
			rightGroup := model.GetGroupByName(right.Domain.Group, leftGroup.Name)
			modelDiffFound := model.Group{}

			diffFound := false
			if diffType == CHANGED {
				diffFound = compareAndUpdateBool(leftGroup.Activated, rightGroup.Activated, diffFound, &modelDiffFound.Activated)
				diffFound = compareAndUpdateString(leftGroup.Description, rightGroup.Description, diffFound, &modelDiffFound.Description)
			}

			checkConfigDiff(leftGroup, rightGroup, &diffResult, diffType)

			if diffFound {
				appendDiffResults(string(diffType), string(GROUP), []string{leftGroup.Name}, modelDiffFound, &diffResult)
			}
		}
	}

	return diffResult
}

func checkConfigDiff(leftGroup model.Group, rightGroup model.Group, diffResult *model.DiffResult, diffType DiffType) {
	if len(leftGroup.Config) == 0 {
		return
	}

	for _, leftConfig := range leftGroup.Config {
		if !slices.Contains(model.ConfigKeys(rightGroup.Config), leftConfig.Key) {
			if diffType == NEW {
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name}, leftConfig, diffResult)
			} else if diffType == DELETED {
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name, leftConfig.Key}, nil, diffResult)
			}
		} else {
			rightConfig := model.GetConfigByKey(rightGroup.Config, leftConfig.Key)
			modelDiffFound := model.Config{}

			diffFound := false
			if diffType == CHANGED {
				diffFound = compareAndUpdateBool(leftConfig.Activated, rightConfig.Activated, diffFound, &modelDiffFound.Activated)
				diffFound = compareAndUpdateString(leftConfig.Description, rightConfig.Description, diffFound, &modelDiffFound.Description)
			}

			checkStrategyDiff(leftConfig, rightConfig, leftGroup, diffResult, diffType)
			checkComponentsDiff(leftConfig, rightConfig, leftGroup, diffResult, diffType)

			if diffFound {
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name, leftConfig.Key}, modelDiffFound, diffResult)
			}
		}
	}
}

func checkStrategyDiff(leftConfig model.Config, rightConfig model.Config, leftGroup model.Group, diffResult *model.DiffResult, diffType DiffType) {
	if len(leftConfig.Strategies) == 0 {
		return
	}

	for _, leftStrategy := range leftConfig.Strategies {
		if !slices.Contains(model.StrategyNames(rightConfig.Strategies), leftStrategy.Strategy) {
			if diffType == NEW {
				appendDiffResults(string(diffType), string(STRATEGY), []string{leftGroup.Name, leftConfig.Key}, leftStrategy, diffResult)
			} else if diffType == DELETED {
				appendDiffResults(string(diffType), string(STRATEGY), []string{leftGroup.Name, leftConfig.Key, leftStrategy.Strategy}, nil, diffResult)
			}
		} else {
			rightStrategy := model.GetStrategyByName(rightConfig.Strategies, leftStrategy.Strategy)
			modelDiffFound := model.Strategy{}

			diffFound := false
			if diffType == CHANGED {
				diffFound = compareAndUpdateBool(leftStrategy.Activated, rightStrategy.Activated, diffFound, &modelDiffFound.Activated)
				diffFound = compareAndUpdateString(leftStrategy.Operation, rightStrategy.Operation, diffFound, &modelDiffFound.Operation)
			}

			checkValuesDiff(leftStrategy, rightStrategy, leftGroup, leftConfig, diffResult, diffType)

			if diffFound {
				appendDiffResults(string(diffType), string(STRATEGY),
					[]string{leftGroup.Name, leftConfig.Key, leftStrategy.Strategy}, modelDiffFound, diffResult)
			}
		}
	}
}

func checkValuesDiff(leftStrategy model.Strategy, rightStrategy model.Strategy, leftGroup model.Group, leftConfig model.Config,
	diffResult *model.DiffResult, diffType DiffType) {

	if len(leftStrategy.Values) == 0 {
		return
	}

	var diff []string
	for _, leftValue := range leftStrategy.Values {
		if (diffType == NEW || diffType == DELETED) && !slices.Contains(rightStrategy.Values, leftValue) {
			diff = append(diff, leftValue)
		}
	}

	if len(diff) > 0 {
		appendDiffResults(string(diffType), string(STRATEGY_VALUE),
			[]string{leftGroup.Name, leftConfig.Key, leftStrategy.Strategy}, diff, diffResult)
	}
}

func checkComponentsDiff(leftConfig model.Config, rightConfig model.Config, leftGroup model.Group,
	diffResult *model.DiffResult, diffType DiffType) {

	if len(leftConfig.Components) == 0 {
		return
	}

	var diff []string
	for _, leftComponent := range leftConfig.Components {
		if (diffType == NEW || diffType == DELETED) && !slices.Contains(rightConfig.Components, leftComponent) {
			diff = append(diff, leftComponent)
		}
	}

	if len(diff) > 0 {
		appendDiffResults(string(diffType), string(COMPONENT), []string{leftGroup.Name, leftConfig.Key}, diff, diffResult)
	}
}

func compareAndUpdateBool(left *bool, right *bool, diffFound bool, modelDiffFound **bool) bool {
	if *left != *right {
		diffFound = true
		*modelDiffFound = right
	}
	return diffFound
}

func compareAndUpdateString(left string, right string, diffFound bool, modelDiffFound *string) bool {
	if left != right {
		diffFound = true
		*modelDiffFound = right
	}
	return diffFound
}

func appendDiffResults(action string, diff string, path []string, content any, diffResult *model.DiffResult) {
	diffResult.Changes = append(diffResult.Changes, model.DiffDetails{
		Action:  action,
		Diff:    diff,
		Path:    path,
		Content: content,
	})
}
