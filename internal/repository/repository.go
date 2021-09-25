package repository

import (
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/lusingander/gogigu"
)

type RepositoryManager struct {
	*gogigu.Repository

	branchesMap map[string][]*Branch
	remotesMap  map[string][]*RemoteBranch
	tagsMap     map[string][]*Tag
}

func (m *RepositoryManager) BranchNames() []string {
	ret := make([]string, 0)
	for _, bs := range m.branchesMap {
		for _, b := range bs {
			ret = append(ret, b.name)
		}
	}
	sort.Strings(ret)
	return ret
}

func (m *RepositoryManager) RemoteBranchNames() []string {
	ret := make([]string, 0)
	for _, rs := range m.remotesMap {
		for _, r := range rs {
			ret = append(ret, r.name)
		}
	}
	sort.Strings(ret)
	return ret
}

func (m *RepositoryManager) TagNames() []string {
	ret := make([]string, 0)
	for _, ts := range m.tagsMap {
		for _, t := range ts {
			ret = append(ret, t.name)
		}
	}
	sort.Strings(ret)
	return ret
}

type Branch struct {
	name string
	hash string
}

type RemoteBranch struct {
	name string
	hash string
}

type Tag struct {
	name string
	hash string
}

func OpenGitRepository(path string) (*RepositoryManager, error) {
	src, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	repo, err := gogigu.Calculate(src, &gogigu.Option{Sort: gogigu.CommitDate})
	if err != nil {
		return nil, err
	}

	branches, remotes, tags, err := getReferences(src)
	if err != nil {
		return nil, err
	}

	rm := &RepositoryManager{
		Repository:  repo,
		branchesMap: branches,
		remotesMap:  remotes,
		tagsMap:     tags,
	}
	return rm, nil
}

func OpenGitRepositoryFromArgs(args []string) (*RepositoryManager, error) {
	if len(args) <= 1 {
		return nil, nil
	}
	return OpenGitRepository(args[1])
}

func getReferences(src *git.Repository) (map[string][]*Branch, map[string][]*RemoteBranch, map[string][]*Tag, error) {
	iter, err := src.References()
	if err != nil {
		return nil, nil, nil, err
	}
	bm := make(map[string][]*Branch)
	rm := make(map[string][]*RemoteBranch)
	tm := make(map[string][]*Tag)
	iter.ForEach(func(r *plumbing.Reference) error {
		hash := r.Hash().String()
		if r.Name().IsBranch() {
			if _, ok := bm[hash]; !ok {
				bm[hash] = make([]*Branch, 0)
			}
			branch := &Branch{
				name: r.Name().Short(),
				hash: hash,
			}
			bm[hash] = append(bm[hash], branch)
		} else if r.Name().IsRemote() {
			if _, ok := rm[hash]; !ok {
				rm[hash] = make([]*RemoteBranch, 0)
			}
			branch := &RemoteBranch{
				name: r.Name().Short(),
				hash: hash,
			}
			rm[hash] = append(rm[hash], branch)
		} else if r.Name().IsTag() {
			if _, ok := tm[hash]; !ok {
				tm[hash] = make([]*Tag, 0)
			}
			tag := &Tag{
				name: r.Name().Short(),
				hash: hash,
			}
			tm[hash] = append(tm[hash], tag)
		}
		return nil
	})
	return bm, rm, tm, nil
}
