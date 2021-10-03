package gogigu

type EdgeType int

const (
	EdgeStraight EdgeType = iota
	EdgeUp
	EdgeDown
	EdgeBranch
	EdgeMerge
)

type Edge struct {
	EdgeType
	PosX int
}

type Edges []*Edge

func calculateEdges(repo *Repository) {
	edges := make(map[int]Edges)
	for i := range repo.Nodes {
		edges[i] = make(Edges, 0)
	}
	for _, n := range repo.Nodes {
		h := n.Hash()
		for _, child := range repo.Children(h) {
			edges[n.PosY()] = append(edges[n.PosY()], &Edge{EdgeUp, n.PosX()})
			if n.PosX() == child.PosX() {
				for y := n.PosY() - 1; y > child.PosY(); y-- {
					edges[y] = append(edges[y], &Edge{EdgeStraight, n.PosX()})
				}
			} else if n.PosX() < child.PosX() {
				edges[n.PosY()] = append(edges[n.PosY()], &Edge{EdgeBranch, child.PosX()})
				for y := n.PosY() - 1; y > child.PosY(); y-- {
					edges[y] = append(edges[y], &Edge{EdgeStraight, child.PosX()})
				}
			}
		}
		for _, parent := range repo.Parents(h) {
			edges[n.PosY()] = append(edges[n.PosY()], &Edge{EdgeDown, n.PosX()})
			if n.PosX() < parent.PosX() {
				edges[n.PosY()] = append(edges[n.PosY()], &Edge{EdgeMerge, parent.PosX()})
				for y := n.PosY() + 1; y < parent.PosY(); y++ {
					edges[y] = append(edges[y], &Edge{EdgeStraight, parent.PosX()})
				}
			}
		}
	}
	repo.edgesMap = edges
}
