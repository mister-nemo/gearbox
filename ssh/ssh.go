package ssh

import (
	"fmt"
	"gearbox/help"
	"github.com/fatih/color"
	"github.com/gearboxworks/go-status"
	"github.com/gearboxworks/go-status/is"
	"github.com/gearboxworks/go-status/only"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

type SSH struct {
	Instance *ssh.Client
	Session  *ssh.Session

	//	// Status polling delays.
	//	NoWait        bool
	//	WaitDelay     time.Duration
	//	WaitRetries   int

	// SSH related.
	Username   string
	Password   string
	Host       string
	Port       string
	PublicKey  string
	StatusLine StatusLine
}

type StatusLine struct {
	Text          string
	Disable       bool
	UpdateDelay   time.Duration
	TermWidth     int
	TermHeight    int
	TerminateFlag bool
}

type Args SSH

const DefaultUsername = "gearbox"
const DefaultPassword = "box"
const DefaultKeyFile = "./keyfile.pub"
const DefaultSshHost = "gearbox.local"
const DefaultSshPort = "22"
const DefaultStatusLineUpdateDelay = time.Second * 2

func NewSSH(args ...Args) *SSH {

	var _args Args
	if len(args) > 0 {
		_args = args[0]
	}

	if _args.Username == "" {
		_args.Username = DefaultUsername
	}

	if _args.Password == "" {
		_args.Password = DefaultPassword
	}

	if _args.PublicKey == "" {
		_args.PublicKey = DefaultKeyFile
	}

	if _args.StatusLine.UpdateDelay == 0 {
		_args.StatusLine.UpdateDelay = DefaultStatusLineUpdateDelay
	}

	if _args.Host == "" {
		_args.Host = DefaultSshHost
	}

	if _args.Port == "" {
		_args.Port = DefaultSshPort
	}

	sshClient := &SSH{}
	*sshClient = SSH(_args)

	// Query VB to see if it exists.
	// If not return nil.

	return sshClient
}

func readPublicKeyFile(file string) (ssh.AuthMethod, error) {

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		// fmt.Printf("# Error reading file '%s': %s\n", file, err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		// fmt.Printf("# Error parsing key '%s': %s\n", signer, err)
		return nil, err
	}

	return ssh.PublicKeys(signer), err
}

func (me *SSH) StartSSH() (sts status.Status) {

	var err error
	for range only.Once {
		sts = EnsureNotNil(me)
		if is.Error(sts) {
			break
		}

		sshConfig := &ssh.ClientConfig{}

		var auth []ssh.AuthMethod

		// Try SSH key file first.
		var keyfile ssh.AuthMethod
		keyfile, err = readPublicKeyFile(me.PublicKey)

		if err == nil && keyfile != nil {
			// Authenticate using SSH key.
			auth = []ssh.AuthMethod{keyfile}
		} else {
			// Authenticate using password
			auth = []ssh.AuthMethod{ssh.Password(me.Password)}
		}

		sshConfig = &ssh.ClientConfig{
			User: me.Username,
			Auth: auth,
			// HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second * 10,
		}

		me.Instance, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", me.Host, me.Port), sshConfig)
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: "SSH error",
			})
			break
		}

		me.Session, err = me.Instance.NewSession()
		defer me.Session.Close()
		defer me.Instance.Close()
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: "SSH error",
			})
			break
		}

		// Set IO
		me.Session.Stdout = os.Stdout
		me.Session.Stderr = os.Stderr
		me.Session.Stdin = os.Stdin
		//	in, _ := me.Session.StdinPipe()

		// Set up terminal modes
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 19200,
			ssh.TTY_OP_OSPEED: 19200,
		}

		// Request pseudo terminal
		fileDescriptor := int(os.Stdin.Fd())
		if terminal.IsTerminal(fileDescriptor) {
			originalState, err := terminal.MakeRaw(fileDescriptor)
			if err != nil {
				sts = status.Wrap(err)
				break
			}
			defer terminal.Restore(fileDescriptor, originalState)

			me.StatusLine.TermWidth, me.StatusLine.TermHeight, err = terminal.GetSize(fileDescriptor)
			if err != nil {
				sts = status.Wrap(err)
				break
			}

			// xterm-256color
			err = me.Session.RequestPty("xterm-256color", me.StatusLine.TermHeight, me.StatusLine.TermWidth, modes)
			if err != nil {
				sts = status.Wrap(err)
				break
			}
		}

		go me.StatusLineWorker()

		go me.exampleHostWorker()

		// Start remote shell
		err = me.Session.Shell()
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: "SSH error",
			})
			break
		}

		/*
			// Loop around input <-> output.
			for {
				reader := bufio.NewReader(os.Stdin)
				str, _ := reader.ReadString('\n')
				fmt.Fprint(in, str)
			}
		*/
		me.Session.Wait()
		me.resetView()
	}
	return sts
}

// StatusLineWorker() - handles the actual updates to the status line
func (me *SSH) StatusLineWorker() {

	me.setView()
	// w := gob.NewEncoder(me.Session)
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, syscall.SIGWINCH)

	for me.StatusLine.TerminateFlag == false {
		// Handle terminal windows size changes properly.
		fileDescriptor := int(os.Stdin.Fd())
		width, height, _ := terminal.GetSize(fileDescriptor)
		if (me.StatusLine.TermWidth != width) || (me.StatusLine.TermHeight != height) {
			me.StatusLine.TermWidth = width
			me.StatusLine.TermHeight = height
			// me.Session.Signal(syscall.SIGWINCH)
			me.Session.WindowChange(height, width)
		} else {
			// Only update if we haven't seen a SIGWINCH - just to wait for things to settle.
			me.displayStatusLine()
		}

		time.Sleep(me.StatusLine.UpdateDelay)
	}

}

func (me *SSH) SetStatusLine(text string) {

	me.StatusLine.Text = text
}

func (me *SSH) displayStatusLine() {
	const savePos = "\033[s"
	const restorePos = "\033[u"
	bottomPos := fmt.Sprintf("\033[%d;0H", me.StatusLine.TermHeight)
	// topPos := fmt.Sprintf("\033[0;0H")

	if me.StatusLine.Disable == false {
		fmt.Printf("%s%s%s%s", savePos, bottomPos, me.StatusLine.Text, restorePos)
	}
}

func (me *SSH) setView() {
	const clearScreen = "\033[H\033[2J"
	scrollFixBottom := fmt.Sprintf("\033[1;%dr", me.StatusLine.TermHeight-1)
	// scrollFixTop := fmt.Sprintf("\033[2;%dr", termHeight)

	if me.StatusLine.Disable == false {
		fmt.Printf(scrollFixBottom)
		fmt.Printf(clearScreen)
	}
}

func (me *SSH) resetView() {
	const savePos = "\033[s"
	const restorePos = "\033[u"
	scrollFixBottom := fmt.Sprintf("\033[1;%dr", me.StatusLine.TermHeight)
	// scrollFixTop := fmt.Sprintf("\033[2;%dr", termHeight)

	if me.StatusLine.Disable == false {
		fmt.Printf(savePos)
		fmt.Printf(scrollFixBottom)
		fmt.Printf(restorePos)

		me.StatusLine.Text = ""
		for i := 0; i <= me.StatusLine.TermWidth; i++ {
			me.StatusLine.Text += " "
		}
		me.displayStatusLine()
	}

}

func stripAnsi(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)

	return re.ReplaceAllString(str, "")
}

// Example host worker. This periodically changes the me.StatusLine.Text from the host side.
// The StatusLineWorker() will update the bottom line using the me.StatusLine.Text.
func (me *SSH) exampleHostWorker() {

	yellow := color.New(color.BgBlack, color.FgHiYellow).SprintFunc()
	magenta := color.New(color.BgBlack, color.FgHiMagenta).SprintFunc()
	green := color.New(color.BgBlack, color.FgHiGreen).SprintFunc()
	normal := color.New(color.BgWhite, color.FgHiBlack).SprintFunc()

	for me.StatusLine.TerminateFlag == false {
		//now := time.Now()
		//dateStr := normal("Date:") + " " + yellow(fmt.Sprintf("%.4d/%.2d/%.2d", now.Year(), now.Month(), now.Day()))
		//timeStr := normal("Time:") + " " + magenta(fmt.Sprintf("%.2d:%.2d:%.2d", now.Hour(), now.Minute(), now.Second()))
		statusStr := normal("Status:") + " " + green("OK")
		infoStr := yellow("You are connected to") + " " + magenta("Gearbox OS")

		//line := fmt.Sprintf("%s	%s %s", statusStr, dateStr, timeStr)
		line := fmt.Sprintf("%s - %s", infoStr, statusStr)

		// Add spaces to ensure it's right justified.
		spaces := ""
		lineLen := len(stripAnsi(line))
		for i := 0; i < me.StatusLine.TermWidth-lineLen; i++ {
			spaces += " "
		}

		me.SetStatusLine(spaces + line) // + fmt.Sprintf("W:%d L:%d S:%d C:%d", me.StatusLine.TermWidth, len(line), len(spaces), lineLen))

		time.Sleep(time.Second * 5)
	}
}

func EnsureNotNil(ssh *SSH) (sts status.Status) {
	if ssh == nil {
		sts = status.Fail(&status.Args{
			Message: "unexpected error",
			Help:    help.ContactSupportHelp(), // @TODO need better support here
		})
	}
	return sts
}
