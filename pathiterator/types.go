package pathiterator

type TreeNode struct {
	FullPath string
	Children map[string]*TreeNode
	IsDir    bool
	Parent   *TreeNode
}

type StrTreePair struct {
	path     string
	treeNode *TreeNode
}

type Queue []StrTreePair

type FileHashToPath map[string][]string
