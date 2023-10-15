package sql

import (
	"strings"
	"testing"

	"github.com/alibaba/ilogtail/pkg/logger"
	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/alibaba/ilogtail/plugins/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logTest struct {
	contents map[string]interface{}
	expected map[string]interface{}
}

type sqlTestCase struct {
	sql  string
	logs []logTest
}

func init() {
	logger.InitTestLogger(logger.OptionOpenMemoryReceiver)
}

func newProcessor(sql string) (*ProcessorSQL, error) {
	ctx := mock.NewEmptyContext("p", "l", "c")
	processor := &ProcessorSQL{
		SQL:        sql,
		NoKeyError: true,
	}
	err := processor.Init(ctx)
	return processor, err
}

func TestProcessorSQL(t *testing.T) {
	tests := []sqlTestCase{
		{
			sql: "select a b, a, a c from log",
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "foobar",
					},
					expected: map[string]interface{}{
						"a": "foobar",
						"b": "foobar",
						"c": "foobar",
					},
				},
				{
					contents: map[string]interface{}{
						"a": "barfoo",
					},
					expected: map[string]interface{}{
						"a": "barfoo",
						"b": "barfoo",
						"c": "barfoo",
					},
				},
			},
		},
		{
			sql: `
SELECT 
	CONCAT_WS(".", timestamp, nanosecond) AS event_time,
	event_type,
	MD5(idfa) AS idfa,
	CASE 
		WHEN user_agent LIKE "%iPhone OS%" THEN "ios"
		ELSE "android"
	END AS os,
	action,
	LOWER(element) AS element
FROM 
	log
WHERE 
	event_type = "js_error";
`,
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"timestamp":  "1234567890",
						"nanosecond": "123456789",
						"event_type": "js_error",
						"idfa":       "abcdefg",
						"user_agent": "Chrome on iOS. Mozilla/5.0 (iPhone; CPU iPhone OS 16_5_1 like Mac OS X)",
						"action":     "click",
						"element":    "#Button",
					},
					expected: map[string]interface{}{
						"event_time": "1234567890.123456789",
						"event_type": "js_error",
						"idfa":       "7ac66c0f148de9519b8bd264312c4d64",
						"os":         "ios",
						"action":     "click",
						"element":    "#button",
					},
				},
				{
					contents: map[string]interface{}{
						"timestamp":  "1234567890",
						"nanosecond": "123456789",
						"event_type": "perf",
						"idfa":       "abcdefg",
						"user_agent": "Chrome on iOS. Mozilla/5.0 (iPhone; CPU iPhone OS 16_5_1 like Mac OS X)",
						"load":       "3",
						"render":     "2",
					},
					expected: nil,
				},
			},
		},
		{
			sql: `select 123, "abc", 1.23, true, false, "a" "b" from log`,
			logs: []logTest{
				{
					contents: map[string]interface{}{},
					expected: map[string]interface{}{
						"123":   "123",
						"'abc'": "abc",
						"1.23":  "1.23",
						"true":  "1",
						"false": "0",
						"b":     "a",
					},
				},
			},
		},
		{
			sql: "select concat('a', coalesce(col1, col2), concat_ws(col3, 'c', col4)) ans from log",
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"col1": "",
						"col2": "b",
						"col3": "d",
						"col4": "e",
					},
					expected: map[string]interface{}{
						"ans": "abcde",
					},
				},
			},
		},
		{
			sql: "select substr(a, 2) c from log",
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "foobar",
					},
					expected: map[string]interface{}{
						"c": "oobar",
					},
				},
			},
		},
		{
			sql: "select substr(a, 2, 4) c from log",
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "foobar",
					},
					expected: map[string]interface{}{
						"c": "ooba",
					},
				},
			},
		},
		{
			sql: "select substr(a from 2 for 4) c from log",
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "foobar",
					},
					expected: map[string]interface{}{
						"c": "ooba",
					},
				},
			},
		},
		{
			sql: `
SELECT 
	CASE a
		WHEN 'v1' THEN "1"
		WHEN 'v2' THEN "2"
		ELSE "3"
	END AS col1
FROM log
`,
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "v1",
					},
					expected: map[string]interface{}{
						"col1": "1",
					},
				},
				{
					contents: map[string]interface{}{
						"a": "v",
					},
					expected: map[string]interface{}{
						"col1": "3",
					},
				},
			},
		},
		{
			sql: `
SELECT 
	CASE
		WHEN a > 'foo' AND TRUE THEN "1"
		WHEN NOT (a < 'd') THEN "2"
		WHEN a != 'a' THEN "3"
		ELSE "4"
	END AS col1
FROM log
`,
			logs: []logTest{
				{
					contents: map[string]interface{}{
						"a": "g",
					},
					expected: map[string]interface{}{
						"col1": "1",
					},
				},
				{
					contents: map[string]interface{}{
						"a": "e",
					},
					expected: map[string]interface{}{
						"col1": "2",
					},
				},
				{
					contents: map[string]interface{}{
						"a": "b",
					},
					expected: map[string]interface{}{
						"col1": "3",
					},
				},
				{
					contents: map[string]interface{}{
						"a": "a",
					},
					expected: map[string]interface{}{
						"col1": "4",
					},
				},
			},
		},
	}

	for _, test := range tests {
		processor, err := newProcessor(test.sql)
		require.NoError(t, err)

		logs := make([]models.PipelineEvent, len(test.logs))

		for i, logTest := range test.logs {
			log := models.NewLog("", nil, "", "", "", models.NewTags(), 0)
			log.GetIndices().AddAll(logTest.contents)
			logs[i] = log
		}

		events := &models.PipelineGroupEvents{
			Events: logs,
		}
		context := pipeline.NewObservePipelineConext(10)
		processor.Process(events, context)
		context.Collector().CollectList(events)

		for i, logTest := range test.logs {
			log := logs[i].(*models.Log)
			contents := log.GetIndices()
			if logTest.expected == nil {
				assert.Nil(t, contents)
			} else {
				assert.Equal(t, logTest.expected, contents.Iterator())
			}
		}
	}
}

func TestNoKeyError(t *testing.T) {
	processor, err := newProcessor("select b from log")
	require.NoError(t, err)

	log := models.NewLog("", nil, "", "", "", models.NewTags(), 0)
	log.GetIndices().Add("a", "test_value")

	logs := &models.PipelineGroupEvents{
		Events: []models.PipelineEvent{log},
	}
	context := pipeline.NewObservePipelineConext(10)
	processor.Process(logs, context)
	context.Collector().CollectList(logs)

	memoryLog, ok := logger.ReadMemoryLog(1)
	require.True(t, ok)
	assert.True(t, strings.Contains(memoryLog, "SQL_FIND_ALARM\tcannot find key:b"), "got: %s", memoryLog)
}
