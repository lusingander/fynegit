package repository

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/lusingander/gogigu"
)

type RefType int

const (
	Branch RefType = iota
	RemoteBranch
	Tag
)

type Ref struct {
	refType RefType
	name    string
	hash    string
}

func (r *Ref) RefType() RefType {
	return r.refType
}

func (r *Ref) Name() string {
	return r.name
}

func (r *Ref) Hash() string {
	return r.hash
}

type RepositoryManager struct {
	*gogigu.Repository

	branchesMap map[string][]*Ref
	remotesMap  map[string][]*Ref
	tagsMap     map[string][]*Ref
}

func (m *RepositoryManager) AllRefs(hash string) []*Ref {
	refs := make([]*Ref, 0)
	if ts, ok := m.tagsMap[hash]; ok {
		refs = append(refs, ts...)
	}
	if bs, ok := m.branchesMap[hash]; ok {
		refs = append(refs, bs...)
	}
	if rs, ok := m.remotesMap[hash]; ok {
		refs = append(refs, rs...)
	}
	return refs
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

func (m *RepositoryManager) FromRefName(name string) *Ref {
	if ref := fromRefNameFrom(m.branchesMap, name); ref != nil {
		return ref
	}
	if ref := fromRefNameFrom(m.remotesMap, name); ref != nil {
		return ref
	}
	if ref := fromRefNameFrom(m.tagsMap, name); ref != nil {
		return ref
	}
	return nil
}

func fromRefNameFrom(refs map[string][]*Ref, name string) *Ref {
	for _, rs := range refs {
		for _, r := range rs {
			if r.name == name {
				return r
			}
		}
	}
	return nil
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

func getReferences(src *git.Repository) (map[string][]*Ref, map[string][]*Ref, map[string][]*Ref, error) {
	iter, err := src.References()
	if err != nil {
		return nil, nil, nil, err
	}
	bm := make(map[string][]*Ref)
	rm := make(map[string][]*Ref)
	tm := make(map[string][]*Ref)
	iter.ForEach(func(r *plumbing.Reference) error {
		hash := r.Hash().String()
		if r.Name().IsBranch() {
			if _, ok := bm[hash]; !ok {
				bm[hash] = make([]*Ref, 0)
			}
			branch := &Ref{
				refType: Branch,
				name:    r.Name().Short(),
				hash:    hash,
			}
			bm[hash] = append(bm[hash], branch)
		} else if r.Name().IsRemote() {
			if _, ok := rm[hash]; !ok {
				rm[hash] = make([]*Ref, 0)
			}
			remote := &Ref{
				refType: RemoteBranch,
				name:    r.Name().Short(),
				hash:    hash,
			}
			rm[hash] = append(rm[hash], remote)
		} else if r.Name().IsTag() {
			if _, ok := tm[hash]; !ok {
				tm[hash] = make([]*Ref, 0)
			}
			tag := &Ref{
				refType: Tag,
				name:    r.Name().Short(),
				hash:    hash,
			}
			tm[hash] = append(tm[hash], tag)
		}
		return nil
	})
	return bm, rm, tm, nil
}

type PatchFileDetail struct {
	name       string
	changeType ChangeType
}

func (d *PatchFileDetail) Name() string {
	return d.name
}

func (d *PatchFileDetail) ChangeType() ChangeType {
	return d.changeType
}

type ChangeType int

const (
	_ ChangeType = iota
	Modify
	Insert
	Delete
	Move
)

func (m *RepositoryManager) PatchFileDetails(target *gogigu.Node) ([]*PatchFileDetail, error) {
	ps := m.Parents(target.Hash())
	if len(ps) == 0 {
		return []*PatchFileDetail{}, nil
	}
	nt, err := target.Commit.Tree()
	if err != nil {
		return nil, err
	}
	pt, err := ps[0].Commit.Tree()
	if err != nil {
		return nil, err
	}
	changes, err := pt.Diff(nt)
	if err != nil {
		return nil, err
	}
	ds := make([]*PatchFileDetail, 0)
	for _, change := range changes {
		changeType, err := changeTypeFrom(change)
		if err != nil {
			return nil, err
		}
		d := &PatchFileDetail{
			name:       nameFrom(change, changeType),
			changeType: changeType,
		}
		ds = append(ds, d)
	}
	return ds, nil
}

func changeTypeFrom(change *object.Change) (ChangeType, error) {
	a, err := change.Action()
	if err != nil {
		return 0, err
	}
	switch a {
	case merkletrie.Modify:
		if change.From.Name == change.To.Name {
			return Modify, nil
		} else {
			return Move, nil
		}
	case merkletrie.Insert:
		return Insert, nil
	case merkletrie.Delete:
		return Delete, nil
	default:
		return 0, fmt.Errorf("invalid type: %v", a)
	}
}

func nameFrom(c *object.Change, ct ChangeType) string {
	if ct == Delete {
		return c.From.Name
	}
	return c.To.Name
}
