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
)

const (
	defaultTagType       string = "fix"
	versionFlagNameLong  string = "version"
	versionFlagNameShort string = "v"
	allowBranch          string = "main"
	usageLong            string = "-version [release|fix]"
	usageShort           string = "shorthand of 'version'"
	usageHelper          string = `Usage: %s
Options:
`
)

var tagType string
var newVersion string

func main() {

	flag.StringVar(&tagType, versionFlagNameLong, defaultTagType, usageLong)
	flag.StringVar(&tagType, versionFlagNameShort, defaultTagType, usageShort)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), usageHelper, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	checkFlagIsPassed := isFlagPassed()
	if !checkFlagIsPassed && tagType != "release" {
		fmt.Printf("You are not provide the parameter, the default will tagging as 'fix'\n")
		confirm := askConfirm("Please confirm to continue")
		if !confirm {
			_, _ = fmt.Fprintf(os.Stdout, "Alright then, see ya!\n")
			os.Exit(0)
		}
	}

	checkBranchStatus, checkBranchMsgOut, checkBranchMsgError := checkCurrentBranch()
	if checkBranchStatus != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", checkBranchMsgError)
		os.Exit(1)
	}
	checkBranchMsgOut = strings.TrimSpace(checkBranchMsgOut)

	checkAllowBranch, _ := regexp.Match(allowBranch, []byte(checkBranchMsgOut))
	if !checkAllowBranch {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Then current branch %v are not allow in list\n", checkBranchMsgOut)
		os.Exit(1)
	}
	isStaging := false
	if strings.TrimSpace(checkBranchMsgOut) == "staging" {
		isStaging = true
	}

	_, _ = fmt.Fprintf(os.Stdout, "Info: Checking if current git hash already tagged...\n")
	currentGitHashStatus, currentGitHashMsgOutput, currentGitHashError := getGitHash()
	if currentGitHashStatus != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Could not check git tag\n")
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", currentGitHashError)
		os.Exit(1)
	}

	checkIfNeedNewTagStatus, checkIfNeedNewTagMsgOutput, _ := checkGitHashContainTagVersion(currentGitHashMsgOutput)
	if checkIfNeedNewTagStatus == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Current git hash %v already contain tag with %v\n", currentGitHashMsgOutput[0:8], checkIfNeedNewTagMsgOutput)
		os.Exit(0)
	}

	currentVersionStatus, currentVersionMsgOut, _ := getCurrentGitTag()
	countDigit := len(splitVersion(currentVersionMsgOut))
	if currentVersionStatus != nil || countDigit < 5 {
		newVersion = generateNewVersion()
		_, _ = fmt.Fprintf(os.Stdout, "Info: No tag could be found, New tag version will be %s\n", newVersion)
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "Info: Current Tag Version is %s\n", currentVersionMsgOut)
		currentYearVersion := splitVersion(currentVersionMsgOut)[0]
		currentMonthVersion := splitVersion(currentVersionMsgOut)[1]
		currentWeekOfTheMonthVersion := splitVersion(currentVersionMsgOut)[2]
		currentReleaseVersion := splitVersion(currentVersionMsgOut)[3]
		currentFixVersion := splitVersion(currentVersionMsgOut)[4]
		if tagType == "release" {
			// Release on same month but new week
			// Previous 2022.3.1.1.10
			// Change to 2022.3.2.1.0
			// Release on the new month
			// Previous 2022.4.1.1.10
			// Change to 2022.5.1.1.0
			if getCurrentYear() > currentYearVersion {
				currentYearVersion = getCurrentYear()
			}
			if getCurrentMonth() > currentMonthVersion {
				currentMonthVersion = getCurrentMonth()
			}
			if currentWeekOfTheMonthVersion != getWeekOfTheMonth() {
				currentWeekOfTheMonthVersion = getWeekOfTheMonth()
				currentReleaseVersion = 1
				currentFixVersion = 0
			} else {
				// Release same week
				// Previous 2022.3.1.1.10
				// Change to 2022.3.1.2.0
				currentReleaseVersion = currentReleaseVersion + 1
				currentFixVersion = 0
			}
		} else if tagType == "fix" {
			// Keep all previous version except increasing fix number
			// Previous 2022.3.1.1.10
			// Change to 2022.3.1.1.11
			// Previous 2022.12.2.2.10
			// Change to 2022.12.2.2.11 // But first week of new year 2023
			currentFixVersion = currentFixVersion + 1
		}
		newVersion = fmt.Sprintf("%d.%d.%d.%d.%d-%s", currentYearVersion, currentMonthVersion, currentWeekOfTheMonthVersion, currentReleaseVersion, currentFixVersion, "staging")
		if isStaging {
			newVersion = fmt.Sprintf("%d.%d.%d.%d.%d", currentYearVersion, currentMonthVersion, currentWeekOfTheMonthVersion, currentReleaseVersion, currentFixVersion)
		}

		_, _ = fmt.Fprintf(os.Stdout, "Info: New tag version will be %s\n", newVersion)

		isTagExistStatus, isTagExistMsgOutput, _ := checkGitTagExist(newVersion)
		if isTagExistStatus == nil {
			_, _ = fmt.Fprintf(os.Stdout, "%v\n", isTagExistMsgOutput)
			_, _ = fmt.Fprintf(os.Stderr, "Error: New tag version %s already tagged, You can do only fix tag version\n", newVersion)
			os.Exit(1)
		}
	}

	setNewTagStatus, _, setNewTagError := setNewVersionTag(newVersion)
	if setNewTagStatus != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Could not set new tag version %v\n", newVersion)
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", setNewTagError)
		os.Exit(1)
	}
	pushNewTagStatus, _, pushNewTagError := pushNewVersionTag(newVersion)
	if pushNewTagStatus != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Could not push new tag version %v to remote repository\n", newVersion)
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", pushNewTagError)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Info: New tag version %v already pushed\n", newVersion)
	_, _ = fmt.Fprintf(os.Stdout, "Info: All set, Goodbye\n")
}

func getCurrentYear() int {
	return carbon.Now().Year()
}

func getCurrentMonth() int {
	return carbon.Now().Month()
}

func getWeekOfTheMonth() int {
	return carbon.Now().WeekOfMonth()
}

func getCurrentGitTag() (error, string, string) {
	return execShell("git describe --abbrev=0 --tags")
}

func checkCurrentBranch() (error, string, string) {
	return execShell("git rev-parse --abbrev-ref HEAD")
}

func getGitHash() (error, string, string) {
	return execShell("git rev-parse HEAD")
}

func checkGitHashContainTagVersion(hash string) (error, string, string) {
	cmd := "git describe --contains " + hash
	return execShell(cmd)
}

func setNewVersionTag(newVersion string) (error, string, string) {
	cmd := "git tag " + newVersion
	return execShell(cmd)
}

func pushNewVersionTag(newVersion string) (error, string, string) {
	cmd := "git push origin " + newVersion
	return execShell(cmd)
}

func checkGitTagExist(tag string) (error, string, string) {
	cmd := "git show-ref --tags " + tag
	return execShell(cmd)
}

func execShell(command string) (error, string, string) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func splitVersion(version string) []int {
	resultString := strings.Split(string(version), ".")
	result := make([]int, len(resultString))
	for idx, i := range resultString {
		convert, err := strconv.Atoi(strings.TrimSpace(i))
		if err != nil {
			panic(err)
		}
		result[idx] = convert
	}
	return result
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
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Sorry dude!, I don't know what to mean :?\n")
			os.Exit(1)
		}
	}
}
