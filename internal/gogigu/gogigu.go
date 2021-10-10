package gogigu

import (
	"log"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	Nodes Nodes

	nodesMap    map[string]*Node
	childrenMap map[string]Nodes
	parentsMap  map[string]Nodes
	edgesMap    map[int]Edges
	maxPosX     int
}

func (r *Repository) MaxPosX() int {
	return r.maxPosX
}

func (r *Repository) Node(hash string) *Node {
	return r.nodesMap[hash]
}

func (r *Repository) Children(hash string) Nodes {
	children, ok := r.childrenMap[hash]
	if ok {
		return children
	}
	return Nodes{}
}

func (r *Repository) ChildrenHashes(hash string) []string {
	children := r.Children(hash)
	ret := make([]string, len(children))
	for i, child := range children {
		ret[i] = child.hash
	}
	return ret
}

func (r *Repository) Parents(hash string) Nodes {
	parents, ok := r.parentsMap[hash]
	if ok {
		return parents
	}
	return Nodes{}
}

func (r *Repository) ParentsHashes(hash string) []string {
	parents := r.Parents(hash)
	ret := make([]string, len(parents))
	for i, parent := range parents {
		ret[i] = parent.hash
	}
	return ret
}

func (r *Repository) Edges(posY int) []*Edge {
	edges, ok := r.edgesMap[posY]
	if ok {
		return edges
	}
	return []*Edge{}
}

func initRepository(repo *git.Repository) (*Repository, error) {
	cIter, err := repo.Log(&git.LogOptions{All: true, Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	nodes := make(Nodes, 0)
	nodesMap := make(map[string]*Node)

	err = cIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()
		n := &Node{
			Commit: c,
			hash:   hash,
		}
		nodes = append(nodes, n)
		nodesMap[hash] = n
		return nil
	})
	if err != nil {
		return nil, err
	}

	parentsMap := make(map[string]Nodes)
	childrenMap := make(map[string]Nodes)
	for _, n := range nodes {
		parentsMap[n.hash] = make(Nodes, 0)
		for _, h := range n.Commit.ParentHashes {
			parentHash := h.String()
			if parentNode, ok := nodesMap[parentHash]; ok {
				parentsMap[n.hash] = append(parentsMap[n.hash], parentNode)
				if _, ok := childrenMap[parentHash]; !ok {
					childrenMap[parentHash] = make(Nodes, 0)
				}
				childrenMap[parentHash] = append(childrenMap[parentHash], n)
			} else {
				log.Printf("node not found: target=%s, parent=%s", n.hash, parentHash)
			}
		}
	}

	return &Repository{
		Nodes:       nodes,
		nodesMap:    nodesMap,
		childrenMap: childrenMap,
		parentsMap:  parentsMap,
	}, nil
}

type Node struct {
	Commit *object.Commit

	hash string
	posY int
	posX int
}

func (n *Node) committedAt() time.Time {
	return n.Commit.Committer.When
}

func (n *Node) Hash() string {
	return n.hash
}

func (n *Node) ShortHash() string {
	return n.hash[:7]
}

func (n *Node) PosY() int {
	return n.posY
}

func (n *Node) PosX() int {
	return n.posX
}

type Nodes []*Node

func (ns Nodes) hashes() []string {
	hs := make([]string, len(ns))
	for i, n := range ns {
		hs[i] = n.hash
	}
	return hs
}

type Option struct {
	Sort
}

type Sort int

const (
	Topological Sort = iota
	CommitDate
)

func Calculate(src *git.Repository, opt *Option) (*Repository, error) {
	repo, err := initRepository(src)
	if err != nil {
		return nil, err
	}

	sortNodes(repo, opt)
	calculatePositions(repo)
	calculateEdges(repo)

	return repo, nil
}
