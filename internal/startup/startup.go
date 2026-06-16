package startup

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/proc"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "startup",
	Short: "Enforce Session Startup Protocol",
	Long: `Enforce and verify the team-flow Session Startup Protocol.

This command ensures that all mandatory startup steps have been completed
before any project work begins.

Steps verified:
  1. flow project detect - Project identification
  2. flow proc run --new - Session creation  
  3. Role adoption - Reading alias and persona

Example:
  flow startup check     # Check if startup is complete
  flow startup enforce   # Enforce startup, exit with error if not complete`,
	RunE: runStartup,
}

var (
	checkMode   bool
	enforceMode bool
)

func init() {
	Cmd.Flags().BoolVar(&checkMode, "check", false, "Check startup status without enforcing")
	Cmd.Flags().BoolVar(&enforceMode, "enforce", false, "Enforce startup protocol")
}

func runStartup(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	
	status, err := proc.CheckSessionStartup(cwd)
	if err != nil {
		return fmt.Errorf("check startup: %w", err)
	}

	printStatus(status)

	if enforceMode && !status.Completed {
		fmt.Println("\n" + proc.GetStartupViolationMessage(status))
		os.Exit(1)
	}

	if !status.Completed {
		fmt.Println("\n⚠️ Session Startup Protocol 未完成")
		return nil
	}

	fmt.Println("\n✅ Session Startup Protocol 已完成")
	return nil
}

func printStatus(status proc.SessionStartupStatus) {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      Session Startup Protocol Status     ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	
	fmt.Println("\nStep 1: flow project detect")
	if status.Step1Detected {
		fmt.Printf("  ✅ 已执行\n")
		fmt.Printf("     Project: %s\n", status.Step1ProjectName)
		fmt.Printf("     Root: %s\n", status.Step1ProjectRoot)
	} else {
		fmt.Println("  ❌ 未执行")
	}

	fmt.Println("\nStep 2: flow proc run --new")
	if status.Step2SessionCreated {
		fmt.Printf("  ✅ 已执行\n")
		fmt.Printf("     Session ID: %s\n", status.Step2SessionID)
	} else {
		fmt.Println("  ❌ 未执行")
	}

	fmt.Println("\nStep 3: Adopt Role")
	if status.Step3RoleAdopted {
		fmt.Printf("  ✅ 已执行\n")
		fmt.Printf("     Role: %s\n", status.Step3RoleAlias)
	} else {
		fmt.Println("  ❌ 未执行")
	}

	fmt.Println("\n──────────────────────────────────────────")
	if status.Completed {
		fmt.Println("✓ 全部完成")
	} else {
		fmt.Println("✗ 未完成")
	}
}
