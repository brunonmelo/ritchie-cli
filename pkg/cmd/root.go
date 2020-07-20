package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZupIT/ritchie-cli/pkg/file/fileutil"
	"github.com/ZupIT/ritchie-cli/pkg/prompt"
	"github.com/ZupIT/ritchie-cli/pkg/slice/sliceutil"
	"github.com/ZupIT/ritchie-cli/pkg/stream"
	"github.com/ZupIT/ritchie-cli/pkg/version"
)

const (
	latestVersionMsg            = "Latest available version: %s"
	versionMsg                  = "%s\n  Build date: %s\n  Built with: %s\n"
	versionMsgWithLatestVersion = "%s\n  %s\n  Build date: %s\n  Built with: %s\n"
	cmdUse                      = "rit"
	cmdShortDescription         = "rit is a NoOps CLI"
	cmdDescription              = `A CLI that developers can build and operate
your applications without help from the infra staff.
Complete documentation available at https://github.com/ZupIT/ritchie-cli`
)

var (
	Version = "dev"
	BuildDate = "unknown"
	StableVersionUrl = "https://commons-repo.ritchiecli.io/stable.txt"
	ErrRitInit = errors.New("To start using rit, you need to initialize rit first.\nCommand: rit init")

	allowList = []string{
		fmt.Sprint(cmdUse),
		fmt.Sprintf("%s help", cmdUse),
		fmt.Sprintf("%s completion zsh", cmdUse),
		fmt.Sprintf("%s completion bash", cmdUse),
		fmt.Sprintf("%s completion fish", cmdUse),
		fmt.Sprintf("%s completion powershell", cmdUse),
		fmt.Sprintf("%s init", cmdUse),
		fmt.Sprintf("%s upgrade", cmdUse),
	}

	upgradeList = []string{
		fmt.Sprint(cmdUse),
	}
)

type rootCmd struct {
	ritchieHome string
	dir         stream.DirCreateChecker
}

func NewRootCmd(ritchieHome string, dir stream.DirCreateChecker) *cobra.Command {
	o := &rootCmd{ritchieHome: ritchieHome, dir: dir}

	cmd := &cobra.Command{
		Use:                cmdUse,
		Short:              cmdShortDescription,
		Long:               cmdDescription,
		Version:            versionFlag(),
		PersistentPreRunE:  o.PreRunFunc(),
		PersistentPostRunE: o.PostRunFunc(),
		RunE:               runHelp,
		SilenceErrors:      true,
		TraverseChildren:   true,
	}
	cmd.PersistentFlags().Bool("stdin", false, "input by stdin")

	return cmd
}

func (ro *rootCmd) PreRunFunc() CommandRunnerFunc {
	return func(cmd *cobra.Command, args []string) error {
		if err := ro.dir.Create(ro.ritchieHome); err != nil {
			return err
		}

		if isAllowList(allowList, cmd) || isCompleteCmd(cmd) {
			return nil
		}

		commonsRepoPath := path.Join(ro.ritchieHome, "repos", "commons")
		if !ro.dir.Exists(commonsRepoPath) {
			return ErrRitInit
		}

		return nil
	}
}

func (ro *rootCmd) PostRunFunc() CommandRunnerFunc {
	return func(cmd *cobra.Command, args []string) error {
		verifyNewVersion(cmd)
		return nil
	}
}

func verifyNewVersion(cmd *cobra.Command) {
	if isAllowList(upgradeList, cmd) {
		resolver := version.DefaultVersionResolver{
			StableVersionUrl: StableVersionUrl,
			FileUtilService:  fileutil.DefaultService{},
			HttpClient:       &http.Client{Timeout: 1 * time.Second},
		}
		prompt.Warning(version.VerifyNewVersion(resolver, Version))
	}
}

func isAllowList(allowList []string, cmd *cobra.Command) bool {
	return sliceutil.Contains(allowList, cmd.CommandPath())
}

func isCompleteCmd(cmd *cobra.Command) bool {
	return strings.Contains(cmd.CommandPath(), "__complete")
}

func versionFlag() string {
	resolver := version.DefaultVersionResolver{
		StableVersionUrl: StableVersionUrl,
		FileUtilService:  fileutil.DefaultService{},
		HttpClient:       &http.Client{Timeout: 1 * time.Second},
	}
	latestVersion, err := resolver.StableVersion()
	if err == nil && latestVersion != Version {
		formattedLatestVersionMsg := prompt.Yellow(fmt.Sprintf(latestVersionMsg, latestVersion))
		return fmt.Sprintf(versionMsgWithLatestVersion, Version, formattedLatestVersionMsg, BuildDate, runtime.Version())
	}
	return fmt.Sprintf(versionMsg, Version, BuildDate, runtime.Version())
}

func runHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
