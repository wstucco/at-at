package at_at

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type Host struct {
	name   string
	root   string
	host   string
	port   int
	status HostStatus

	cmdline string
	process *exec.Cmd

	client *http.Client

	Error HostError
}

type HostList map[string]*Host

type HostStatus int
type HostError int

const (
	Stopped HostStatus = iota
	Starting
	Running
	Stopping
	Error
)

const (
	NoError HostError = iota
	Unavailable
	NotFound
)

func NewHost(name string, root string, port int) *Host {
	host := &Host{
		name:   name,
		root:   root,
		host:   "0.0.0.0",
		port:   port,
		client: &http.Client{},
		Error:  NoError,
	}

	return host.validate()
}

func (h *Host) Run() {
	if h.status == Stopped {
		h.status = Starting

		var err error

		if h.process, err = h.createProcess(); err != nil {
			h.status = Error
			Logger().Printf("[%s] cannot create command: %s", h.name, err)
			return
		}

		if err = h.runProcess(); err != nil {
			h.status = Error
			Logger().Printf("[%s] cannot run command: %s", h.name, err)
			return
		}

	}
}

func (h *Host) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	w, err := h.client.Do(h.requestFrom(req))
	if err != nil {
		Logger().Printf("[%s] error sending request: %v", h.name, err)
		return
	}

	defer w.Body.Close()

	contents, err := ioutil.ReadAll(w.Body)
	if err != nil {
		Logger().Printf("[%s] error reading response: %v", h.name, err)
		return
	}

	res.Write(contents)
}

func (h *Host) requestFrom(req *http.Request) *http.Request {
	req.URL.Host = fmt.Sprintf("%s:%d", h.host, h.port)
	req.URL.Scheme = "http"
	req.RequestURI = ""
	return req
}

func (h *Host) SetHost(host string) {
	h.host = host
}

func (h *Host) createProcess() (*exec.Cmd, error) {
	procfile := path.Join(h.root, ".at_at")
	buf, err := ioutil.ReadFile(procfile)
	if err != nil {
		return nil, err
	}

	h.cmdline = string(buf)

	cmd := createCommand(h.cmdline)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", h.port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOST=%s", h.host))
	cmd.Env = append(cmd.Env, fmt.Sprintf("APP_ROOT=%s", h.root))

	cmd.Dir = h.root

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

func (h *Host) runProcess() error {
	Logger().Printf("[%s] starting on host %s, port %d and app_root %s", h.name, h.host, h.port, h.root)
	if err := h.process.Start(); err != nil {
		Logger().Printf("[%s] cannot start process: %s\n", h.name, err)
		h.status = Error
		return err
	}

	done := make(chan error, 1)

	go func() {
		address := fmt.Sprintf("%s:%d", h.host, h.port)
		for {
			if con, err := net.Dial("tcp", address); err == nil {
				con.Close()
				done <- nil
			}
		}
	}()

	select {
	case <-done:
		h.status = Running
		Logger().Printf("[%s] is ready to accept connections", h.name)
	}

	go func() { done <- h.process.Wait() }()

	go func() {
		select {
		case err := <-done:
			if err != nil {
				Logger().Printf("[%s] process died with error: %v", h.name, err)
			} else {
				Logger().Printf("[%s] exited gracefully", h.name)
			}
		}
	}()

	return nil
}

func (h *Host) validate() *Host {

	if target, err := linkTarget(h.root); err != nil {
		h.status = Error
		h.Error = NotFound
		Logger().Printf("[AT-AT] skipping '%s': %s\n", h.name, err)
	} else {
		h.root = target
	}

	if !fileExists(path.Join(h.root, ".at_at")) {
		h.status = Error
		h.Error = Unavailable
		Logger().Printf("[AT-AT] skipping '%s': no .at_at file found\n", h.name)
	}

	if h.Error == NoError {
		Logger().Printf("[%s] serving host '%s' from folder %s\n", h.name, h.name, h.root)
	}

	return h
}

func linkTarget(link string) (string, error) {
	if target, err := filepath.EvalSymlinks(link); err == nil {
		if path, err := filepath.Abs(target); err == nil {
			return path, nil
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func createCommand(cmd string) *exec.Cmd {
	return exec.Command(findExecutable("sh"), "-c", cmd)
}

func findExecutable(cmd string) string {
	ret, _ := exec.LookPath(cmd)
	return ret
}
