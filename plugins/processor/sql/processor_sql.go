package sql

import (
	"fmt"

	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/xwb1989/sqlparser"
)

type ProcessorSQL struct {
	SQL          string
	stmt         *sqlparser.Select
	context      pipeline.Context
	executePlans []executePlan
}

type stringLogContents models.KeyValues[string]

type executePlan struct {
	sourceKey string
	destKey   string
	sFunc     scalarFunc
}

type scalarFunc func(string) string

const pluginName = "processor_sql"

func (p *ProcessorSQL) Init(context pipeline.Context) error {
	if p.SQL == "" {
		return fmt.Errorf("SQL can't be empty for plugin %v", pluginName)
	}

	stmt, err := sqlparser.Parse(p.SQL)
	if err != nil {
		return fmt.Errorf("sql parse error: %v", err)
	}

	var ok bool
	p.stmt, ok = stmt.(*sqlparser.Select)
	if !ok {
		return fmt.Errorf("not select")
	}

	// build execute plans
	for _, expr := range p.stmt.SelectExprs {
		//TODO: destkey can't be deplicated
		switch expr := expr.(type) {
		case *sqlparser.AliasedExpr:
			fmt.Println("Expression:", sqlparser.String(expr.Expr)) // print the expression
			sKey := sqlparser.String(expr.Expr)
			dKey := expr.As.String()
			if dKey == "" {
				dKey = sKey
			}
			plan := executePlan{
				sourceKey: sKey,
				destKey:   dKey,
			}
			p.executePlans = append(p.executePlans, plan)
			if expr.As.String() != "" {
				fmt.Println("Alias:", expr.As) // print the alias
			}
		default:
			fmt.Println("not support")
		}
	}

	p.context = context
	return nil
}

func (p *ProcessorSQL) Description() string {
	return "sql"
}

// func (p *ProcessorSQL) getDestKeys() {
// 	p.stmt.SelectExprs.
// }

func (p *ProcessorSQL) Process(in *models.PipelineGroupEvents, context pipeline.PipelineContext) {
	// fmt.Println("😶‍🌫️😶‍🌫️😶‍🌫️😶‍🌫️😶‍🌫️😶‍🌫️")

	for _, event := range in.Events {
		p.processEvent(event)
	}
	context.Collector().Collect(in.Group, in.Events...)
}

func (p *ProcessorSQL) processEvent(event models.PipelineEvent) {
	if event.GetType() != models.EventTypeLogging {
		fmt.Println("eventtypt not support")
		return
	}
	log := event.(*models.Log)

	originalContents := log.GetIndices()

	newContents := models.NewLogContents()

	for _, plan := range p.executePlans {
		newContents.Add(plan.destKey, originalContents.Get(plan.sourceKey))
	}

	log.SetIndices(newContents)
}
