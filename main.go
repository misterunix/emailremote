package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/knadh/go-pop3"
	"github.com/misterunix/sniffle/jsonio"
	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Emailaddress  string
	Emailpassword string
}

type Commands struct {
	From    string
	Subject string
	Body    string
}

var commands []Commands
var config EmailConfig

func main() {

	fmt.Println("Hello World")

	home := os.Getenv("HOME")
	fmt.Println(home)

	configFilePath := "~/.emailremote/config.json"
	configFilePath = buildHomePath(configFilePath)

	if !fileExists(configFilePath) {
		config.Emailaddress = "none@example.com"
		config.Emailpassword = "12345"
		jsonio.SaveJSon(configFilePath, config)
		fmt.Println("Basic configuration json file created.")
		os.Exit(0)
	}

	jsonio.LoadJSon(configFilePath, &config)

	popit()
	l := len(commands)
	fmt.Println("Command Count", l)

	if l == 0 {
		return // we are done
	}

	for _, c := range commands {

		s := strings.ToUpper(c.Subject)
		f := c.From

		sc := strings.Split(s, " ") // sc : Split commands

		if strings.Compare(sc[0], "LIST") == 0 {
			List(f)
		}

		fmt.Println("sc size", len(sc))

		// check for # of works on the subject line
		if len(sc) < 2 {
			continue
		}

		fmt.Printf("%+v\n", sc)

		sc[1] = strings.TrimSpace(sc[1])
		if strings.Compare(sc[0], "PING") == 0 {
			if !checkIPAddress(sc[1]) {
				ReturnEmail(c.From, s, "FQHN not supported at this time.")
				continue
			}
			ss, err := RunExecutable(1, sc[1])
			if err != nil {
				continue
			}
			ReturnEmail(c.From, s, ss)

		}

		if strings.Compare(sc[0], "TRACE-U") == 0 {
			if !checkIPAddress(sc[1]) {
				ReturnEmail(c.From, s, "FQHN not supported at this time.")
				continue
			}
			ss, err := RunExecutable(2, sc[1])
			if err != nil {
				continue
			}
			ReturnEmail(c.From, s, ss)

		}

		if strings.Compare(sc[0], "TRACE-I") == 0 {
			if !checkIPAddress(sc[1]) {
				ReturnEmail(c.From, s, "FQHN not supported at this time.")
				continue
			}
			ss, err := RunExecutable(4, sc[1])
			if err != nil {
				continue
			}
			ReturnEmail(c.From, s, ss)

		}

		if strings.Compare(sc[0], "MTR") == 0 {
			if !checkIPAddress(sc[1]) {
				ReturnEmail(c.From, s, "FQHN not supported at this time.")
				continue
			}
			ss, err := RunExecutable(3, sc[1])
			if err != nil {
				continue
			}
			ReturnEmail(c.From, s, ss)

		}

	}

}

// buildHomePath : expand ~/ to full home path
func buildHomePath(path string) string {
	home := os.Getenv("HOME")
	// expand tilde
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(home, path[2:])
	}

	return path
}

// RunExecutable : Runs the selected command and returns the result
func RunExecutable(cmd int, parameter string) (string, error) {
	var c *exec.Cmd

	switch cmd {
	case 1: // ping IP
		c = exec.Command("ping", "-c", "10", parameter)
	case 2: // traceroute IP
		c = exec.Command("traceroute", parameter)
	case 3:
		c = exec.Command("mtr", "-c", "10", "-r", parameter)
	case 4: // traceroute IP
		c = exec.Command("traceroute", "-I", parameter)

	default:
		return "", errors.New("o command given or recognized")
	}

	fmt.Printf("%+v\n", c)
	//c := exec.Command(SConfig.ExecutablePath, "-b", "-P", py, fn)

	stdoutpipe, err := c.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = c.Start()
	if err != nil {
		return "", err
	}

	rawdata, err := ioutil.ReadAll(stdoutpipe)
	if err != nil {
		return "", err
	}

	/*
		poutput := strings.Split(string(rawdata), "\n")

		for _, v := range poutput {
			if !strings.HasSuffix(v, "Scene") {
				continue
			}
			parts := strings.Split(v, " ")
			if len(parts) != 3 {
				fmt.Println("Could not parse results from exec. Please add an issue to the dev team.")
				fmt.Println("Stopping server.")
				CleanQuit()
			}
			startframe, _ := strconv.Atoi(parts[0])
			endframe, _ := strconv.Atoi(parts[1])
			return startframe, endframe
		}
		return "",nil
	*/

	// fmt.Println("Command returned", len(rawdata), "bytes")
	return string(rawdata), nil
}

// ReturnEmail : Send the results from the commands to the sender.
func ReturnEmail(returnemail string, subject string, message string) error {
	fmt.Println("Return Email")
	fmt.Println(returnemail)
	fmt.Println(subject)
	fmt.Println(message)

	m := gomail.NewMessage()
	m.SetHeader("From", config.Emailaddress)
	m.SetHeader("To", returnemail)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)
	//m.AddAlternative("text/html", bodyh)
	//m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.gmail.com", 587, config.Emailaddress, config.Emailpassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	return nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Connect to email and retreive the emails and store them to be processed.
func popit() {

	// gmail is hard coded at this time.
	p := pop3.New(pop3.Opt{
		Host:       "pop.gmail.com",
		Port:       995,
		TLSEnabled: true,
	})

	// Create a new pop3
	c, err := p.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Quit()

	// Authenticate.
	if err := c.Auth(config.Emailaddress, config.Emailpassword); err != nil {
		log.Fatal(err)
	}

	// count : The number of messages
	count, _, _ := c.Stat()
	//fmt.Println("total messages=", count, "size=", size)

	/* used during debuging
	// Pull the list of all message IDs and their sizes.
	msgs, _ := c.List(0)
	for _, m := range msgs {
		fmt.Println("id=", m.ID, "size=", m.Size)
	}
	*/

	// Pull all messages on the server. Message IDs go from 1 to N.
	for id := 1; id <= count; id++ {
		m, _ := c.Retr(id)

		from := m.Header.Get("From")
		subject := m.Header.Get("Subject")
		if len(from) < 4 {
			continue
		}

		// The body of the email is not needed

		cc := Commands{}                // cc : New tempory Commands struct
		cc.From = from                  // Set the from
		cc.Subject = subject            // Set the subject
		commands = append(commands, cc) // Add from and subject into the command slice

	}

	// Delete all the messages. Server only executes deletions after a successful Quit()
	//for id := 1; id <= count; id++ {
	//c.Dele(id)
	//}

	//	for k, j := range commands {
	//		fmt.Println(k, j)
	//	}

}

// checkIPAddress : Check to see if an IP address is valid. Returns true if valid, false if not.
func checkIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		if govalidator.IsDNSName(ip) {
			return true
		}
		return false
	} else {
		return true
	}
}
