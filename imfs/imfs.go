package imfs

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type File struct {
	Name        string
	Size        int64
	CreatedAt   time.Time
	ModifiedAt  time.Time
	IsDirectory bool
	Content     []byte
	Children    []*File
	Parent      *File
}

// Shell is a simple REPL for interacting with the file system
// NB: This system should... mostly be provably correct.
// TODO(nigel): Add unit tests to cover state transitions.
type Shell struct {
	Root *File
	Cwd  *File
}

func NewShell() *Shell {
	root := &File{
		Name:        "/",
		IsDirectory: true,
	}
	return &Shell{
		Root: root,
		Cwd:  root,
	}
}

func (s *Shell) Ls() []string {
	if s.Cwd == nil {
		return nil
	}

	names := make([]string, 0, len(s.Cwd.Children))

	for _, child := range s.Cwd.Children {
		names = append(names, child.Name)
		if child.IsDirectory {
			fmt.Printf("%s/\n", child.Name)
		} else {
			fmt.Println(child.Name)
		}
	}

	return names
}

func (s *Shell) Cd(name string) {
	if name == "" {
		return
	}

	switch name {
	case "/":
		s.Cwd = s.Root
		return
	case "..":
		if s.Cwd.Parent != nil {
			s.Cwd = s.Cwd.Parent
		}
		return
	}

	if strings.HasPrefix(name, "/") {
		s.Cwd = s.Root
		name = name[1:]
		if name == "" {
			return
		}
	}

	components := strings.Split(name, "/")
	currentDir := s.Cwd

	for _, component := range components {
		if component == "" {
			continue
		}

		if component == ".." {
			if currentDir.Parent != nil {
				currentDir = currentDir.Parent
			}
			continue
		}

		found := false
		for _, child := range currentDir.Children {
			if child.Name == component && child.IsDirectory {
				currentDir = child
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("cd: no such directory: %s\n", name)
			return
		}
	}

	s.Cwd = currentDir
}

func (s *Shell) Pwd() string {
	if s.Cwd == s.Root {
		fmt.Println("/")
		return "/"
	}

	path := s.Cwd.Name
	current := s.Cwd.Parent
	for current != nil && current != s.Root {
		path = current.Name + "/" + path
		current = current.Parent
	}
	path = "/" + path
	fmt.Println(path)
	return path
}

func (s *Shell) RedirectWrite(filename, content string, shouldAppend bool) {
	if filename == "" {
		return
	}

	var targetFile *File
	for _, child := range s.Cwd.Children {
		if child.Name == filename {
			if child.IsDirectory {
				return // Can't write to a directory
			}
			targetFile = child
			break
		}
	}

	if targetFile == nil {
		// Check if there's a directory with this name before creating a new file
		for _, child := range s.Cwd.Children {
			if child.Name == filename && child.IsDirectory {
				return // Can't create a file with the same name as a directory
			}
		}

		targetFile = &File{
			Name:        filename,
			IsDirectory: false,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Parent:      s.Cwd,
		}
		s.Cwd.Children = append(s.Cwd.Children, targetFile)
	}

	if shouldAppend {
		targetFile.Content = append(targetFile.Content, []byte(content)...)
	} else {
		targetFile.Content = []byte(content)
	}

	targetFile.Size = int64(len(targetFile.Content))
	targetFile.ModifiedAt = time.Now()
}

func (s *Shell) Move(source, dest string) {
	panic("implement me!")
}

func (s *Shell) Copy(source, dest string) {
	panic("implement me!")
}

func (s *Shell) Mkdir(name string) {
	if name == "" {
		fmt.Println("Usage: mkdir <directory_name>")
		return
	}

	for _, child := range s.Cwd.Children {
		if child.Name == name {
			return
		}
	}

	newDir := &File{
		Name:        name,
		IsDirectory: true,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Parent:      s.Cwd,
	}

	s.Cwd.Children = append(s.Cwd.Children, newDir)
}

func (s *Shell) Touch(name string) {
	if name == "" {
		fmt.Println("Usage: touch <file_name>")
		return
	}

	// TODO(nigel): Replace with a map.
	for _, child := range s.Cwd.Children {
		if child.Name == name {
			// Update modification time if file exists
			// TODO(nigel): Is this necessary?
			child.ModifiedAt = time.Now()
			return
		}
	}

	newFile := &File{
		Name:        name,
		IsDirectory: false,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Parent:      s.Cwd,
	}
	s.Cwd.Children = append(s.Cwd.Children, newFile)
}

func (s *Shell) Cat(name string) string {
	if name == "" {
		return ""
	}

	// Handle parent directory navigation
	if strings.HasPrefix(name, "../") {
		if s.Cwd.Parent == nil {
			return ""
		}
		// Save current directory
		current := s.Cwd
		// Move to parent
		s.Cwd = s.Cwd.Parent
		// Get content
		content := s.Cat(strings.TrimPrefix(name, "../"))
		// Restore current directory
		s.Cwd = current
		return content
	}

	// Find the file
	for _, child := range s.Cwd.Children {
		if child.Name == name {
			if child.IsDirectory {
				return ""
			}
			return string(child.Content)
		}
	}

	return ""
}

func (s *Shell) Find(name string) string {
	if name == "" {
		return ""
	}

	var searchDir func(dir *File, currentPath string) string
	searchDir = func(dir *File, currentPath string) string {
		for _, child := range dir.Children {
			childPath := currentPath + "/" + child.Name
			if strings.Contains(child.Name, name) {
				return childPath
			}
			if child.IsDirectory {
				if result := searchDir(child, childPath); result != "" {
					return result
				}
			}
		}
		return ""
	}

	return searchDir(s.Root, "")
}

func (s *Shell) Remove(name string, recursive bool) {
	if name == "" {
		fmt.Println("Usage: remove <name> [recursive]")
		return
	}

	// Find the file/directory to remove
	var target *File
	var targetIndex int
	for i, child := range s.Cwd.Children {
		if child.Name == name {
			target = child
			targetIndex = i
			break
		}
	}

	if target == nil {
		return // File/directory doesn't exist
	}

	// Check if trying to remove a non-empty directory without recursive flag
	if target.IsDirectory && len(target.Children) > 0 && !recursive {
		fmt.Printf("Cannot remove '%s': Directory not empty\n", name)
		return
	}

	// Remove from parent's children
	s.Cwd.Children = append(s.Cwd.Children[:targetIndex], s.Cwd.Children[targetIndex+1:]...)
}

func (s *Shell) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s> ", s.Pwd())
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(input, " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		switch cmd {
		case "exit":
			return
		case "ls":
			s.Ls()
		case "cd":
			s.Cd(arg)
		case "pwd":
			s.Pwd()
		case "mkdir":
			s.Mkdir(arg)
		case "touch":
			s.Touch(arg)
		case "cat":
			s.Cat(arg)
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}
