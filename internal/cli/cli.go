package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/winkeep/winkeep/internal/config"
	"github.com/winkeep/winkeep/internal/process"
	"github.com/winkeep/winkeep/internal/session"
)

var (
	version = "dev"
	cfg     *config.Config
	procMgr *process.Manager
	sessMgr *session.Manager
)

func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "winkeep",
	Short: "Keep your processes running after SSH disconnect on Windows",
	Long: `WinKeep is a tool that keeps processes running in the background
even after your SSH session is closed. Like nohup, but for Windows.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		procMgr = process.NewManager(cfg.BaseDir)
		sessMgr = session.NewManager(cfg.BaseDir, procMgr)
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Run a command in the background",
	Long:  "Run a command that will keep running even after SSH disconnects.",
	Example: `  winkeep run -- python server.py
  winkeep run --name myapp -- node app.js
  winkeep run --session myproject -- npm start`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		sessionName, _ := cmd.Flags().GetString("session")

		if len(args) == 0 {
			return fmt.Errorf("command is required. Usage: winkeep run -- <command> [args...]")
		}

		command := args[0]
		cmdArgs := []string{}
		if len(args) > 1 {
			cmdArgs = args[1:]
		}

		if name == "" {
			name = command
		}

		var sessionID string
		if sessionName != "" {
			s, err := sessMgr.GetByName(sessionName)
			if err != nil {
				s, err = sessMgr.Create(sessionName)
				if err != nil {
					return fmt.Errorf("create session: %w", err)
				}
			}
			sessionID = s.ID
		}

		p, err := procMgr.Run(name, command, cmdArgs, sessionID)
		if err != nil {
			return fmt.Errorf("run process: %w", err)
		}

		if sessionID != "" {
			sessMgr.AddProcess(sessionID, p.ID)
		}

		fmt.Printf("Process started in background\n")
		fmt.Printf("  ID:   %s\n", p.ID)
		fmt.Printf("  PID:  %d\n", p.PID)
		fmt.Printf("  Logs: winkeep logs %s\n", p.ID)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all background processes",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		processes, err := procMgr.List()
		if err != nil {
			return err
		}

		if len(processes) == 0 {
			fmt.Println("No background processes.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPID\tSTATUS\tUPTIME")
		for _, p := range processes {
			uptime := time.Since(p.StartTime).Truncate(time.Second)
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", p.ID, p.Name, p.PID, p.Status, uptime)
		}
		w.Flush()
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Show logs of a background process",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tail, _ := cmd.Flags().GetInt("tail")
		output, err := procMgr.Logs(args[0], tail)
		if err != nil {
			return err
		}
		if output == "" {
			fmt.Println("(no logs)")
		} else {
			fmt.Print(output)
		}
		return nil
	},
}

var killCmd = &cobra.Command{
	Use:   "kill <id>",
	Short: "Stop a background process",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := procMgr.Kill(args[0]); err != nil {
			return err
		}
		fmt.Printf("Process %s stopped\n", args[0])
		return nil
	},
}

var sessionCmd = &cobra.Command{
	Use:     "session",
	Short:   "Manage sessions",
	Aliases: []string{"sess", "s"},
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := sessMgr.Create(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Session created: %s (ID: %s)\n", s.Name, s.ID)
		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all sessions",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := sessMgr.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPROCESSES\tCREATED")
		for _, s := range sessions {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.ID, s.Name, len(s.Processes), s.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		w.Flush()
		return nil
	},
}

var sessionKillCmd = &cobra.Command{
	Use:   "kill <name>",
	Short: "Kill all processes in a session and remove it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sessMgr.KillByName(args[0]); err != nil {
			return err
		}
		fmt.Printf("Session %s killed\n", args[0])
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("winkeep %s\n", version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	runCmd.Flags().StringP("name", "n", "", "Process name (default: command name)")
	runCmd.Flags().StringP("session", "s", "", "Attach to session")

	logsCmd.Flags().IntP("tail", "t", 50, "Number of lines to show")

	sessionCmd.AddCommand(sessionCreateCmd, sessionListCmd, sessionKillCmd)

	rootCmd.AddCommand(runCmd, listCmd, logsCmd, killCmd, sessionCmd, versionCmd)
}
