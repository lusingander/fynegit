package repository

import (
	"github.com/go-git/go-git/v5"
	"github.com/lusingander/gogigu"
)

func OpenGitRepository(path string) (*gogigu.Repository, error) {
	src, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	repo, err := gogigu.Calculate(src)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
