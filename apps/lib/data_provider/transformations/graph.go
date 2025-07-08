package transformations

import (
	"fmt"
	"math"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

const TransformationDataSourceId = types.DataSourceId("transformation")

type TransformationGraph struct {
	dependencyGraph       *simple.DirectedGraph
	orderedNodes          []graph.Node
	nodeToValueId         map[graph.Node]types.ValueId
	valueIdToNode         map[types.ValueId]graph.Node
	parsedTransformations map[types.ValueId]*Expression
	currentVals           map[string]types.DataSourceValueUpdate
}

func NewTransformationGraph(
	dependencyGraph *simple.DirectedGraph,
	orderedNodes []graph.Node,
	nodeToValueId map[graph.Node]types.ValueId,
	valueIdToNode map[types.ValueId]graph.Node,
	parsedTransformations map[types.ValueId]*Expression,
) *TransformationGraph {
	return &TransformationGraph{
		dependencyGraph:       dependencyGraph,
		orderedNodes:          orderedNodes,
		nodeToValueId:         nodeToValueId,
		valueIdToNode:         valueIdToNode,
		parsedTransformations: parsedTransformations,
		currentVals:           make(map[string]types.DataSourceValueUpdate),
	}
}

func (tg *TransformationGraph) ProcessSourceUpdates(sourceUpdates types.DataSourceUpdateMap) types.DataSourceUpdateMap {
	finalUpdateMap := make(types.DataSourceUpdateMap)

	updateTime := time.Now()

	// do a breadth-first traversal get all affected nodes
	dirtyTransformationNodes := make(map[graph.Node]interface{})
	queue := make([]graph.Node, 0)
	for valueId, sourceUpdate := range sourceUpdates {
		queue = append(queue, tg.valueIdToNode[valueId])
		finalUpdateMap[valueId] = sourceUpdate
		tg.currentVals[string(valueId)] = sourceUpdate
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		nodeIterator := tg.dependencyGraph.From(current.ID())
		for nodeIterator.Next() {
			nextNode := nodeIterator.Node()
			if _, seen := dirtyTransformationNodes[nextNode]; !seen {
				dirtyTransformationNodes[nextNode] = struct{}{}
				queue = append(queue, nextNode)
			}
		}
	}

	// update dirty transformations in topological order
	for _, node := range tg.orderedNodes {
		if _, isDirty := dirtyTransformationNodes[node]; isDirty {
			transformationValueId := tg.nodeToValueId[node]
			transformation := tg.parsedTransformations[transformationValueId]
			transformationValue := transformation.Eval(tg.currentVals)
			if math.IsNaN(transformationValue) {
				continue
			}

			computed := types.DataSourceValueUpdate{
				ValueId:      transformationValueId,
				DataSourceId: TransformationDataSourceId,
				Time:         updateTime,
				Value:        transformationValue,
			}
			finalUpdateMap[transformationValueId] = computed
			tg.currentVals[string(transformationValueId)] = computed
		}
	}

	return finalUpdateMap
}

func BuildTransformationGraph(transformations []types.DataProviderTransformationConfig, sourceIds map[types.ValueId]interface{}) (*TransformationGraph, error) {
	g := simple.NewDirectedGraph()

	// allow translating node <-> value id
	nodeToValueId := make(map[graph.Node]types.ValueId)
	valueIdToNode := make(map[types.ValueId]graph.Node)

	parsedTransformations := make(map[types.ValueId]*Expression)
	for _, transformation := range transformations {
		expr, err := parse(transformation.Formula)
		if err != nil {
			return nil, err
		}
		parsedTransformations[transformation.Id] = expr

		node := g.NewNode()
		g.AddNode(node)
		nodeToValueId[node] = transformation.Id
		if _, exists := valueIdToNode[transformation.Id]; exists {
			return nil, fmt.Errorf("duplicate value id: %v", transformation.Id)
		}
		valueIdToNode[transformation.Id] = node
	}

	for sourceId := range sourceIds {
		node := g.NewNode()
		g.AddNode(node)
		nodeToValueId[node] = sourceId
		if _, exists := valueIdToNode[sourceId]; exists {
			return nil, fmt.Errorf("duplicate value id: %v", sourceId)
		}
		valueIdToNode[sourceId] = node
	}

	for _, transformation := range transformations {
		expr, ok := parsedTransformations[transformation.Id]
		if !ok {
			return nil, fmt.Errorf("no such transformation: %s", transformation.Id)
		}

		deps := expr.getDependencies()
		for _, dep := range deps {
			_, sourcePriceExists := sourceIds[types.ValueId(dep)]
			if !sourcePriceExists {
				_, transformationExists := valueIdToNode[types.ValueId(dep)]
				if !transformationExists {
					return nil, fmt.Errorf("no such source or transformation id: %s", dep)
				}
			}

			g.SetEdge(g.NewEdge(valueIdToNode[types.ValueId(dep)], valueIdToNode[transformation.Id]))
		}
	}

	orderedNodes, err := topo.Sort(g)
	if err != nil {
		return nil, fmt.Errorf("could not linearize price id graph - there may be circular dependencies: %v", err)
	}

	transformationGraph := NewTransformationGraph(g, orderedNodes, nodeToValueId, valueIdToNode, parsedTransformations)

	return transformationGraph, nil
}
