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

const TransformationDataSourceID = types.DataSourceID("transformation")

type TransformationGraph struct {
	dependencyGraph       *simple.DirectedGraph
	orderedNodes          []graph.Node
	nodeToValueID         map[graph.Node]types.ValueID
	valueIDToNode         map[types.ValueID]graph.Node
	parsedTransformations map[types.ValueID]*Expression
	currentVals           map[string]types.DataSourceValueUpdate
}

func NewTransformationGraph(
	dependencyGraph *simple.DirectedGraph,
	orderedNodes []graph.Node,
	nodeToValueID map[graph.Node]types.ValueID,
	valueIDToNode map[types.ValueID]graph.Node,
	parsedTransformations map[types.ValueID]*Expression,
) *TransformationGraph {
	return &TransformationGraph{
		dependencyGraph:       dependencyGraph,
		orderedNodes:          orderedNodes,
		nodeToValueID:         nodeToValueID,
		valueIDToNode:         valueIDToNode,
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
	for valueID, sourceUpdate := range sourceUpdates {
		queue = append(queue, tg.valueIDToNode[valueID])
		finalUpdateMap[valueID] = sourceUpdate
		tg.currentVals[string(valueID)] = sourceUpdate
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
			transformationValueID := tg.nodeToValueID[node]
			transformation := tg.parsedTransformations[transformationValueID]
			transformationValue := transformation.Eval(tg.currentVals)
			if math.IsNaN(transformationValue) {
				continue
			}

			computed := types.DataSourceValueUpdate{
				ValueID:      transformationValueID,
				DataSourceID: TransformationDataSourceID,
				Time:         updateTime,
				Value:        transformationValue,
			}
			finalUpdateMap[transformationValueID] = computed
			tg.currentVals[string(transformationValueID)] = computed
		}
	}

	return finalUpdateMap
}

func BuildTransformationGraph(
	transformations []types.DataProviderTransformationConfig,
	sourceIDs map[types.ValueID]any,
) (*TransformationGraph, error) {
	g := simple.NewDirectedGraph()

	// allow translating node <-> value id
	nodeToValueID := make(map[graph.Node]types.ValueID)
	valueIDToNode := make(map[types.ValueID]graph.Node)

	parsedTransformations := make(map[types.ValueID]*Expression)
	for _, transformation := range transformations {
		expr, err := parse(transformation.Formula)
		if err != nil {
			return nil, err
		}
		parsedTransformations[transformation.ID] = expr

		node := g.NewNode()
		g.AddNode(node)
		nodeToValueID[node] = transformation.ID
		if _, exists := valueIDToNode[transformation.ID]; exists {
			return nil, fmt.Errorf("duplicate value id: %v", transformation.ID)
		}
		valueIDToNode[transformation.ID] = node
	}

	for sourceID := range sourceIDs {
		node := g.NewNode()
		g.AddNode(node)
		nodeToValueID[node] = sourceID
		if _, exists := valueIDToNode[sourceID]; exists {
			return nil, fmt.Errorf("duplicate value id: %v", sourceID)
		}
		valueIDToNode[sourceID] = node
	}

	for _, transformation := range transformations {
		expr, ok := parsedTransformations[transformation.ID]
		if !ok {
			return nil, fmt.Errorf("no such transformation: %s", transformation.ID)
		}

		deps := expr.getDependencies()
		for _, dep := range deps {
			_, sourcePriceExists := sourceIDs[types.ValueID(dep)]
			if !sourcePriceExists {
				_, transformationExists := valueIDToNode[types.ValueID(dep)]
				if !transformationExists {
					return nil, fmt.Errorf("no such source or transformation id: %s", dep)
				}
			}

			g.SetEdge(g.NewEdge(valueIDToNode[types.ValueID(dep)], valueIDToNode[transformation.ID]))
		}
	}

	orderedNodes, err := topo.Sort(g)
	if err != nil {
		return nil, fmt.Errorf("could not linearize price id graph - there may be circular dependencies: %v", err)
	}

	transformationGraph := NewTransformationGraph(g, orderedNodes, nodeToValueID, valueIDToNode, parsedTransformations)

	return transformationGraph, nil
}
