package transformations

import (
	"fmt"
	"math"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

const TransformationDataSourceId = types.DataSourceID("transformation")

type TransformationGraph struct {
	dependencyGraph       *simple.DirectedGraph
	orderedNodes          []graph.Node
	nodeToValueId         map[graph.Node]types.ValueID
	valueIdToNode         map[types.ValueID]graph.Node
	parsedTransformations map[types.ValueID]*Expression
	currentVals           map[string]types.DataSourceValueUpdate
}

func NewTransformationGraph(
	dependencyGraph *simple.DirectedGraph,
	orderedNodes []graph.Node,
	nodeToValueId map[graph.Node]types.ValueID,
	valueIdToNode map[types.ValueID]graph.Node,
	parsedTransformations map[types.ValueID]*Expression,
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
	dirtyTransformationNodes := make(map[graph.Node]any)
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
				ValueID:      transformationValueId,
				DataSourceID: TransformationDataSourceId,
				Time:         updateTime,
				Value:        transformationValue,
			}
			finalUpdateMap[transformationValueId] = computed
			tg.currentVals[string(transformationValueId)] = computed
		}
	}

	return finalUpdateMap
}

func BuildTransformationGraph(
	transformations []types.DataProviderTransformationConfig,
	sourceIds map[types.ValueID]any,
) (*TransformationGraph, error) {
	g := simple.NewDirectedGraph()

	// allow translating node <-> value id
	nodeToValueId := make(map[graph.Node]types.ValueID)
	valueIdToNode := make(map[types.ValueID]graph.Node)

	parsedTransformations := make(map[types.ValueID]*Expression)
	for _, transformation := range transformations {
		expr, err := parse(transformation.Formula)
		if err != nil {
			return nil, err
		}
		parsedTransformations[transformation.ID] = expr

		node := g.NewNode()
		g.AddNode(node)
		nodeToValueId[node] = transformation.ID
		if _, exists := valueIdToNode[transformation.ID]; exists {
			return nil, fmt.Errorf("duplicate value id: %v", transformation.ID)
		}
		valueIdToNode[transformation.ID] = node
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
		expr, ok := parsedTransformations[transformation.ID]
		if !ok {
			return nil, fmt.Errorf("no such transformation: %s", transformation.ID)
		}

		deps := expr.getDependencies()
		for _, dep := range deps {
			_, sourcePriceExists := sourceIds[types.ValueID(dep)]
			if !sourcePriceExists {
				_, transformationExists := valueIdToNode[types.ValueID(dep)]
				if !transformationExists {
					return nil, fmt.Errorf("no such source or transformation id: %s", dep)
				}
			}

			g.SetEdge(g.NewEdge(valueIdToNode[types.ValueID(dep)], valueIdToNode[transformation.ID]))
		}
	}

	orderedNodes, err := topo.Sort(g)
	if err != nil {
		return nil, fmt.Errorf("could not linearize price id graph - there may be circular dependencies: %v", err)
	}

	transformationGraph := NewTransformationGraph(g, orderedNodes, nodeToValueId, valueIdToNode, parsedTransformations)

	return transformationGraph, nil
}
