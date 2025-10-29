package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTagType       = "fix"
	versionFlagNameLong  = "version"
	versionFlagNameShort = "v"
	allowBranch          = "main"
	usageLong            = "-version [release|fix]"
	usageShort           = "shorthand of 'version'"
	usageHelper          = `Usage: %s
Options:
`
	expectedVersionParts = 5
)

var branchRegex = regexp.MustCompile(allowBranch)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Info: All set, Goodbye\n")
}

func run() error {
	// Parse flags
	tagType := parseFlags()

	// Validate branch
	isStaging, err := validateBranch()
	if err != nil {
		return err
	}

	// Check if current hash is already tagged
	if err := checkCurrentHashNotTagged(); err != nil {
		return err
	}

	// Calculate new version
	newVersion, err := calculateNewVersion(tagType, isStaging)
	if err != nil {
		return err
	}

	// Tag and push
	return tagAndPush(newVersion)
}

func parseFlags() string {
	var tagType string
	flag.StringVar(&tagType, versionFlagNameLong, defaultTagType, usageLong)
	flag.StringVar(&tagType, versionFlagNameShort, defaultTagType, usageShort)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageHelper, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	checkFlagIsPassed := isFlagPassed()
	if !checkFlagIsPassed && tagType != "release" {
		fmt.Printf("You are not provide the parameter, the default will tagging as 'fix'\n")
		confirm := askConfirm("Please confirm to continue")
		if !confirm {
			fmt.Fprintf(os.Stdout, "Alright then, see ya!\n")
			os.Exit(0)
		}
	}

	return tagType
}

func validateBranch() (bool, error) {
	checkBranchMsgOut, err := checkCurrentBranch()
	if err != nil {
		return false, fmt.Errorf("failed to check current branch: %w", err)
	}
	checkBranchMsgOut = strings.TrimSpace(checkBranchMsgOut)

	checkAllowBranch := branchRegex.Match([]byte(checkBranchMsgOut))
	if !checkAllowBranch {
		return false, fmt.Errorf("current branch %v is not allowed", checkBranchMsgOut)
	}

	isStaging := strings.TrimSpace(checkBranchMsgOut) == "staging"
	return isStaging, nil
}

func checkCurrentHashNotTagged() error {
	fmt.Fprintf(os.Stdout, "Info: Checking if current git hash already tagged...\n")
	currentGitHash, err := getGitHash()
	if err != nil {
		return fmt.Errorf("could not check git hash: %w", err)
	}

	checkIfNeedNewTagMsgOutput, err := checkGitHashContainTagVersion(currentGitHash)
	if err == nil {
		// No error means the hash is already tagged
		return fmt.Errorf("current git hash %v already contains tag %v", currentGitHash[0:8], checkIfNeedNewTagMsgOutput)
	}

	return nil
}

func calculateNewVersion(tagType string, isStaging bool) (string, error) {
	currentVersionMsgOut, err := getCurrentGitTag()
	if err != nil {
		newVersion := generateNewVersion()
		fmt.Fprintf(os.Stdout, "Info: No tag could be found, New tag version will be %s\n", newVersion)
		return newVersion, nil
	}

	versionParts, err := splitVersion(currentVersionMsgOut)
	if err != nil || len(versionParts) < expectedVersionParts {
		newVersion := generateNewVersion()
		fmt.Fprintf(os.Stdout, "Info: Invalid tag format, New tag version will be %s\n", newVersion)
		return newVersion, nil
	}

	fmt.Fprintf(os.Stdout, "Info: Current Tag Version is %s\n", currentVersionMsgOut)

	currentYearVersion := versionParts[0]
	currentMonthVersion := versionParts[1]
	currentWeekOfTheMonthVersion := versionParts[2]
	currentReleaseVersion := versionParts[3]
	currentFixVersion := versionParts[4]

	if tagType == "release" {
		currentYearVersion, currentMonthVersion, currentWeekOfTheMonthVersion, currentReleaseVersion, currentFixVersion =
			calculateReleaseVersion(currentYearVersion, currentMonthVersion, currentWeekOfTheMonthVersion, currentReleaseVersion)
	} else if tagType == "fix" {
		currentFixVersion = currentFixVersion + 1
	}

	newVersion := formatVersion(currentYearVersion, currentMonthVersion, currentWeekOfTheMonthVersion, currentReleaseVersion, currentFixVersion, isStaging)
	fmt.Fprintf(os.Stdout, "Info: New tag version will be %s\n", newVersion)

	// Check if tag already exists
	_, err = checkGitTagExist(newVersion)
	if err == nil {
		return "", fmt.Errorf("new tag version %s already exists, you can only do fix tag version", newVersion)
	}

	return newVersion, nil
}

func calculateReleaseVersion(currentYear, currentMonth, currentWeek, currentRelease int) (int, int, int, int, int) {
	year := currentYear
	month := currentMonth
	week := currentWeek
	release := currentRelease
	fix := 0

	if getCurrentYear() > year {
		year = getCurrentYear()
	}
	if getCurrentMonth() > month {
		month = getCurrentMonth()
	}
	if week != getWeekOfTheMonth() {
		week = getWeekOfTheMonth()
		release = 1
	} else {
		release = release + 1
	}

	return year, month, week, release, fix
}

func formatVersion(year, month, week, release, fix int, isStaging bool) string {
	if isStaging {
		return fmt.Sprintf("%d.%d.%d.%d.%d", year, month, week, release, fix)
	}
	return fmt.Sprintf("%d.%d.%d.%d.%d-staging", year, month, week, release, fix)
}

func tagAndPush(newVersion string) error {
	if err := setNewVersionTag(newVersion); err != nil {
		return fmt.Errorf("could not set new tag version %v: %w", newVersion, err)
	}

	if err := pushNewVersionTag(newVersion); err != nil {
		return fmt.Errorf("could not push new tag version %v to remote repository: %w", newVersion, err)
	}

	fmt.Fprintf(os.Stdout, "Info: New tag version %v already pushed\n", newVersion)
	return nil
}

func getCurrentYear() int {
	return time.Now().Year()
}

func getCurrentMonth() int {
	return int(time.Now().Month())
}

func getWeekOfTheMonth() int {
	now := time.Now()
	// Calculate week of month: (day of month - 1) / 7 + 1
	return (now.Day()-1)/7 + 1
}

func getCurrentGitTag() (string, error) {
	return execShell("git", "describe", "--abbrev=0", "--tags")
}

func checkCurrentBranch() (string, error) {
	return execShell("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func getGitHash() (string, error) {
	return execShell("git", "rev-parse", "HEAD")
}

func checkGitHashContainTagVersion(hash string) (string, error) {
	return execShell("git", "describe", "--contains", hash)
}

func setNewVersionTag(newVersion string) error {
	_, err := execShell("git", "tag", newVersion)
	return err
}

func pushNewVersionTag(newVersion string) error {
	_, err := execShell("git", "push", "origin", newVersion)
	return err
}

func checkGitTagExist(tag string) (string, error) {
	return execShell("git", "show-ref", "--tags", tag)
}

func execShell(command string, args ...string) (string, error) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func splitVersion(version string) ([]int, error) {
	version = strings.TrimSpace(version)
	resultString := strings.Split(version, ".")
	result := make([]int, len(resultString))
	for idx, i := range resultString {
		convert, err := strconv.Atoi(strings.TrimSpace(i))
		if err != nil {
			return nil, fmt.Errorf("invalid version format: %w", err)
		}
		result[idx] = convert
	}
	return result, nil
}

func generateNewVersion() string {
	return fmt.Sprintf("%d.%d.%d.%d.%d", getCurrentYear(), getCurrentMonth(), getWeekOfTheMonth(), 1, 0)
}

func isFlagPassed() bool {
	passed := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == versionFlagNameShort || f.Name == versionFlagNameLong {
			passed = true
		}
	})
	return passed
}

func askConfirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", message)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
		fmt.Fprintf(os.Stderr, "Sorry dude!, I don't know what to mean :?\n")
	}
}
