package filesystem

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	fs := New()

	root := &file{
		name:     "/",
		path:     "/",
		isDir:    true,
		children: make(Children),
		parent:   nil,
	}

	expected := Fs{
		root:       root,
		currentDir: root,
	}

	if !reflect.DeepEqual(fs, expected) {
		t.Errorf("filesystem.New() = %+v, want %+v", fs, expected)
	}
}

func TestPrintWorkingDirectory(t *testing.T) {
	fs := New()

	pwd := fs.PrintWorkingDirectory()

	if pwd != fs.currentDir.path {
		t.Errorf("filesystem.PrintWorkingDirectory() = %+v, want %+v", pwd, fs.currentDir.path)
	}
}

func TestCreateDir(t *testing.T) {
	fs := New()

	// Successful creation of directories
	succCreate := func(name string) func(t *testing.T) {
		return func(t *testing.T) {
			err := fs.CreateDir(name)

			if err != nil {
				t.Errorf("filesystem.CreateDir(%s) returned an error %s", name, err)
			}

			err = fs.ChangeDir(name)

			if err != nil {
				t.Errorf("filesystem.CreateDir(%s) did not create a dir %s; error: %s", name, name, err)
			}

			fs.ChangeDir("/")
		}
	}

	t.Run("usr", succCreate("usr"))
	t.Run("etc/", succCreate("etc/"))
	t.Run("/opt/", succCreate("/opt/"))
	t.Run("/home", succCreate("/home"))
	t.Run("/usr/share", succCreate("/usr/share"))
	t.Run("opt/local", succCreate("opt/local"))

	// Failed creation of directories
	failCreate := func(name string, expectedErr error, duplicate bool) func(t *testing.T) {
		return func(t *testing.T) {
			err := fs.CreateDir(name)

			if err == nil {
				t.Errorf("filesystem.CreateDir(%s) didn't return an error; expected error %s", name, expectedErr)
			} else if err != expectedErr {
				t.Errorf("filesystem.CreateDir(%s) returned %s; expected error %s", name, err, expectedErr)
			}

			if !duplicate {
				err = fs.ChangeDir(name)

				if err == nil {
					t.Errorf("filesystem.CreateDir(%s) did create a dir %s, when it shouldn't have; expected error: %s", name, name, expectedErr)
				}

				fs.ChangeDir("/")
			}
		}
	}

	t.Run("usr/share/this/is-bad", failCreate("usr/share/this/is-bad", ErrWalkFail, false))
	t.Run("usr/share/", failCreate("usr/share/", ErrDuplicateDir, true))
	t.Run("/usr/share/this/is-bad", failCreate("usr/share/this/is-bad", ErrWalkFail, false))
	t.Run("/usr/share/", failCreate("usr/share/", ErrDuplicateDir, true))
}

func TestChangeDir(t *testing.T) {
	fs := New()

	fs.CreateDir("usr")

	succChange := func(name string) func(t *testing.T) {
		return func(t *testing.T) {
			err := fs.ChangeDir(name)

			if err != nil {
				t.Errorf("filesystem.ChangeDir(%s) returned an error %s", name, err)
			}

			fs.ChangeDir("/")
		}
	}

	t.Run("usr", succChange("usr"))
	t.Run("usr/", succChange("usr/"))
	t.Run("/usr/", succChange("/usr/"))
	t.Run("/usr", succChange("/usr"))

	// Nested cd
	fs.CreateDir("/usr/share")

	t.Run("/usr/share", succChange("/usr/share"))
	t.Run("usr/share", succChange("usr/share"))
}

func TestWalk(t *testing.T) {
	fs := New()

	fs.CreateDir("usr")
	fs.CreateDir("usr/share")

	succWalk := func(name string, node *file) func(t *testing.T) {
		return func(t *testing.T) {
			walk, err := fs.currentDir.walk(name)

			if err != nil {
				t.Errorf("f.walk(%s) returned an error %s", name, err)
			}

			if !reflect.DeepEqual(walk, node) {
				t.Errorf("f.Walk(%s) didn't return the correct node at %s", name, name)
			}
		}
	}

	t.Run("usr/share from /", succWalk("usr/share", fs.currentDir.children["usr"].children["share"]))
	t.Run("usr from /", succWalk("usr", fs.currentDir.children["usr"]))

	fs.ChangeDir("usr")

	t.Run("../ from usr", succWalk("../", fs.currentDir.parent))

	fs.ChangeDir("share")

	t.Run("../../ from /usr/share", succWalk("../../", fs.currentDir.parent.parent))

	t.Run("/usr from /usr/share", succWalk("/usr", fs.currentDir.parent))
}
