package common

var (
	RootDir    = ".minigit"
	ObjectDir  = "objects"
	RefsDir    = "refs"
	HEAD       = "HEAD"
	IndexFile  = "index.json"
	CommitFile = "commit"
	TreeFile   = "tree"
	BlobFile   = "blob"
	HeadDir    = "heads"
)

type Index map[string]string

type TreeNode struct {
	Sha  string
	Name string
	Mode string
}