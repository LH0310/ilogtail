package sql

import (
	"testing"

	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/alibaba/ilogtail/plugins/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProcessor(sql string) (*ProcessorSQL, error) {
	ctx := mock.NewEmptyContext("p", "l", "c")
	processor := &ProcessorSQL{
		SQL: sql,
	}
	err := processor.Init(ctx)
	return processor, err
}

func TestRename(t *testing.T) {
	processor, err := newProcessor("select a b, a, a c from log")
	require.NoError(t, err)

	log := models.NewLog("", nil, "", "", "", models.NewTags(), 0)
	contents := log.GetIndices()
	contents.Add("a", "foobar")

	logs := &models.PipelineGroupEvents{
		Events: []models.PipelineEvent{log},
	}
	context := pipeline.NewObservePipelineConext(10)
	processor.Process(logs, context)
	context.Collector().CollectList(logs)

	expectedContents := map[string]interface{}{
		"a": "foobar",
		"b": "foobar",
		"c": "foobar",
	}
	contents = log.GetIndices()
	assert.Equal(t, expectedContents, contents.Iterator())

	// assert.True(t, contents.Contains("b"))
	// assert.Equal(t, "foobar", contents.Get("b"))
	// memoryLog, ok := logger.ReadMemoryLog(1)

}

func TestExample(t *testing.T) {
	sql := `
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
`
	processor, err := newProcessor(sql)

	require.NoError(t, err)
	log0 := models.NewLog("", nil, "", "", "", models.NewTags(), 0)
	log0.GetIndices().AddAll(map[string]interface{}{
		"timestamp":  "1234567890",
		"nanosecond": "123456789",
		"event_type": "js_error",
		"idfa":       "abcdefg",
		"user_agent": "Chrome on iOS. Mozilla/5.0 (iPhone; CPU iPhone OS 16_5_1 like Mac OS X)",
		"action":     "click",
		"element":    "#Button",
	})

	logs := &models.PipelineGroupEvents{
		Events: []models.PipelineEvent{log0},
	}
	context := pipeline.NewObservePipelineConext(10)
	processor.Process(logs, context)
	context.Collector().CollectList(logs)

	expectedContents := map[string]interface{}{
		"event_time": "1234567890.123456789",
		"event_type": "js_error",
		"idfa":       "7ac66c0f148de9519b8bd264312c4d64",
		"os":         "ios",
		"action":     "click",
		"element":    "#button",
	}

	contents := log0.GetIndices()
	assert.Equal(t, expectedContents, contents.Iterator())
}
