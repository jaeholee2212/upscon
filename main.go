package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("unsupported")
	}
}

func run() {
	fmt.Printf("Running %v as %v\n", os.Args[2:], os.Getpid())

	params := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", params...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Child: running %v as %v\n", os.Args[2:], os.Getpid())

	cgroups()

	must(syscall.Sethostname([]byte("upscon-host")))
	must(syscall.Chroot("/root/netshoot"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	must(syscall.Mount("tmpfs", "/run", "tmpfs", 0, ""))
	must(syscall.Mount("tmpfs", "/dev", "tmpfs", 0, ""))
	must(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""))
	must(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""))
	defer syscall.Unmount("/proc", 0)
	defer syscall.Unmount("/run", 0)
	defer syscall.Unmount("/dev", 0)
	defer syscall.Unmount("/tmp", 0)
	defer syscall.Unmount("/sys", 0)

	os.Setenv("PATH", "$PATH:/bin:/usr/local/bin:/usr/bin:")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
}

func cgroups() {
	fmt.Println("Cgroups")
	const cgroups = "/sys/fs/cgroup/"

	upscon := filepath.Join(cgroups, "pids", "upscon")
	must(os.Mkdir(upscon, 0755))

	pidsMaxPath := filepath.Join(upscon, "pids.max")
	must(ioutil.WriteFile(pidsMaxPath, []byte("20"), 0700))

	procsPath := filepath.Join(upscon, "cgroup.procs")
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)
	ioutil.WriteFile(procsPath, []byte(pidStr), 0700)
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
