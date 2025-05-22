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

		// TODO(nigel): Why is this necessary? Can this be removed?
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
		return "/"
	}

	path := s.Cwd.Name
	current := s.Cwd.Parent
	for current != nil && current != s.Root {
		path = current.Name + "/" + path
		current = current.Parent
	}
	path = "/" + path
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
		for _, child := range s.Cwd.Children {
			if child.Name == filename && child.IsDirectory {
				return
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
	if source == "" || dest == "" {
		fmt.Println("Usage: mv <source> <destination>")
		return
	}

	// Save current directory
	originalCwd := s.Cwd

	// Find source file/directory
	var sourceFile *File
	var sourceIndex int
	for i, child := range s.Cwd.Children {
		if child.Name == source {
			sourceFile = child
			sourceIndex = i
			break
		}
	}

	if sourceFile == nil {
		return // Source doesn't exist
	}

	// Handle absolute paths
	if strings.HasPrefix(dest, "/") {
		s.Cwd = s.Root
		dest = dest[1:]
		if dest == "" {
			return
		}
	}

	// Remove trailing slash if present
	dest = strings.TrimSuffix(dest, "/")

	// Split destination path into components
	components := strings.Split(dest, "/")
	currentDir := s.Cwd

	// Navigate to the target directory, but DO NOT create directories
	for i := 0; i < len(components); i++ {
		component := components[i]
		if component == "" {
			continue
		}

		if component == ".." {
			if currentDir.Parent != nil {
				currentDir = currentDir.Parent
			}
			continue
		}

		// If this is the last component and it's not "..", check if it's a directory
		if i == len(components)-1 {
			var destDir *File
			for _, child := range currentDir.Children {
				if child.Name == component && child.IsDirectory {
					destDir = child
					break
				}
			}

			if destDir != nil {
				// Move into this directory, keep original name
				for _, c := range destDir.Children {
					if c.Name == sourceFile.Name {
						s.Cwd = originalCwd
						return // Target already exists
					}
				}
				// Remove source from its current location
				originalCwd.Children = append(originalCwd.Children[:sourceIndex], originalCwd.Children[sourceIndex+1:]...)
				sourceFile.Parent = destDir
				destDir.Children = append(destDir.Children, sourceFile)
				s.Cwd = originalCwd
				return
			}

			// Check if target already exists
			for _, child := range currentDir.Children {
				if child.Name == component {
					s.Cwd = originalCwd
					return // Target already exists
				}
			}

			// If we're moving to the same location with the same name, do nothing
			if currentDir == originalCwd && component == sourceFile.Name {
				s.Cwd = originalCwd
				return
			}

			// Remove source from its current location
			originalCwd.Children = append(originalCwd.Children[:sourceIndex], originalCwd.Children[sourceIndex+1:]...)

			// Update source's parent and name
			sourceFile.Parent = currentDir
			sourceFile.Name = component

			// Add source to new location
			currentDir.Children = append(currentDir.Children, sourceFile)

			// Restore original directory
			s.Cwd = originalCwd
			return
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
			s.Cwd = originalCwd
			return // Destination directory doesn't exist, abort move
		}
	}

	// If we get here, we're moving to the current directory
	// Check if target already exists
	for _, child := range currentDir.Children {
		if child.Name == sourceFile.Name {
			s.Cwd = originalCwd
			return // Target already exists
		}
	}

	// Remove source from its current location
	originalCwd.Children = append(originalCwd.Children[:sourceIndex], originalCwd.Children[sourceIndex+1:]...)

	// Update source's parent
	sourceFile.Parent = currentDir

	// Add source to new location
	currentDir.Children = append(currentDir.Children, sourceFile)

	// Restore original directory
	s.Cwd = originalCwd
}

func (s *Shell) Copy(source, dest string) {
	if source == "" || dest == "" {
		return
	}

	// Save current directory
	originalCwd := s.Cwd

	// Find source file/directory
	var sourceFile *File
	for _, child := range s.Cwd.Children {
		if child.Name == source {
			sourceFile = child
			break
		}
	}

	if sourceFile == nil {
		return // Source doesn't exist
	}

	// Handle absolute paths
	if strings.HasPrefix(dest, "/") {
		s.Cwd = s.Root
		dest = dest[1:]
	}

	// Split destination path into components
	components := strings.Split(dest, "/")
	lastComponent := components[len(components)-1]
	components = components[:len(components)-1]

	// Navigate to destination directory
	for _, component := range components {
		if component == "" {
			continue
		}

		if component == ".." {
			if s.Cwd.Parent != nil {
				s.Cwd = s.Cwd.Parent
			}
			continue
		}

		found := false
		for _, child := range s.Cwd.Children {
			if child.Name == component && child.IsDirectory {
				s.Cwd = child
				found = true
				break
			}
		}

		if !found {
			// Destination directory doesn't exist
			s.Cwd = originalCwd
			return
		}
	}

	// Create a deep copy of the source file/directory
	var copyFile func(*File) *File
	copyFile = func(f *File) *File {
		newFile := &File{
			Name:        f.Name,
			Size:        f.Size,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			IsDirectory: f.IsDirectory,
			Content:     make([]byte, len(f.Content)),
			Parent:      s.Cwd,
		}
		copy(newFile.Content, f.Content)

		if f.IsDirectory {
			newFile.Children = make([]*File, 0, len(f.Children))
			for _, child := range f.Children {
				newChild := copyFile(child)
				newChild.Parent = newFile
				newFile.Children = append(newFile.Children, newChild)
			}
		}

		return newFile
	}

	// Create the copy with the new name
	newFile := copyFile(sourceFile)
	newFile.Name = lastComponent

	// Add the copy to the destination directory
	s.Cwd.Children = append(s.Cwd.Children, newFile)

	// Restore original working directory
	s.Cwd = originalCwd
}

func (s *Shell) Mkdir(name string, createParents bool) {
	if name == "" {
		fmt.Println("Usage: mkdir [-p] <directory_name>")
		return
	}

	// Handle absolute paths
	if strings.HasPrefix(name, "/") {
		s.Cwd = s.Root
		name = name[1:]
		if name == "" {
			return
		}
	}

	components := strings.Split(name, "/")
	currentDir := s.Cwd

	// Navigate to the target directory
	for i := 0; i < len(components)-1; i++ {
		component := components[i]
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
			if !createParents {
				fmt.Printf("mkdir: no such directory: %s\n", name)
				return
			}
			// Create the parent directory
			newDir := &File{
				Name:        component,
				IsDirectory: true,
				CreatedAt:   time.Now(),
				ModifiedAt:  time.Now(),
				Parent:      currentDir,
			}
			currentDir.Children = append(currentDir.Children, newDir)
			currentDir = newDir
		}
	}

	// Get the final component (directory name)
	dirname := components[len(components)-1]
	if dirname == "" {
		return
	}

	// Check if directory already exists
	for _, child := range currentDir.Children {
		if child.Name == dirname {
			return
		}
	}

	// Create new directory
	newDir := &File{
		Name:        dirname,
		IsDirectory: true,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Parent:      currentDir,
	}
	currentDir.Children = append(currentDir.Children, newDir)
}

func (s *Shell) Touch(name string) {
	if name == "" {
		fmt.Println("Usage: touch <file_name>")
		return
	}

	// Handle absolute paths
	if strings.HasPrefix(name, "/") {
		s.Cwd = s.Root
		name = name[1:]
		if name == "" {
			return
		}
	}

	components := strings.Split(name, "/")
	currentDir := s.Cwd

	// Navigate to the target directory
	for i := 0; i < len(components)-1; i++ {
		component := components[i]
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
			fmt.Printf("touch: no such directory: %s\n", name)
			return
		}
	}

	// Get the final component (file name)
	filename := components[len(components)-1]
	if filename == "" {
		return
	}

	// Check if file already exists
	for _, child := range currentDir.Children {
		if child.Name == filename {
			// Update modification time if file exists
			child.ModifiedAt = time.Now()
			return
		}
	}

	// Create new file
	newFile := &File{
		Name:        filename,
		IsDirectory: false,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Parent:      currentDir,
	}
	currentDir.Children = append(currentDir.Children, newFile)
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
		return
	}

	if target.IsDirectory && len(target.Children) > 0 && !recursive {
		fmt.Printf("Cannot remove '%s': Directory not empty\n", name)
		return
	}

	s.Cwd.Children = append(s.Cwd.Children[:targetIndex], s.Cwd.Children[targetIndex+1:]...)
}

func (s *Shell) Clear() {
	// ANSI escape sequence to clear the screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")
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
			fmt.Println(s.Pwd())
		case "mkdir":
			createParents := false
			if strings.HasPrefix(arg, "-p ") {
				createParents = true
				arg = strings.TrimPrefix(arg, "-p ")
			}
			s.Mkdir(arg, createParents)
		case "touch":
			s.Touch(arg)
		case "cat":
			fmt.Println(s.Cat(arg))
		case "clear":
			s.Clear()
		case "mv":
			parts := strings.SplitN(arg, " ", 2)
			if len(parts) == 2 {
				s.Move(parts[0], parts[1])
			} else {
				fmt.Println("Usage: move <source> <destination>")
			}
		case "cp":
			parts := strings.SplitN(arg, " ", 2)
			if len(parts) == 2 {
				s.Copy(parts[0], parts[1])
			} else {
				fmt.Println("Usage: copy <source> <destination>")
			}
		case "find":
			if result := s.Find(arg); result != "" {
				fmt.Println(result)
			}
		case "rm":
			parts := strings.SplitN(arg, " ", 2)
			recursive := false
			if len(parts) == 2 && parts[1] == "-r" {
				recursive = true
				arg = parts[0]
			}
			s.Remove(arg, recursive)
		case "write":
			parts := strings.SplitN(arg, " ", 2)
			if len(parts) == 2 {
				s.RedirectWrite(parts[0], parts[1], false)
			} else {
				fmt.Println("Usage: write <file> <content>")
			}
		case "append":
			parts := strings.SplitN(arg, " ", 2)
			if len(parts) == 2 {
				s.RedirectWrite(parts[0], parts[1], true)
			} else {
				fmt.Println("Usage: append <file> <content>")
			}
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}
