package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-shellwords"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type Manager struct {
	session   *discordgo.Session
	config    *Config
	log       chan string
	enableLog bool
	command   *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
}

func NewManager(c *Config) *Manager {
	m := &Manager{}
	m.config = c
	return m
}

func (m *Manager) Start() error {
	session, err := discordgo.New("Bot " + m.config.Token)
	if err != nil {
		return err
	}
	m.session = session
	session.AddHandler(func(s *discordgo.Session, mc *discordgo.MessageCreate) {
		m.createMessageHandler(s, mc)
	})
	session.Identify.Intents = discordgo.IntentsGuildMessages

	if err := session.Open(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Close() {
	_ = m.session.Close()
}

func (m *Manager) createMessageHandler(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == s.State.User.ID {
		return
	}
	if mc.ChannelID != m.config.ChannelID {
		return
	}
	log.Print(mc.Content)
	switch c := mc.Content; {
	case strings.HasPrefix(c, m.config.Prefix+"start"):
		m.launchServer(s)
	case strings.HasPrefix(c, m.config.Prefix+"stop"):
		m.stopServer(s)
	case strings.HasPrefix(c, m.config.Prefix+"cmd"):
		m.execServer(s, mc)
	case strings.HasPrefix(c, m.config.Prefix+"status"):
		m.showStatus(s)
	}
}

func (m *Manager) launchServer(s *discordgo.Session) {
	launchCommands, err := shellwords.Parse(m.config.LaunchCommand)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Error: launchCommand is wrong!")
		return
	}
	if m.command != nil && m.command.Process.Pid > 0 {
		m.command = nil
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Server is running.")
		return
	}
	m.command = exec.Command(launchCommands[0], launchCommands[1:]...)
	m.stdin, err = m.command.StdinPipe()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to get stdin pipe.")
		return
	}
	m.stdout, err = m.command.StdoutPipe()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to get stdout pipe.")
		return
	}
	m.readLog()
	if err := m.command.Start(); err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to start server.")
		return
	}
	_, _ = s.ChannelMessageSend(m.config.ChannelID, "Started server!")
}

func (m *Manager) stopServer(s *discordgo.Session) {
	if m.command == nil || m.command.Process.Pid <= 0 {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Server is stopped.")
		return
	}
	writer := bufio.NewWriter(m.stdin)
	if _, err := writer.WriteString("stop\n"); err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}
	err := writer.Flush()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}
	if err := m.command.Wait(); err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to stop server.")
		return
	}
	_, _ = s.ChannelMessageSend(m.config.ChannelID, "Success to stop server.")
}

func (m *Manager) execServer(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if len(mc.Content) < len(m.config.Prefix)+4 || len(strings.TrimSpace(mc.Content[len(m.config.Prefix)+4:])) == 0 {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Usage: "+m.config.Prefix+"cmd <server command>")
		return
	}

	if m.command == nil || m.command.Process.Pid <= 0 {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Server is stopped.")
		return
	}

	writer := bufio.NewWriter(m.stdin)
	subCmd := mc.Content[len(m.config.Prefix)+4:]
	if _, err := writer.WriteString(subCmd + "\n"); err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}
	err := writer.Flush()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}

	ctx := context.Background()
	res, err := m.readTimeout(ctx, 1)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to read result log.")
		return
	}
	logLen := len(res)
	if logLen >= 1800 {
		logLen = 1800
	}
	_, _ = s.ChannelMessageSend(m.config.ChannelID, "Sent Command.\n```\n"+res[:logLen]+"\n```")
}

func (m *Manager) showStatus(s *discordgo.Session) {
	isRunning := "Stopped"
	if m.command == nil || m.command.Process.Pid <= 0 {
		isRunning = "Running"
	}
	v, err := mem.VirtualMemory()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to get memory information.")
	}
	a, err := load.Avg()
	if err != err {
		_, _ = s.ChannelMessageSend(m.config.ChannelID, "Failed to get load information.")
	}

	message := fmt.Sprintf(
		"Server: %s\nLoad: 1min %.3f 5min %.3f 15min %.3f\nMemory(GB): total %2.3f used %2.3f per %2.1f",
		isRunning,
		a.Load1,
		a.Load5,
		a.Load15,
		float64(v.Total)/1000000000,
		float64(v.Used)/1000000000,
		v.UsedPercent,
	)
	_, _ = s.ChannelMessageSend(m.config.ChannelID, message)
}

func (m *Manager) readLog() {
	m.log = make(chan string)
	go func() {
		buff := make([]byte, 1024)
		var err error
		var n int
		for err == nil {
			n, err = m.stdout.Read(buff)
			if n > 0 && m.enableLog {
				m.log <- string(buff[:n])
			}
		}
		close(m.log)
	}()
}

func (m *Manager) readTimeout(ctx context.Context, t int) (string, error) {
	str := ""
	done := make(chan struct{})
	defer close(done)
	m.enableLog = true
	defer func() { m.enableLog = false }()

	go func() {
		prevLen := 0
		count := 0
		for {
			time.Sleep(time.Second)
			if len(str) == prevLen {
				count += 1
			} else {
				count = 0
			}
			if count >= t {
				done <- struct{}{}
				return
			}
			prevLen = len(str)
		}
	}()

	for {
		select {
		case s := <-m.log:
			str += s
		case <-done:
			return str, nil
		case <-ctx.Done():
			return str, ctx.Err()
		}
	}
}
