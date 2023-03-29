package cmdwrapper

import (
	"os"
	"os/exec"

	"go.uber.org/zap"
)

type CmdWrapper struct {
	cmd    *exec.Cmd
	logger *zap.SugaredLogger
	sigs   <-chan os.Signal
}

func NewCmdWrapper(logger *zap.SugaredLogger) *CmdWrapper {

	args := os.Args[1:]
	binary, err := exec.LookPath(args[0])
	if err != nil {
		logger.Panicf("Command not found: %v", err)
	}

	cmd := exec.Command(binary, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	logger.Infof("Created wrapper for command: %v", cmd)

	return &CmdWrapper{
		cmd:    cmd,
		logger: logger,
	}
}

func (c *CmdWrapper) Subscribe(sigs <-chan os.Signal) {
	c.sigs = sigs

	go func() {
		for sig := range c.sigs {
			c.logger.Infof("Received signal: %v", sig)
			c.signal(sig)
		}
	}()
}

func (c *CmdWrapper) Start() {
	err := c.cmd.Start()

	if err != nil {
		c.logger.Panicf("Unable to run command: %v", err)
	}
}

func (c *CmdWrapper) Wait() (int, error) {
	c.logger.Debug("Start wait")
	err := c.cmd.Wait()
	c.logger.Debug("End wait")

	return c.cmd.ProcessState.ExitCode(), err
}

func (c *CmdWrapper) signal(sig os.Signal) {
	if c.cmd != nil && c.cmd.ProcessState != nil && c.cmd.ProcessState.Exited() {
		c.logger.Debug("Wrapped command not running. Skipping sending signal")
		return
	}

	c.logger.Infof("Sending signal to wrapped command: %v", sig)

	err := c.cmd.Process.Signal(sig)

	if err != nil {
		c.logger.Errorf("Unable to send signal to wrapped command: %v", err)
	}
}
