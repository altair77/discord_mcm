package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	shellwords "github.com/mattn/go-shellwords"
)

type Manager struct {
	session *discordgo.Session
	config  *Config
	log     chan string
	command *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
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
	m.session.Close()
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
		m.launchServer(s, mc)
	case strings.HasPrefix(c, m.config.Prefix+"stop"):
		m.stopServer(s, mc)
	case strings.HasPrefix(c, m.config.Prefix+"cmd"):
		m.execServer(s, mc)
	case strings.HasPrefix(c, m.config.Prefix+"log"):
		m.showLog(s, mc)
	}
}

func (m *Manager) launchServer(s *discordgo.Session, mc *discordgo.MessageCreate) {
	launchCommands, err := shellwords.Parse(m.config.LaunchCommand)
	if err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Error: launchCommand is worng!")
		return
	}
	if m.command != nil && m.command.Process.Pid > 0 {
		m.command = nil
		s.ChannelMessageSend(m.config.ChannelID, "Server is running.")
		return
	}
	m.command = exec.Command(launchCommands[0], launchCommands[1:]...)
	m.stdin, err = m.command.StdinPipe()
	if err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to get stdin pipe.")
		return
	}
	m.stdout, err = m.command.StdoutPipe()
	if err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to get stdout pipe.")
		return
	}
	m.readLog()
	if err := m.command.Start(); err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to start server.")
		return
	}
	s.ChannelMessageSend(m.config.ChannelID, "Started server!")
}

func (m *Manager) stopServer(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if m.command == nil || m.command.Process.Pid <= 0 {
		s.ChannelMessageSend(m.config.ChannelID, "Server is stopped.")
		return
	}
	writer := bufio.NewWriter(m.stdin)
	if _, err := writer.WriteString("stop\n"); err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}
	writer.Flush()
	if err := m.command.Wait(); err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to stop server.")
		return
	}
	s.ChannelMessageSend(m.config.ChannelID, "Success to stop server.")
}

func (m *Manager) execServer(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if m.command == nil || m.command.Process.Pid <= 0 {
		s.ChannelMessageSend(m.config.ChannelID, "Server is stopped.")
		return
	}

	ctx := context.Background()
	if _, err := m.readTimeout(ctx, 1); err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to read pre log.")
		return
	}

	writer := bufio.NewWriter(m.stdin)
	subCmd := mc.Content[len(m.config.Prefix)+4:]
	if _, err := writer.WriteString(subCmd + "\n"); err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Failed to send command.")
		return
	}
	writer.Flush()

	log, err := m.readTimeout(ctx, 1)
	if err != nil {
		s.ChannelMessageSend(m.config.ChannelID, "Fialed to read result log.")
		return
	}
	logLen := len(log)
	if logLen >= 1800 {
		logLen = 1800
	}
	s.ChannelMessageSend(m.config.ChannelID, "Sent Command.\n```\n"+log[:logLen]+"\n```")
}

func (m *Manager) showLog(s *discordgo.Session, mc *discordgo.MessageCreate) {

}

func (m *Manager) readLog() {
	m.log = make(chan string)
	go func() {
		buff := make([]byte, 1024)
		var err error
		var n int
		for err == nil {
			n, err = m.stdout.Read(buff)
			if n > 0 {
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
