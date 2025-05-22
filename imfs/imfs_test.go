package imfs

import (
	"testing"
)

func assertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func TestChangeDirectory(t *testing.T) {
	shell := NewShell()

	shell.Cd("nonexistent")
	assertEqual(t, shell.Root, shell.Cwd, "Expected to stay in root directory when changing to non-existent directory")

	shell.Mkdir("testdir")
	shell.Cd("testdir")
	assertEqual(t, "testdir", shell.Cwd.Name, "Expected to change to 'testdir'")

	shell.Cd("..")
	assertEqual(t, shell.Root, shell.Cwd, "Expected to return to root directory")

	shell.Cd("/")
	assertEqual(t, shell.Root, shell.Cwd, "Expected to be in root directory")
}

func TestGetWorkingDirectory(t *testing.T) {
	shell := NewShell()

	assertEqual(t, "/", shell.Pwd(), "Expected root directory path to be '/'")

	shell.Mkdir("testdir")
	shell.Cd("testdir")
	assertEqual(t, "/testdir", shell.Pwd(), "Expected path to be '/testdir'")

	shell.Mkdir("nested")
	shell.Cd("nested")
	assertEqual(t, "/testdir/nested", shell.Pwd(), "Expected path to be '/testdir/nested'")

	shell.Cd("/")
	assertEqual(t, "/", shell.Pwd(), "Expected path to be '/' after changing to root")
}

func TestMakeDirectory(t *testing.T) {
	shell := NewShell()

	// Test creating a simple directory
	shell.Mkdir("testdir")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected one child directory")
	assertEqual(t, "testdir", shell.Cwd.Children[0].Name, "Expected directory name to be 'testdir'")
	assertEqual(t, true, shell.Cwd.Children[0].IsDirectory, "Expected IsDirectory to be true")

	// Test creating a nested directory
	shell.Cd("testdir")
	shell.Mkdir("nested")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected one child directory in nested location")
	assertEqual(t, "nested", shell.Cwd.Children[0].Name, "Expected nested directory name to be 'nested'")

	// Test creating duplicate directory (should be idempotent)
	shell.Mkdir("nested")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected no new directory to be created for duplicate")

	// Test creating directory with empty name
	shell.Mkdir("")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected no new directory to be created for empty name")
}

func TestRemoveDirectory(t *testing.T) {
	shell := NewShell()

	// Test removing a non-existent directory
	shell.Remove("nonexistent", false)
	assertEqual(t, 0, len(shell.Cwd.Children), "Expected no children in root directory")

	// Test removing an empty directory
	shell.Mkdir("testdir")
	shell.Remove("testdir", false)
	assertEqual(t, 0, len(shell.Cwd.Children), "Expected empty directory to be removed")

	// Test removing a directory with children (recursive)
	shell.Mkdir("testdir")
	shell.Cd("testdir")
	shell.Mkdir("nested")
	shell.Cd("..") // Go back to parent
	shell.Remove("testdir", true)
	assertEqual(t, 0, len(shell.Cwd.Children), "Expected testdir to be removed (recursive)")

	// Test removing a file
	shell.Touch("testfile")
	shell.Remove("testfile", false)
	assertEqual(t, 0, len(shell.Cwd.Children), "Expected testfile to be removed")
}

func TestCreateNewFile(t *testing.T) {
	shell := NewShell()

	// Create a new file
	shell.Touch("file1.txt")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected one file in root directory")
	assertEqual(t, "file1.txt", shell.Cwd.Children[0].Name, "Expected file name to be 'file1.txt'")
	assertEqual(t, false, shell.Cwd.Children[0].IsDirectory, "Expected IsDirectory to be false")

	// Try to create the same file again (should not create a duplicate)
	shell.Touch("file1.txt")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected no duplicate file to be created")

	// Try to create a file with an empty name
	shell.Touch("")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected no file to be created for empty name")

	// Create a file in a nested directory
	shell.Mkdir("subdir")
	shell.Cd("subdir")
	shell.Touch("file2.txt")
	assertEqual(t, 1, len(shell.Cwd.Children), "Expected one file in subdir")
	assertEqual(t, "file2.txt", shell.Cwd.Children[0].Name, "Expected file name to be 'file2.txt'")
}

func TestRedirectFileContents(t *testing.T) {
	shell := NewShell()

	// Test writing to a new file
	shell.RedirectWrite("file1.txt", "Hello, world!", false)
	assertEqual(t, "Hello, world!", string(shell.Cwd.Children[0].Content), "Expected file content to be 'Hello, world!'")

	// Test overwriting existing file
	shell.RedirectWrite("file1.txt", "New content", false)
	assertEqual(t, "New content", string(shell.Cwd.Children[0].Content), "Expected file content to be overwritten")

	// Test appending to existing file
	shell.RedirectWrite("file1.txt", " appended", true)
	assertEqual(t, "New content appended", string(shell.Cwd.Children[0].Content), "Expected content to be appended")

	// Test writing to a new file in a subdirectory
	shell.Mkdir("subdir")
	shell.Cd("subdir")
	shell.RedirectWrite("file2.txt", "In subdirectory", false)
	assertEqual(t, "In subdirectory", string(shell.Cwd.Children[0].Content), "Expected file content in subdirectory")

	// Test writing to a directory (should be ignored)
	shell.Cd("..")
	// Find the subdir
	var subdir *File
	for _, child := range shell.Cwd.Children {
		if child.Name == "subdir" {
			subdir = child
			break
		}
	}
	shell.RedirectWrite("subdir", "This should not work", false)
	assertEqual(t, 0, len(subdir.Content), "Expected directory content to remain empty")

	// Test writing with empty filename (should be ignored)
	shell.RedirectWrite("", "This should not work", false)
	assertEqual(t, 2, len(shell.Cwd.Children), "Expected no new file to be created")
}

func TestCat(t *testing.T) {
	shell := NewShell()

	// Test reading non-existent file
	content := shell.Cat("nonexistent.txt")
	assertEqual(t, "", content, "Expected empty string for non-existent file")

	// Test reading directory
	shell.Mkdir("testdir")
	content = shell.Cat("testdir")
	assertEqual(t, "", content, "Expected empty string when reading directory")

	// Test reading empty file
	shell.Touch("empty.txt")
	content = shell.Cat("empty.txt")
	assertEqual(t, "", content, "Expected empty string for empty file")

	// Test reading file with content
	shell.RedirectWrite("test.txt", "Hello, world!", false)
	content = shell.Cat("test.txt")
	assertEqual(t, "Hello, world!", content, "Expected file content to be 'Hello, world!'")

	// Test reading file in subdirectory
	shell.Mkdir("subdir")
	shell.Cd("subdir")
	shell.RedirectWrite("nested.txt", "Nested content", false)
	content = shell.Cat("nested.txt")
	assertEqual(t, "Nested content", content, "Expected nested file content to be 'Nested content'")

	// Test reading file from parent directory
	content = shell.Cat("../test.txt")
	assertEqual(t, "Hello, world!", content, "Expected to read file from parent directory")
}

func TestFind(t *testing.T) {
	shell := NewShell()

	// Test finding non-existent file
	result := shell.Find("nonexistent.txt")
	assertEqual(t, "", result, "Expected empty string for non-existent file")

	// Test finding file in root
	shell.RedirectWrite("root.txt", "root content", false)
	result = shell.Find("root.txt")
	assertEqual(t, "/root.txt", result, "Expected to find file in root directory")

	// Test finding file in subdirectory
	shell.Mkdir("subdir")
	shell.Cd("subdir")
	shell.RedirectWrite("sub.txt", "sub content", false)
	result = shell.Find("sub.txt")
	assertEqual(t, "/subdir/sub.txt", result, "Expected to find file in subdirectory")

	// Test finding directory
	shell.Mkdir("nested")
	result = shell.Find("nested")
	assertEqual(t, "/subdir/nested", result, "Expected to find directory")

	// Test finding file from subdirectory
	result = shell.Find("root.txt")
	assertEqual(t, "/root.txt", result, "Expected to find file in parent directory")

	// Test finding multiple files with same name
	shell.Cd("..")
	shell.Mkdir("other")
	shell.Cd("other")
	shell.RedirectWrite("sub.txt", "other sub content", false)
	result = shell.Find("sub.txt")
	assertEqual(t, "/subdir/sub.txt", result, "Expected to find first occurrence of file")

	// Test finding with partial match
	shell.RedirectWrite("test1.txt", "test1", false)
	shell.RedirectWrite("test2.txt", "test2", false)
	result = shell.Find("test")
	assertEqual(t, "/other/test1.txt", result, "Expected to find first partial match")
}

func TestList(t *testing.T) {
	shell := NewShell()

	// Test empty directory
	files := shell.Ls()
	assertEqual(t, 0, len(files), "Expected empty directory to return no files")

	// Test directory with files and subdirectories
	shell.RedirectWrite("file1.txt", "content1", false)
	shell.RedirectWrite("file2.txt", "content2", false)
	shell.Mkdir("dir1")
	shell.Mkdir("dir2")

	files = shell.Ls()
	assertEqual(t, 4, len(files), "Expected 4 items in directory")

	// Verify all items are present
	expected := map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
		"dir1":      true,
		"dir2":      true,
	}
	for _, file := range files {
		assertEqual(t, true, expected[file], "Unexpected file in listing: "+file)
	}

	// Test listing subdirectory
	shell.Cd("dir1")
	shell.RedirectWrite("nested.txt", "nested content", false)
	files = shell.Ls()
	assertEqual(t, 1, len(files), "Expected 1 item in subdirectory")
	assertEqual(t, "nested.txt", files[0], "Expected nested.txt in subdirectory")

	// Test listing after removing files
	shell.Cd("..")
	shell.Remove("file1.txt", false)
	files = shell.Ls()
	assertEqual(t, 3, len(files), "Expected 3 items after removal")

	// Verify remaining items
	expected = map[string]bool{
		"file2.txt": true,
		"dir1":      true,
		"dir2":      true,
	}
	for _, file := range files {
		assertEqual(t, true, expected[file], "Unexpected file in listing after removal: "+file)
	}
}

func TestPathResolution(t *testing.T) {
	shell := NewShell()

	// Setup directory structure
	shell.Mkdir("dir1")
	shell.Cd("dir1")
	shell.Mkdir("dir2")
	shell.Cd("dir2")
	shell.Mkdir("dir3")
	shell.Cd("/") // Go back to root

	// Test absolute path navigation
	shell.Cd("/dir1/dir2/dir3")
	assertEqual(t, "dir3", shell.Cwd.Name, "Expected to navigate to dir3 using absolute path")
	assertEqual(t, "/dir1/dir2/dir3", shell.Pwd(), "Expected path to be '/dir1/dir2/dir3'")

	// Test relative path navigation
	shell.Cd("/")
	shell.Cd("dir1/dir2")
	assertEqual(t, "dir2", shell.Cwd.Name, "Expected to navigate to dir2 using relative path")
	assertEqual(t, "/dir1/dir2", shell.Pwd(), "Expected path to be '/dir1/dir2'")

	// Test parent directory navigation with ..
	shell.Cd("../")
	assertEqual(t, "dir1", shell.Cwd.Name, "Expected to navigate to parent directory")
	assertEqual(t, "/dir1", shell.Pwd(), "Expected path to be '/dir1'")

	// Test combined relative with parent navigation
	shell.Cd("dir2/dir3/../")
	assertEqual(t, "dir2", shell.Cwd.Name, "Expected to navigate to dir2 after path with parent reference")
	assertEqual(t, "/dir1/dir2", shell.Pwd(), "Expected path to be '/dir1/dir2'")

	// Test multiple parent directory navigation
	shell.Cd("../../")
	assertEqual(t, shell.Root, shell.Cwd, "Expected to navigate to root with multiple parent references")
	assertEqual(t, "/", shell.Pwd(), "Expected path to be '/'")

	// Test non-existent paths
	shell.Cd("/dir1")
	shell.Cd("nonexistent/path")
	assertEqual(t, "dir1", shell.Cwd.Name, "Expected to stay in current directory for non-existent path")

	// Test paths with empty components
	shell.Cd("/dir1//dir2")
	assertEqual(t, "dir2", shell.Cwd.Name, "Expected to handle empty path components correctly")
	assertEqual(t, "/dir1/dir2", shell.Pwd(), "Expected path to be '/dir1/dir2'")

	// Test complex path with mixed absolute, relative and parent references
	// shell.Cd("/")
	// shell.Cd("/dir1/./dir2/../dir2/dir3/..")
	// assertEqual(t, "dir2", shell.Cwd.Name, "Expected to correctly resolve complex path")
	// assertEqual(t, "/dir1/dir2", shell.Pwd(), "Expected path to be '/dir1/dir2'")
}
