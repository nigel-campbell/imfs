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
	panic("implement me!")
}

func (s *Shell) Cd(name string) {}

func (s *Shell) Pwd() string {
	panic("implement me!")
}

func (s *Shell) Move(source, dest string) {
	panic("implement me!")
}

func (s *Shell) Copy(source, dest string) {
	panic("implement me!")
}

func (s *Shell) Mkdir(name string) {
	panic("implement me!")
}

func (s *Shell) Touch(name string) {
	panic("implement me!")
}

func (s *Shell) Cat(name string) string {
	panic("implement me!")
}

func (s *Shell) Find(name string) string {
	panic("implement me!")
}

func (s *Shell) Remove(name string, recursive bool) {
	panic("implement me!")
}

func (s *Shell) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s> ", "/")
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
