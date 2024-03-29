// Code generated by "enumer -type Entity -output entity_str_gen.go"; DO NOT EDIT.

package action

import (
	"fmt"
	"strings"
)

const _EntityName = "UnknownPullRequestIssue"

var _EntityIndex = [...]uint8{0, 7, 18, 23}

const _EntityLowerName = "unknownpullrequestissue"

func (i Entity) String() string {
	if i >= Entity(len(_EntityIndex)-1) {
		return fmt.Sprintf("Entity(%d)", i)
	}
	return _EntityName[_EntityIndex[i]:_EntityIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _EntityNoOp() {
	var x [1]struct{}
	_ = x[Unknown-(0)]
	_ = x[PullRequest-(1)]
	_ = x[Issue-(2)]
}

var _EntityValues = []Entity{Unknown, PullRequest, Issue}

var _EntityNameToValueMap = map[string]Entity{
	_EntityName[0:7]:        Unknown,
	_EntityLowerName[0:7]:   Unknown,
	_EntityName[7:18]:       PullRequest,
	_EntityLowerName[7:18]:  PullRequest,
	_EntityName[18:23]:      Issue,
	_EntityLowerName[18:23]: Issue,
}

var _EntityNames = []string{
	_EntityName[0:7],
	_EntityName[7:18],
	_EntityName[18:23],
}

// EntityString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func EntityString(s string) (Entity, error) {
	if val, ok := _EntityNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _EntityNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Entity values", s)
}

// EntityValues returns all values of the enum
func EntityValues() []Entity {
	return _EntityValues
}

// EntityStrings returns a slice of all String values of the enum
func EntityStrings() []string {
	strs := make([]string, len(_EntityNames))
	copy(strs, _EntityNames)
	return strs
}

// IsAEntity returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Entity) IsAEntity() bool {
	for _, v := range _EntityValues {
		if i == v {
			return true
		}
	}
	return false
}
