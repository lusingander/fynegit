package gogigu

import (
	"fmt"
	"strings"
)

func calculatePositions(repo *Repository) error {
	ns := repo.Nodes
	activeNodes := make(Nodes, 0)
	maxPosX := 0
	for i, n := range ns {
		childrenHashes := filteredChildrenHashes(n, repo)
		if len(childrenHashes) > 0 {
			updated := false
			for _, childHash := range childrenHashes {
				if isIn(childHash, activeNodes.hashes()) {
					activeNodes = updateActiveNodes(n, activeNodes, childHash, childrenHashes)
					updated = true
					break
				}
			}
			if !updated {
				activeNodes = append(activeNodes, n)
			}
		} else {
			activeNodes = append(activeNodes, n)
		}

		decidePositionX(n, activeNodes)
		decidePositionY(n, i)

		if maxPosX < n.posX {
			maxPosX = n.posX
		}
	}
	repo.maxPosX = maxPosX
	return nil
}

func filteredChildrenHashes(n *Node, repo *Repository) []string {
	hs := make([]string, 0)
	childrenHashes := repo.ChildrenHashes(n.hash)
	for _, h := range childrenHashes {
		childParentsHashes := repo.ParentsHashes(h)
		if len(childParentsHashes) > 0 && childParentsHashes[0] == n.hash {
			hs = append(hs, h)
		}
	}
	return hs
}

func isIn(targetHash string, hashes []string) bool {
	for _, hash := range hashes {
		if hash == targetHash {
			return true
		}
	}
	return false
}

func updateActiveNodes(targetNode *Node, activeNodes Nodes, targetChildHash string, childrenHashes []string) Nodes {
	newActiveNodes := Nodes{}
	for _, n := range activeNodes {
		if n.hash == targetChildHash {
			newActiveNodes = append(newActiveNodes, targetNode)
		} else if !isIn(n.hash, childrenHashes) {
			newActiveNodes = append(newActiveNodes, n)
		}
	}
	return newActiveNodes
}

func decidePositionX(target *Node, activeNodes Nodes) {
	for i, n := range activeNodes {
		if target == n {
			target.posX = i
			return
		}
	}
	panic(fmt.Errorf("node not found [target = %v, activeNodes = %v]", target.debugString(), activeNodes.debugString()))
}

func decidePositionY(target *Node, i int) {
	target.posY = i
}

func (n *Node) debugString() string {
	hash := n.ShortHash()
	when := n.committedAt()
	msg := strings.Split(n.Commit.Message, "\n")[0]
	return fmt.Sprintf("[%v, %v, %v]", hash, when, msg)
}

func (ns Nodes) debugString() string {
	s := ""
	for _, n := range ns {
		s += n.debugString()
		s += ", "
	}
	return s
}
