/*
	alauncher is an alias generation tool for Windows.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

const ProductName = "alauncher"
const ProductVersion = "0.0.1"

const AlaIniName = ".ala.ini"

func ____util___() {
}

func success() {
	os.Exit(0)
}

func abort(msg string) {
	fmt.Printf("[Error!] %s\n", msg)
	os.Exit(1)
}

func warn(msg string) {
	fmt.Printf("[Warning!] %s\n", msg)
}

func debugprint(useDebugprint bool, msg string) {
	if useDebugprint {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func isInvalidFilename(filename string) bool {
	invalidChars := "\\/:*?\"<>|"
	for _, rune := range invalidChars {
		if strings.Index(filename, string(rune)) != -1 {
			return true
		}
	}
	return false
}

func file2list(filepath string) []string {
	fp, err := os.Open(filepath)
	if err != nil {
		abort(err.Error())
	}
	defer fp.Close()

	lines := []string{}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines
}

func list2file(filepath string, lines []string) {
	fp, err := os.Create(filepath)
	if err != nil {
		abort(err.Error())
	}
	defer fp.Close()

	writer := bufio.NewWriter(fp)
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
}

func string2file(filepath string, contents string) {
	fp, err := os.Create(filepath)
	if err != nil {
		abort(err.Error())
	}
	defer fp.Close()

	writer := bufio.NewWriter(fp)
	writer.WriteString(contents)
	writer.Flush()
}

func isExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func isExistingDirectory(filepath string) bool {
	if !isExist(filepath) {
		return false
	}
	info, _ := os.Stat(filepath)
	return info.IsDir()
}

func ____funcs____() {
}

func createNewAlaIni(filepath string) {
	var initializedContent string = `[_options]
outdir=<A directory passed to PATH>

[_variables]
sys32=%windir%\system32

[np]
bin=notepad.exe

[fo]
bin=%sys32%\control.exe
prm=folders
`

	fmt.Printf("An ini file '%s' not found, so create firstly.\n", filepath)
	string2file(filepath, initializedContent)
}

func editAlaIni(filepath string) {
	err := exec.Command("cmd", "/c", "start", "", filepath).Start()
	if err != nil {
		abort("Fail to open file '" + filepath + "' with the current association.")
	}
}

func removeAlaIni(filepath string) {
	err := os.Remove(filepath)
	if err != nil {
		abort("Fail to remove file " + filepath)
	}
}

func removeAllCommand(outdir string, doDryRun bool) {
	// listup all files
	filesOrDirs, err := ioutil.ReadDir(outdir)
	if err != nil {
		abort("Failed to remove in ioutil.ReadDir()")
	}

	// extract only *.bat files.
	removees := []string{}
	for _, fileOrDir := range filesOrDirs {
		if fileOrDir.IsDir() {
			continue
		}
		filename := fileOrDir.Name()
		if string([]rune(filename)[len(filename)-4:]) != ".bat" {
			continue
		}
		fullpath := filepath.Join(outdir, filename)
		removees = append(removees, fullpath)
	}

	// remove
	for _, removee := range removees {
		if doDryRun {
			fmt.Println("[DryRun] Remove " + removee)
			continue
		}
		err := os.Remove(removee)
		if err != nil {
			warn("Cannot remove " + removee)
		}
	}
}

func createCommandFile(outdir string, basename string, content string, doDryRun bool){
	extension := ".bat"
	filename := basename + extension
	fullpath := filepath.Join(outdir, filename)

	if doDryRun {
		fmt.Println("[DryRun] Create Command " + fullpath + " with:")
		fmt.Println(content)
		return
	}
	string2file(fullpath, content)
}

func ____option____() {
}

type Option struct {
	Outdir string
}

func NewOption(loaded *ini.File) Option {
	option := Option{}

	sec, err := loaded.GetSection("_options")
	if err != nil {
		abort("A section [_options] does not exists.")
	}

	option.Outdir = sec.Key("outdir").Value()
	if !isExistingDirectory(option.Outdir) {
		abort("An entry outdir=xxx do not exists or invalid directory. > " + option.Outdir)
	}

	return option
}

func ____variable____() {
}

type VariableDeployer struct {
	kv map[string]string
}

func NewVariableDeployer(loaded *ini.File) VariableDeployer {
	deployer := VariableDeployer{}
	deployer.kv = map[string]string{} // map needs explicit initialization.

	sec, err := loaded.GetSection("_variables")
	if err != nil {
		abort("A section [_variables] does not exists.")
	}

	keysObj := sec.Keys()
	for _, keyObj := range keysObj {
		k := keyObj.Name()
		v := keyObj.String()
		_, isExist := deployer.kv[k]
		if isExist {
			continue
		}
		deployer.kv[k] = v
	}

	return deployer
}

func (deployer *VariableDeployer) AddAlaVariables() {
	deployer.kv["s"] = " ";
}

func (deployer *VariableDeployer) Deploy(beforestr string) string {
	afterstr := beforestr

	for {
		original := afterstr

		for k, v := range deployer.kv {
			query := fmt.Sprintf("%%%s%%", k)
			afterstr = strings.Replace(afterstr, query, v, -1)
		}

		if original == afterstr {
			// No more variable, so can break immediately.
			//
			// Furthermore, pass here when circular referencing.
			// For example:
			//
			// c1=%c2%
			// c2=%c1%
			//
			// About c1, deploying progress are either below (Case.A) or (Case.B)
			// (Case.A)  c1 -> %c2% -> %c1%   (If the parse order is c1 -> c2)
			// (Case.B)  c1 -> %c2%           (If the parse order is c2 -> c1)
			// If Case.A then it can be pass here
			// If Case.B then it also can be pass here with c2 of Case.B
			break
		}
	}

	return afterstr
}

func ____command____() {
}

type Command struct {
	Name            string
	Rawbin          string
	Bin             string
	Dir             string
	Prm             string
	Aliases         []string
	IsIgnored       bool
	DisableSetlocal bool
	DisableEchoOff  bool
}

func NewCommand() Command {
	command := Command{}
	return command
}

func (command *Command) ImportFromGoIni(section *ini.Section) {
	// [Note] go-ini always use strings.TrimSpace when handling a string.
	// So, the command like 'rawbin=prompt $$ ' actually appeared as 'prompt $$'.
	//                                       ^                                 ^
	//                                                                    Trimmed
	command.Name = section.Name()
	command.Rawbin = section.Key("rawbin").Value()
	command.Bin = section.Key("bin").Value()
	command.Dir = section.Key("dir").Value()
	command.Prm = section.Key("prm").Value()

	aliases := section.Key("alias").Value()
	if aliases != "" {
		command.Aliases = strings.Split(aliases, ",")
	}

	disables := section.Key("disable").Value()
	if disables == "" {
		return
	}
	disablesArray := strings.Split(disables, ",")
	for _, disableEntry := range disablesArray {
		if disableEntry == "setlocal" {
			command.DisableSetlocal = true
		}
		if disableEntry == "echooff" {
			command.DisableEchoOff = true
		}
	}

	existsIgnored := section.Key("ignore_this").Value()
	if existsIgnored != "" {
		command.IsIgnored = true
	}
}

func ____generator____() {
}

type Generator struct {
	command    Command
	deployer VariableDeployer
}

func NewGenerator(command Command, deployer VariableDeployer) Generator {
	generator := Generator{}
	generator.command = command
	generator.deployer = deployer
	return generator
}

func (generator *Generator) abortIfInvalid() {
	command := generator.command
	name := command.Name

	// Check invalid chars as a filename
	// because it is used as an command filename
	if isInvalidFilename(name) {
		abort("Invalid command because invalid as a filename: " + name)
	}
}

func (generator *Generator) Generate() string {
	command := generator.command
	deployer := generator.deployer

	if command.IsIgnored {
		return ""
	}

	generator.abortIfInvalid()

	statementSetlocal := "setlocal"
	statementEchoOff := "@echo off"
	if command.DisableSetlocal {
		statementSetlocal = ""
	}
	if command.DisableEchoOff {
		 statementEchoOff = ""
	}

	// [Template raw bin]

	var templateRawbin string = `%s

%s

%s`

	if command.Rawbin != "" {
		batchContents := fmt.Sprintf(templateRawbin, statementEchoOff, statementSetlocal, command.Rawbin)
		batchContents = deployer.Deploy(batchContents)
		return batchContents
	}

	// [Template bin]

	// Acceptable variation of 'bin', 'dir' and 'prm'
	//
	// bin dir prm  valid?
	// o   o   o    y
	// o   o   x    y       simply, No parameter.
	// o   x   o    y       dir is completed with 'cur'
	// o   x   x    y       dir is completed from 'cur', and no parameter.
	// x   o   o    n
	// x   o   x    n
	// x   x   o    n
	// x   x   x    n

	var templateBin string = `%s

%s

pushd %s

start "" "%s" %s

popd`

	if command.Bin == "" {
		abort("Invalid command, No 'bin' value about [" + command.Name + "]")
	}

	binaryPath := command.Bin

	parameter := command.Prm

	// variation of 'dir' value
	// - "cur": Current Directory
	// - ""   : Same as "cur"
	// - "bin": Directory of bin value
	// - Else : Use it

	directory := ""
	switch command.Dir {
	case "cur", "":
		directory = "%cd%"
	case "bin":
		directory = filepath.Dir(binaryPath)
	default:
		directory = command.Dir
	}

	batchContents := fmt.Sprintf(templateBin, statementEchoOff, statementSetlocal, directory, binaryPath, parameter)
	batchContents = deployer.Deploy(batchContents)
	return batchContents
}

func (generator *Generator) GenerateAlias() string {
	command := generator.command

	if len(command.Aliases) == 0 {
		return ""
	}

	// Do not accept disable=echooff because it is enough the parent settings.
	var templateAlias string = `@echo off

call %%~dp0%s.bat %s
`

	batchContents := fmt.Sprintf(templateAlias, command.Name, command.Prm)
	return batchContents
}

func ____argument____() {
}

type Args struct {
	debugPrint *bool
	doInit     *bool

	doDryRun  *bool
	doEditIni *bool
	doRemove  *bool
}

func argparse(isForTest bool) Args {
	args := Args{}

	args.doDryRun = flag.Bool("dryrun", false, "DryRun(Not Remove/Write operation but display only), about -run and -edit.")
	args.doEditIni = flag.Bool("edit", false, "Edit the ini file with the current association of .ini.")
	args.doRemove = flag.Bool("remove", false, "Remove current aliases.")

	args.doInit = flag.Bool("debug-init", false, "[DEBUG] Remove the ini file. You must re-execute to initialize the ini file.")
	args.debugPrint = flag.Bool("debug-print", false, "[DEBUG] print all options with name and value.")

	isShowingVersion := flag.Bool("version", false, "Show this alauncher version.")

	flag.Parse()

	if isForTest == true {
		return args
	}

	if *isShowingVersion {
		fmt.Printf("%s %s\n", ProductName, ProductVersion)
		success()
	}

	// Preprocess
	// ----------

	printOption := func(flg *flag.Flag) {
		fmt.Printf("%s=%s\n", flg.Name, flg.Value)
	}
	if *args.debugPrint {
		fmt.Println("==== Options ====")
		flag.VisitAll(printOption)
	}

	return args
}

func main() {
	isForTest := false
	args := argparse(isForTest)
	doInit := *args.doInit
	doRemove := *args.doRemove
	doEditIni := *args.doEditIni
	doDryRun := *args.doDryRun
	useDebugprint := *args.debugPrint

	dirHome := os.Getenv("HOME")
	if dirHome == "" {
		dirHome = os.Getenv("USERPROFILE")
		if dirHome == "" {
			abort("Fail to search home directory.")
		}
	}
	debugprint(useDebugprint, "HomeDirectory: "+dirHome)

	filepathAlaIni := filepath.Join(dirHome, AlaIniName)

	if !isExist(filepathAlaIni) {
		createNewAlaIni(filepathAlaIni)
		editAlaIni(filepathAlaIni)
		success()
	}

	if doInit {
		removeAlaIni(filepathAlaIni)
		success()
	}

	if doEditIni {
		editAlaIni(filepathAlaIni)
		success()
	}

	cfg, err := ini.Load(filepathAlaIni)
	if err != nil {
		abort("Fail to read ini file '" + filepathAlaIni + "'.")
	}

	// Parse.

	option := NewOption(cfg)
	debugprint(useDebugprint, "OutDir: "+option.Outdir)

	deployer := NewVariableDeployer(cfg)
	deployer.AddAlaVariables()

	sections := cfg.Sections()
	commands := []Command{}
	for _, section := range sections {
		sectionName := section.Name()
		if sectionName == "DEFAULT" {
			continue
		}
		if string([]rune(sectionName)[:1]) == "_" {
			continue
		}
		command := NewCommand()
		command.ImportFromGoIni(section)
		commands = append(commands, command)
	}
	for _, command := range commands {
		debugprint(useDebugprint, "Read command '"+command.Name+"'")
	}

	// Create.
	// Firstly remove all and secondly create.

	removeAllCommand(option.Outdir, doDryRun)
	if doRemove {
		success()
	}

	for _, command := range commands {
		if command.IsIgnored {
			continue
		}
		gen := NewGenerator(command, deployer)
		createCommandFile(option.Outdir, command.Name, gen.Generate(), doDryRun)
		// Generate aliases of the Command if given.
		aliasContents := gen.GenerateAlias()
		for _, aliasName := range command.Aliases {
			createCommandFile(option.Outdir, aliasName, aliasContents, doDryRun)
		}
	}

}
