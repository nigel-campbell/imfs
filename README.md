# IMFS - In-Memory File System

`imfs` is a simple shell wrapping an in-memory file system implementation written in Go. It provides a CLI that mimics common Unix-like file system operations, allowing users to interact with a virtual file system entirely in memory.

## Features

- **File System Operations**
  - Create, read, update, and delete files and directories
  - Navigate through directories using `cd`
  - List directory contents with `ls`
  - Show current working directory with `pwd`
  - Create directories with `mkdir` (including parent directories with `-p` flag)
  - Create empty files with `touch`
  - Move files/directories with `mv`
  - Copy files/directories with `cp`
  - Remove files/directories with `rm` (with recursive option `-r`)
  - Search for files with `find`
  - View file contents with `cat`
  - Write/append content to files with `write` and `append`

- **Path Support**
  - Absolute paths (starting with `/`)
  - Relative paths
  - Parent directory navigation (`..`)

- **File System Features**
  - File metadata tracking (creation time, modification time, size)
  - Directory hierarchy support
  - In-memory storage for files and directories

## Usage

To start the IMFS shell, run:

```bash
go run imfs.go
```

### Available Commands

- `ls` - List directory contents
- `cd <path>` - Change directory
- `pwd` - Print working directory
- `mkdir [-p] <path>` - Create directory (use -p to create parent directories)
- `touch <file>` - Create empty file
- `cat <file>` - Display file contents
- `mv <source> <destination>` - Move file/directory
- `cp <source> <destination>` - Copy file/directory
- `rm [-r] <path>` - Remove file/directory (use -r for recursive removal)
- `find <pattern>` - Search for files
- `write <file> <content>` - Write content to file
- `append <file> <content>` - Append content to file
- `clear` - Clear the screen
- `exit` - Exit the shell

### Examples

```bash
# Create a directory structure
mkdir -p /home/user/documents

# Create and write to a file
write /home/user/documents/note.txt "Hello, World!"

# Append content to a file
append /home/user/documents/note.txt "This is a new line"

# View file contents
cat /home/user/documents/note.txt

# Move a file
mv /home/user/documents/note.txt /home/user/note.txt

# Copy a directory
cp -r /home/user/documents /home/user/backup

# Find files
find note.txt
```

## Implementation Details

IMFS is implemented as a simple in-memory file system using Go's standard library. The system maintains a tree structure of files and directories, with each node containing metadata such as:

- Name
- Size
- Creation time
- Modification time
- Content (for files)
- Children (for directories)
- Parent reference

## Limitations

- The file system is in-memory only and does not persist data between sessions
- No file permissions or ownership system
- No symbolic links or hard links
- No file locking mechanism
- Limited error handling for edge cases

## Future Improvements

- Add data persistence
- Implement file permissions
- Add support for symbolic and hard links
- Improve error handling
- Add file locking mechanism
- Implement file compression
- Add support for file attributes

## License

This project is open source and available under the MIT License. 