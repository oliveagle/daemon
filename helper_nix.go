// +build linux darwin

package daemon

import (
	"os"
	"os/exec"
//	"os/user"
//    "fmt"
    "strconv"
)

// Service constants
const (
	rootPrivileges = "You must have root user privileges. Possibly using 'sudo' command should help"
	success        = "\t\t\t\t\t[  \033[32mOK\033[0m  ]" // Show colored "OK"
	failed         = "\t\t\t\t\t[\033[31mFAILED\033[0m]" // Show colored "FAILED"
)

func IsExecutable(path string) (bool, error) {
	in, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer in.Close()

	stat, err := in.Stat()
	if err != nil {
		return false, err
	}

	if stat.Mode()&0111 != 0 {
		// executable
		return true, nil
	} else {
		return false, nil
	}
}

// Lookup path for executable file
func executablePath(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return execPath()
		}
		return path, nil
	}
	return execPath()
}

// Check root rights to use system service
func checkPrivileges() bool {
    // on linux, user depends on cgo. if cgo disabled. user.Current will not work.
//    usr, err := user.Current()
//    fmt.Printf("checkPrivileges: err: %v, user: %+v\n", err, usr)
//	if usr, err := user.Current(); err == nil && usr.Gid == "0" {
//		return true
//	}
    if gid, err := get_gid(); err == nil && gid == 0 {
        return true
    }
	return false
}

func get_gid() (int, error) {
    gid := -1
    err := ReadCommand(func(line string) error{
        if v, err := strconv.Atoi(line); err == nil {
            gid = v
            return nil
        } else {
            return err
        }
    }, "id", "-g")
    return gid, err
}
