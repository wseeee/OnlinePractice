package service

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"encoding/json"
	"errors"
)

// ParseTestCases 解析测试用例 JSON 数组
func ParseTestCases(testCases []string, problemIdentity string) ([]*model.TestCase, error) {
	testCaseBasics := make([]*model.TestCase, 0, len(testCases))
	for _, testCase := range testCases {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		if err != nil {
			return nil, errors.New("测试用例格式错误")
		}
		if _, ok := caseMap["input"]; !ok {
			return nil, errors.New("测试用例缺少 input 字段")
		}
		if _, ok := caseMap["output"]; !ok {
			return nil, errors.New("测试用例缺少 output 字段")
		}
		testCaseBasics = append(testCaseBasics, &model.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: problemIdentity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		})
	}
	return testCaseBasics, nil
}
