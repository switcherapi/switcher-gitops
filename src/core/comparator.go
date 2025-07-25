package core

import (
	"encoding/json"
	"slices"

	"github.com/switcherapi/switcher-gitops/src/model"
)

type DiffType string
type DiffResult string

type IComparatorService interface {
	CheckSnapshotDiff(left model.Snapshot, right model.Snapshot, diffType DiffType) model.DiffResult
	MergeResults(diffResults []model.DiffResult) model.DiffResult
	NewSnapshotFromJson(jsonData []byte) model.Snapshot
	RemoveDeleted(diffResult model.DiffResult) model.DiffResult
}

type ComparatorService struct{}

const (
	NEW     DiffType = "NEW"
	CHANGED DiffType = "CHANGED"
	DELETED DiffType = "DELETED"

	GROUP     DiffResult = "GROUP"
	CONFIG    DiffResult = "CONFIG"
	STRATEGY  DiffResult = "STRATEGY"
	RELAY     DiffResult = "RELAY"
	COMPONENT DiffResult = "COMPONENT"
)

func NewComparatorService() *ComparatorService {
	return &ComparatorService{}
}

func (c *ComparatorService) NewSnapshotFromJson(jsonData []byte) model.Snapshot {
	var snapshot model.Snapshot
	json.Unmarshal(jsonData, &snapshot)
	return snapshot
}

func (c *ComparatorService) CheckSnapshotDiff(left model.Snapshot, right model.Snapshot, diffType DiffType) model.DiffResult {
	diffResult := model.DiffResult{}
	return checkGroupDiff(left, right, diffType, diffResult)
}

func (c *ComparatorService) MergeResults(diffResults []model.DiffResult) model.DiffResult {
	var result model.DiffResult

	for _, diffResult := range diffResults {
		result.Changes = append(result.Changes, diffResult.Changes...)
	}

	return result
}

func (c *ComparatorService) RemoveDeleted(diffResult model.DiffResult) model.DiffResult {
	diff := model.DiffResult{Changes: []model.DiffDetails{}}
	for _, change := range diffResult.Changes {
		if change.Action != string(DELETED) {
			diff.Changes = append(diff.Changes, change)
		}
	}

	return diff
}

func checkGroupDiff(left model.Snapshot, right model.Snapshot, diffType DiffType, diffResult model.DiffResult) model.DiffResult {
	for _, leftGroup := range left.Domain.Group {
		if !slices.Contains(model.GroupNames(right.Domain.Group), leftGroup.Name) {
			switch diffType {
			case NEW:
				appendDiffResults(string(diffType), string(GROUP), []string{}, leftGroup, &diffResult)
			case DELETED:
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
			switch diffType {
			case NEW:
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name}, leftConfig, diffResult)
			case DELETED:
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name, leftConfig.Key}, nil, diffResult)
			}
		} else {
			rightConfig := model.GetConfigByKey(rightGroup.Config, leftConfig.Key)
			modelDiffFound := model.Config{}

			diffFound := false
			if diffType == CHANGED {
				diffFound = compareAndUpdateBool(leftConfig.Activated, rightConfig.Activated, diffFound, &modelDiffFound.Activated)
				diffFound = compareAndUpdateString(leftConfig.Description, rightConfig.Description, diffFound, &modelDiffFound.Description)
				diffFound = compareAndUpdateRelay(leftConfig, rightConfig, diffFound, &modelDiffFound)
			}

			checkStrategyDiff(leftConfig, rightConfig, leftGroup, diffResult, diffType)
			checkComponentsDiff(leftConfig, rightConfig, leftGroup, diffResult, diffType)
			checkRelayDiff(leftConfig, rightConfig, leftGroup, diffResult, diffType)

			if diffFound {
				appendDiffResults(string(diffType), string(CONFIG), []string{leftGroup.Name, leftConfig.Key}, modelDiffFound, diffResult)
			}
		}
	}
}

func checkStrategyDiff(leftConfig model.Config, rightConfig model.Config, leftGroup model.Group,
	diffResult *model.DiffResult, diffType DiffType) {

	if len(leftConfig.Strategies) == 0 {
		return
	}

	for _, leftStrategy := range leftConfig.Strategies {
		if !slices.Contains(model.StrategyNames(rightConfig.Strategies), leftStrategy.Strategy) {
			switch diffType {
			case NEW:
				appendDiffResults(string(diffType), string(STRATEGY),
					[]string{leftGroup.Name, leftConfig.Key}, leftStrategy, diffResult)
			case DELETED:
				appendDiffResults(string(diffType), string(STRATEGY),
					[]string{leftGroup.Name, leftConfig.Key, leftStrategy.Strategy}, nil, diffResult)
			}
		} else {
			rightStrategy := model.GetStrategyByName(rightConfig.Strategies, leftStrategy.Strategy)
			modelDiffFound := model.Strategy{}

			diffFound := false
			if diffType == CHANGED {
				diffFound = compareAndUpdateBool(leftStrategy.Activated, rightStrategy.Activated, diffFound, &modelDiffFound.Activated)
				diffFound = compareAndUpdateString(leftStrategy.Operation, rightStrategy.Operation, diffFound, &modelDiffFound.Operation)
				diffFound = compareAndUpdateArray(leftStrategy.Values, rightStrategy.Values, diffFound, &modelDiffFound.Values)
			}

			if diffFound {
				appendDiffResults(string(diffType), string(STRATEGY),
					[]string{leftGroup.Name, leftConfig.Key, leftStrategy.Strategy}, modelDiffFound, diffResult)
			}
		}
	}
}

func checkRelayDiff(leftConfig model.Config, rightConfig model.Config, leftGroup model.Group,
	diffResult *model.DiffResult, diffType DiffType) {

	if diffType != DELETED {
		return
	}

	if leftConfig.Relay != nil && rightConfig.Relay == nil {
		appendDiffResults(string(DELETED), string(RELAY),
			[]string{leftGroup.Name, leftConfig.Key}, nil, diffResult)
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
		appendDiffResults(string(diffType), string(COMPONENT),
			[]string{leftGroup.Name, leftConfig.Key}, diff, diffResult)
	}
}

func compareAndUpdateRelay(leftConfig model.Config, rightConfig model.Config, diffFound bool, modelDiffFound *model.Config) bool {
	if leftConfig.Relay == nil || rightConfig.Relay == nil {
		return diffFound
	}

	modelDiffFound.Relay = &model.Relay{}
	diffFound = compareAndUpdateString(leftConfig.Relay.Type, rightConfig.Relay.Type, diffFound, &modelDiffFound.Relay.Type)
	diffFound = compareAndUpdateString(leftConfig.Relay.Method, rightConfig.Relay.Method, diffFound, &modelDiffFound.Relay.Method)
	diffFound = compareAndUpdateString(leftConfig.Relay.Endpoint, rightConfig.Relay.Endpoint, diffFound, &modelDiffFound.Relay.Endpoint)
	diffFound = compareAndUpdateString(leftConfig.Relay.Description, rightConfig.Relay.Description, diffFound, &modelDiffFound.Relay.Description)
	diffFound = compareAndUpdateBool(leftConfig.Relay.Activated, rightConfig.Relay.Activated, diffFound, &modelDiffFound.Relay.Activated)

	return diffFound
}

func compareAndUpdateBool(left *bool, right *bool, diffFound bool, modelDiffFound **bool) bool {
	// Bool are required and will assume right is equal to left if right is nil
	// E.g. when Respository (right) has not been set, it will assume the value from the left (API)
	if right == nil {
		right = new(bool)
		*right = *left
		diffFound = true
		*modelDiffFound = right
	} else if *left != *right {
		diffFound = true
		*modelDiffFound = right
	}

	return diffFound
}

func compareAndUpdateString(left string, right string, diffFound bool, modelDiffFound *string) bool {
	// Strings are optional and will only evaluate if right is not empty
	if right != "" && left != right {
		diffFound = true
		*modelDiffFound = right
	}
	return diffFound
}

func compareAndUpdateArray(left []string, right []string, diffFound bool, modelDiffFound *[]string) bool {
	// Arrays are optional and will only evaluate if right is not empty
	slices.Sort(left)
	slices.Sort(right)

	if slices.Compare(left, right) != 0 {
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
