package pathiterator

import (
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (tn *TreeNode) GetFuserMode() int {
	mode := fuse.S_IFREG
	if tn.IsDir {
		mode = fuse.S_IFDIR
	}
	return mode
}

func (tn *TreeNode) ReadDir(r *[]fuse.DirEntry) {
	for name, child := range tn.Children {
		mode := child.GetFuserMode()
		d := fuse.DirEntry{
			Name: name,
			Mode: uint32(mode),
		}
		*r = append(*r, d)
	}
}

func (tn *TreeNode) LookUp(name *string, mode *int) (*TreeNode, bool) {
	child, exists := tn.Children[*name]
	if exists {
		*mode = child.GetFuserMode()
		return child, true
	}

	return nil, false
}
