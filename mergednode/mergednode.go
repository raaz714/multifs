package mergednode

import (
	"context"
	"multifs/pathiterator"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
)

type MergedNode struct {
	fs.Inode
	treeNode *pathiterator.TreeNode
}

func NewMergedNode() *MergedNode {
	return &MergedNode{treeNode: pathiterator.RootTree}
}

var _ = (fs.NodeGetattrer)((*MergedNode)(nil))

func (mn *MergedNode) Getattr(
	ctx context.Context,
	f fs.FileHandle,
	out *fuse.AttrOut,
) syscall.Errno {
	if f != nil {
		if fga, ok := f.(fs.FileGetattrer); ok {
			return fga.Getattr(ctx, out)
		}
	}
	var p string

	if mn.treeNode.FullPath == "" {
		out.Mode = fuse.S_IFDIR
		return fs.OK
	} else {
		p = mn.treeNode.FullPath
	}

	var err error
	st := syscall.Stat_t{}
	err = syscall.Lstat(p, &st)
	if err != nil {
		return fs.ToErrno(err)
	}
	out.FromStat(&st)
	return fs.OK
}

var _ = (fs.NodeSetattrer)((*MergedNode)(nil))

func (mn *MergedNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	p := mn.treeNode.FullPath
	fsa, ok := f.(fs.FileSetattrer)
	if ok && fsa != nil {
		fsa.Setattr(ctx, in, out)
	} else {
		if m, ok := in.GetMode(); ok {
			if err := syscall.Chmod(p, m); err != nil {
				return fs.ToErrno(err)
			}
		}

		uid, uok := in.GetUID()
		gid, gok := in.GetGID()
		if uok || gok {
			suid := -1
			sgid := -1
			if uok {
				suid = int(uid)
			}
			if gok {
				sgid = int(gid)
			}
			if err := syscall.Chown(p, suid, sgid); err != nil {
				return fs.ToErrno(err)
			}
		}

		mtime, mok := in.GetMTime()
		atime, aok := in.GetATime()

		if mok || aok {
			ta := unix.Timespec{Nsec: unix.UTIME_OMIT}
			tm := unix.Timespec{Nsec: unix.UTIME_OMIT}
			var err error
			if aok {
				ta, err = unix.TimeToTimespec(atime)
				if err != nil {
					return fs.ToErrno(err)
				}
			}
			if mok {
				tm, err = unix.TimeToTimespec(mtime)
				if err != nil {
					return fs.ToErrno(err)
				}
			}
			ts := []unix.Timespec{ta, tm}
			if err := unix.UtimesNanoAt(unix.AT_FDCWD, p, ts, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return fs.ToErrno(err)
			}
		}

		if sz, ok := in.GetSize(); ok {
			if err := syscall.Truncate(p, int64(sz)); err != nil {
				return fs.ToErrno(err)
			}
		}
	}

	fga, ok := f.(fs.FileGetattrer)
	if ok && fga != nil {
		fga.Getattr(ctx, out)
	} else {
		st := syscall.Stat_t{}
		err := syscall.Lstat(p, &st)
		if err != nil {
			return fs.ToErrno(err)
		}
		out.FromStat(&st)
	}
	return fs.OK
}

var _ = (fs.NodeOpener)((*MergedNode)(nil))

func (mn *MergedNode) Open(
	ctx context.Context,
	flags uint32,
) (fs.FileHandle, uint32, syscall.Errno) {
	p := pathiterator.GetFilePath(mn.treeNode.FullPath)
	log.Debug("Path map", mn.treeNode.FullPath, p)
	fl := flags // &^ syscall.O_APPEND
	f, err := syscall.Open(p, int(fl), 0)
	if err != nil {
		log.Error(err)
		return nil, 0, fs.ToErrno(err)
	}
	lf := fs.NewLoopbackFile(f)
	return lf, fuse.FOPEN_KEEP_CACHE, 0
}

var _ = (fs.NodeReaddirer)((*MergedNode)(nil))

func (mn *MergedNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	var r []fuse.DirEntry
	mn.treeNode.ReadDir(&r)

	return fs.NewListDirStream(r), fs.OK
}

var _ = (fs.NodeLookuper)((*MergedNode)(nil))

func (mn *MergedNode) Lookup(
	ctx context.Context,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	var mode int
	treeNode, exists := mn.treeNode.LookUp(&name, &mode)

	if exists {
		stable := fs.StableAttr{
			Mode: uint32(mode),
		}
		ops := &MergedNode{treeNode: treeNode}
		return mn.NewInode(ctx, ops, stable), 0
	}

	return nil, fs.ENOATTR
}
