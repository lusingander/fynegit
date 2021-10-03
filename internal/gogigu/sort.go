package gogigu

import "sort"

func sortNodes(repo *Repository, opt *Option) {
	ns := repo.Nodes
	sort.Slice(ns, func(i, j int) bool {
		return ns[i].committedAt().Before(ns[j].committedAt())
	})
	switch opt.Sort {
	case Topological:
		repo.Nodes = dfsTopologicalSort(ns, repo)
	case CommitDate:
		repo.Nodes = bfsTopologicalSort(ns, repo)
	}
}

func bfsTopologicalSort(ns Nodes, repo *Repository) Nodes {
	sorted := make([]*Node, 0, len(ns))
	visited := make(map[string]struct{})
	q := NewQueue()
	var bfs func(n *Node)
	bfs = func(n *Node) {
		if _, ok := visited[n.hash]; ok {
			return
		}
		visited[n.hash] = struct{}{}
		children := repo.Children(n.hash)
		for _, child := range children {
			q.Enqueue(child)
		}
		for len(*q) > 0 {
			bfs(q.Dequeue())
		}
		sorted = append(sorted, n)
	}
	for _, n := range ns {
		bfs(n)
	}
	return sorted
}

func dfsTopologicalSort(ns Nodes, repo *Repository) Nodes {
	sorted := make([]*Node, 0, len(ns))
	visited := make(map[string]struct{})
	var dfs func(n *Node)
	dfs = func(n *Node) {
		if _, ok := visited[n.hash]; ok {
			return
		}
		visited[n.hash] = struct{}{}
		children := repo.Children(n.hash)
		for _, child := range children {
			dfs(child)
		}
		sorted = append(sorted, n)
	}
	for _, n := range ns {
		dfs(n)
	}
	return sorted
}
