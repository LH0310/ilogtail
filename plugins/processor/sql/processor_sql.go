package sql

import (
	"errors"
	"fmt"

	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/xwb1989/sqlparser"
)

type ProcessorSQL struct {
	SQL                string
	context            pipeline.Context
	newKeys            []string
	newValueEvaluators []*stringEvaluator
	whereEvaluator     condEvaluator
}

type stringLogContents models.KeyValues[string]

const pluginName = "processor_sql"

func (p *ProcessorSQL) Description() string {
	return "sql"
}

func (p *ProcessorSQL) Init(context pipeline.Context) error {
	if p.SQL == "" {
		return fmt.Errorf("SQL can't be empty for plugin %v", pluginName)
	}

	stmt, err := sqlparser.Parse(p.SQL)
	if err != nil {
		return fmt.Errorf("sql parse error: %v", err)
	}

	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		return errors.New("Not select stmt")
	}

	initScalarFuncs()

	err = p.handleSelectExprs(sel.SelectExprs)
	if err != nil {
		return err
	}
	err = p.handleWherExpr(sel.Where)
	if err != nil {
		return err
	}

	p.context = context
	return nil
}

func (p *ProcessorSQL) handleSelectExprs(sels sqlparser.SelectExprs) (err error) {
	p.newKeys = make([]string, len(sels))
	p.newValueEvaluators = make([]*stringEvaluator, len(sels))
	for i, sel := range sels {
		// TODO: add support for StarExpr
		aliaExpr := sel.(*sqlparser.AliasedExpr)
		if aliaExpr.As.IsEmpty() {
			p.newKeys[i] = sqlparser.String(aliaExpr.Expr)
		} else {
			p.newKeys[i] = aliaExpr.As.String()
		}

		p.newValueEvaluators[i], err = compileStringExpr(&aliaExpr.Expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ProcessorSQL) handleWherExpr(where *sqlparser.Where) (err error) {
	if where == nil {
		return
	}
	p.whereEvaluator, err = compileCondExpr(&where.Expr)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProcessorSQL) Process(in *models.PipelineGroupEvents, context pipeline.PipelineContext) {
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

	originalContents, err := toStringLogContents(log.GetIndices())
	if err != nil {
		panic("Not string log")
	}

	if p.whereEvaluator != nil && !p.whereEvaluator(originalContents) {
		log.SetIndices(nil)
		return
	}

	newContents := models.NewLogContents()

	for i, eval := range p.newValueEvaluators {
		v := eval.evaluate(originalContents)
		newContents.Add(p.newKeys[i], v)
	}

	log.SetIndices(newContents)
}

func toStringLogContents(logContents models.LogContents) (stringLogContents, error) {
	slc := newStringLogContents()
	for key, value := range logContents.Iterator() {
		value, ok := value.(string)
		if !ok {
			return nil, errors.New("not stringLogContents")
		}
		slc.Add(key, value)
	}
	return slc, nil
}

func newStringLogContents() stringLogContents {
	return models.NewKeyValues[string]()
}
