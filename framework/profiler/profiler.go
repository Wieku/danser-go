package profiler

import "github.com/wieku/danser-go/framework/qpc"

type ProfileType string

const (
	PRoot        = "Root"
	PUpdate      = "Update"
	PSched       = "Scheduler"
	PDraw        = "Draw"
	PSleep       = "Sleep"
	PInput       = "Input"
	PSwapBuffers = "Swap Buffers"
)

type PNode struct {
	parent *PNode

	NodeName string
	NodeType ProfileType

	Nodes []*PNode

	TimeTotal float64

	lastStartTime float64
}

func (node *PNode) GetFullName() string {
	return "[" + node.NodeName + "] - " + string(node.NodeType)
}

var previousNode *PNode

var currentNode *PNode

func Reset() {
	if currentNode != nil && currentNode.parent != nil {
		panic("profiler tree has not been closed")
	}

	previousNode = currentNode
	currentNode = nil
}

func GetLastProfileResult() *PNode {
	return previousNode
}

func StartGroup(funcName string, pType ProfileType) {
	var node *PNode

	if currentNode != nil {
		for i := 0; i < len(currentNode.Nodes); i++ {
			if currentNode.Nodes[i].NodeName == funcName && currentNode.Nodes[i].NodeType == pType {
				node = currentNode.Nodes[i]
				break
			}
		}
	}

	if node == nil {
		node = &PNode{
			parent:   currentNode,
			NodeName: funcName,
			NodeType: pType,
		}

		if currentNode != nil {
			currentNode.Nodes = append(currentNode.Nodes, node)
		}
	}

	node.lastStartTime = qpc.GetMilliTimeF()

	currentNode = node
}

func EndGroup() {
	if currentNode == nil {
		panic("profiler tree is empty")
	}

	currentNode.TimeTotal = qpc.GetMilliTimeF() - currentNode.lastStartTime

	if currentNode.parent != nil {
		currentNode = currentNode.parent
	}
}
