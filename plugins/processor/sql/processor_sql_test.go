package sql

import (
	"testing"

	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/alibaba/ilogtail/plugins/test/mock"
	"github.com/stretchr/testify/assert"
)

func newProcessor(sql string) (*ProcessorSQL, error) {
	ctx := mock.NewEmptyContext("p", "l", "c")
	processor := &ProcessorSQL{
		SQL: sql,
	}
	err := processor.Init(ctx)
	return processor, err
}

func TestFeasibility(t *testing.T) {
	processor, err := newProcessor("select a b, a, a c from log")
	if err != nil {
		println(err)
	}

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
