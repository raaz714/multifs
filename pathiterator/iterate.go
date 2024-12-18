package pathiterator

import (
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jwangsadinata/go-multimap/slicemultimap"

	xxHash "github.com/cespare/xxhash"
)

type MulMap struct {
	*slicemultimap.MultiMap
	sync.RWMutex
}

type SafeMap struct {
	StrMap map[string]string
	mutex  sync.RWMutex
}

var (
	RootTree *TreeNode = nil
	FileHToP MulMap    = MulMap{MultiMap: slicemultimap.New()}
	FilePToH SafeMap   = SafeMap{StrMap: make(map[string]string)}
)

func NewTree(roots []string, mountDir string) error {
	if RootTree != nil {
		return nil
	}
	q := &Queue{}
	RootTree = &TreeNode{FullPath: "", IsDir: true, Children: make(map[string]*TreeNode)}
	for _, root := range roots {
		pair := &StrTreePair{path: root, treeNode: RootTree}
		q.Enqueue(pair)
	}

	var fileList []string = []string{}

	for !q.Empty() {
		tempQ := &Queue{}

		for !q.Empty() {
			topPair, err := q.Dequeue()
			if err != nil {
				break
			}

			files, err := os.ReadDir(topPair.path)
			if err != nil {
				continue
			}

			for _, file := range files {
				t, exists := topPair.treeNode.Children[file.Name()]
				if !exists {
					fullPath := filepath.Join(topPair.path, file.Name())
					t = &TreeNode{FullPath: fullPath, IsDir: file.IsDir(), Children: make(map[string]*TreeNode)}
					topPair.treeNode.Children[file.Name()] = t
					t.Parent = topPair.treeNode
				}
				if file.IsDir() {
					newPath := filepath.Join(topPair.path, file.Name())
					newPair := &StrTreePair{path: newPath, treeNode: t}
					tempQ.Enqueue(newPair)
				} else {
					newPath := filepath.Join(topPair.path, file.Name())
					fileList = append(fileList, newPath)
					// go hashFileAndStoreMap(newPath)
				}
			}
		}
		q.EnqueueArray(*tempQ)
	}

	go processHashFiles(&fileList)

	return nil
}

func populateFromCachedHash() {
	homeDir, _ := os.UserHomeDir()
	cachedPath := filepath.Join(homeDir, ".multifs", "cache")
	cachedFile := "ptoh"

	if _, err := os.Stat(filepath.Join(cachedPath, cachedFile)); err != nil {
		return
	} else {
		f, _ := os.Open(filepath.Join(cachedPath, cachedFile))
		dec := gob.NewDecoder(f)
		if err := dec.Decode(&FilePToH); err != nil {
			log.Warn(err)
		}
	}
}

func saveToCachedHash() {
	homeDir, _ := os.UserHomeDir()
	cachedPath := filepath.Join(homeDir, ".multifs", "cache")
	cachedFile := "ptoh"

	if _, err := os.Stat(filepath.Join(cachedPath, cachedFile)); err != nil {
		err = os.MkdirAll(cachedPath, 0700) // Create your file
		if err != nil {
			log.Warn("Cannot create", cachedPath, err)
			return
		}
	}

	f, err := os.OpenFile(filepath.Join(cachedPath, cachedFile), os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		log.Warn("Cannot create", cachedFile, err)
		return
	}
	log.Info("Opened File", "file", f.Name())
	enc := gob.NewEncoder(f)
	if err := enc.Encode(&FilePToH); err != nil {
		log.Warn(err)
	}
	f.Close()
}

func processHashFiles(fileList *[]string) {
	defer TimeTrack(time.Now(), "processHashFiles")
	populateFromCachedHash()
	var file string
	POOL := 5

	jobs := make(chan string, POOL)
	var wg sync.WaitGroup

	for i := 0; i < POOL; i++ {
		go hashWorker(jobs, &wg)
	}

	for len(*fileList) > 0 {
		file, *fileList = (*fileList)[len(*fileList)-1], (*fileList)[:len(*fileList)-1]
		wg.Add(1)
		jobs <- file
	}
	wg.Wait()
	close(jobs)
	saveToCachedHash()
}

func hashWorker(jobs <-chan string, wg *sync.WaitGroup) {
	for j := range jobs {
		hashFileAndStoreMap(j)
		wg.Done()
	}
}

func GetFilePath(p string) string {
	FilePToH.mutex.RLock()
	hash := FilePToH.StrMap[p]
	FilePToH.mutex.RUnlock()

	FileHToP.RLock()
	paths, exists := FileHToP.Get(hash)
	FileHToP.RUnlock()

	if !exists {
		return p
	}

	return paths[rand.Intn(len(paths))].(string)
}

func hashFileAndStoreMap(newPath string) {
	defer TimeTrack(time.Now(), "Hashing "+newPath)
	hash, exists := FilePToH.StrMap[newPath]

	if !exists {
		hash, err := getFileSHA256(newPath)
		if err != nil {
			return
		}

		FilePToH.mutex.Lock()
		FilePToH.StrMap[newPath] = hash
		FilePToH.mutex.Unlock()
	}

	FileHToP.Lock()
	FileHToP.Put(hash, newPath)
	FileHToP.Unlock()
}

func getFileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	defer f.Close()

	if err != nil {
		return "", err
	}

	h := xxHash.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Error("Error hashing file", p, err)
	}
	return fmt.Sprintf("%x", (h.Sum(nil))), nil
}
