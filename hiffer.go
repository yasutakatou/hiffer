package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"gopkg.in/ini.v1"

	"crypto/md5"
	"encoding/hex"

	"github.com/abiosoft/readline"
	"github.com/fatih/color"
	"github.com/kylelemons/godebug/diff"
	"github.com/nsf/termbox-go"
	"github.com/saintfish/chardet"
	"github.com/yasutakatou/ishell"
)

var DEBUG bool
var RECORD bool
var TEST string
var TEMPDir string
var prompt string
var useShell string
var maxHistorys int
var prevDir string
var cancelCTRLc chan os.Signal
var OSDIR string
var rs1Letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var shell *ishell.Shell
var termSize int

var completer readline.AutoCompleter
var inputs string
var filelists []string
var tabFlag bool

type historyData struct {
	Command string `json:"Command"`
	Output  string `json:"Output"`
	Sum     string `json:"Sum"`
}

var History = []historyData{}

type convertData struct {
	Word     string `json:"Word"`
	PreAfter string `json:"PreAfter"`
	Regex    string `json:"Regex"`
	Replace  string `json:"Replace"`
}

var Convert = []convertData{}

//FYI: https://journal.lampetty.net/entry/capturing-stdout-in-golang
type Capturer struct {
	saved         *os.File
	bufferChannel chan string
	out           *os.File
	in            *os.File
}

type Lens struct {
	fileName        string
	buffer          []string
	top             int
	bottom          int
	margin          int
	showLineNumbers string
}

const (
	defaultColor = termbox.ColorDefault
)

func init() {
	shell = ishell.New()

	rand.Seed(time.Now().UnixNano())

	RECORD = true
	DEBUG = false
	TEMPDir = "tmp"
	useShell = "/bin/bash"
	TEST = "all"
	//prompt = "`PWD`"
	prompt = ">>"
	maxHistorys = 99
	prevDir, _ = filepath.Abs(".")
	filelists = listFile(".")
	shell.AddCompleters(filelists)

	cancelCTRLc = make(chan os.Signal)
	signal.Notify(cancelCTRLc, os.Interrupt)

	if runtime.GOOS == "linux" {
		OSDIR = "/"
	} else {
		OSDIR = "\\"
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	_, termSize = termbox.Size()
	termSize = termSize - 1
}

func main() {
	_DEBUG := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_configFile := flag.String("config", ".hiffer", "[-config=config file (default: .hiffer)]")

	flag.Parse()
	configFile := string(*_configFile)
	DEBUG = bool(*_DEBUG)

	if Exists(configFile) == true {
		loadDotFile(configFile)
	}

	if DEBUG == true {
		t := time.Now()
		const layout = "2006-01-02-15-04-05"
		saveDotFile(configFile + "_" + t.Format(layout))
	}

	envCmd := &ishell.Cmd{Name: "@env",
		Help: "set VALUES, or show VALUES",
		Func: commandHandler}

	envCmd.AddCmd(&ishell.Cmd{Name: "DEBUG=true"})
	envCmd.AddCmd(&ishell.Cmd{Name: "RECORD=true"})
	envCmd.AddCmd(&ishell.Cmd{Name: "TEST=\"\""})
	envCmd.AddCmd(&ishell.Cmd{Name: "SHELL=\"\""})
	envCmd.AddCmd(&ishell.Cmd{Name: "MAXHISTORYS="})
	envCmd.AddCmd(&ishell.Cmd{Name: "PROMPT=\"\""})
	envCmd.AddCmd(&ishell.Cmd{Name: "TEMP=\"\""})
	shell.AddCmd(envCmd)

	shell.AddCmd(&ishell.Cmd{Name: "@show",
		Help: "show history without output",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "@del",
		Help: "del history",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "@ins",
		Help: "insert history",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "@test",
		Help: "test!",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "exit",
		Help: "exit and export exists history",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "@kill",
		Help: "clear all historys",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "@export",
		Help: "export config",
		Func: commandHandler})

	shell.AddCmd(&ishell.Cmd{Name: "default",
		Help: "default is add history",
		Func: commandHandler})

	shell.Run()
}

func saveDotFile(filename string) bool {
	if len(filename) == 0 {
		return false
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	writeFile(file, "[TEST]")
	writeFile(file, TEST+"\n")

	writeFile(file, "[SHELL]")
	writeFile(file, useShell+"\n")

	writeFile(file, "[MAXHISTORYS]")
	writeFile(file, strconv.Itoa(maxHistorys)+"\n")

	writeFile(file, "[PROMPT]")
	writeFile(file, prompt+"\n")

	writeFile(file, "[TEMP]")
	writeFile(file, TEMPDir+"\n")

	writeFile(file, "[CONVERT]")
	for i := 0; i < len(Convert); i++ {
		writeFile(file, Convert[i].Word+","+Convert[i].PreAfter+","+Convert[i].Regex+","+Convert[i].Replace)
	}
	writeFile(file, "\n")

	writeFile(file, "[HISTORY]")
	for i := 0; i < len(History); i++ {
		writeFile(file, History[i].Command+","+History[i].Output+","+History[i].Sum)
	}

	return true
}

func writeFile(file *os.File, strs string) bool {
	_, err := file.WriteString(strs + "\n")
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func setSingleConfig(config *string, configType, datas string) {
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			*config = v
		}
		if DEBUG == true {
			fmt.Println(" -- " + configType + " --")
			fmt.Println(v)
		}
	}
}

func loadDotFile(filename string) {
	loadOptions := ini.LoadOptions{}
	loadOptions.UnparseableSections = []string{"TEST", "SHELL", "MAXHISTORYS", "PROMPT", "TEMP", "CONVERT", "HISTORY"}

	cfg, err := ini.LoadSources(loadOptions, filename)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	setSingleConfig(&TEST, "TEST", cfg.Section("TEST").Body())

	setSingleConfig(&useShell, "SHELL", cfg.Section("SHELL").Body())
	maxHis := ""
	setSingleConfig(&maxHis, "MAXHISTORYS", cfg.Section("MAXHISTORYS").Body())
	cnt, err := strconv.Atoi(maxHis)
	if err == nil {
		maxHistorys = cnt
	}

	setSingleConfig(&prompt, "PROMPT", cfg.Section("PROMPT").Body())
	setSingleConfig(&TEMPDir, "TEMP", cfg.Section("TEMP").Body())

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(cfg.Section("CONVERT").Body(), -1) {
		if len(v) > 0 {
			out := strings.Split(v, ",")
			Convert = append(Convert, convertData{Word: out[0], PreAfter: out[1], Regex: out[2], Replace: out[3]})
		}
	}
	if DEBUG == true {
		fmt.Println(" -- CONVERT --")
		DisplayConvert()
	}

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(cfg.Section("HISTORY").Body(), -1) {
		if len(v) > 0 {
			out := strings.Split(v, ",")
			History = append(History, historyData{Command: out[0], Output: out[1], Sum: out[2]})
		}
	}
	if DEBUG == true {
		fmt.Println(" -- HISTORY --")
		DisplayHistory(false)
	}
}

func commandHandler(c *ishell.Context) {
	tabFlag = false
	params := ""

	if len(c.Args) > 0 {
		params = c.Args[0]
		for i := 1; i < len(c.Args); i++ {
			params += " "
			params += c.Args[i]
		}
	}

	if DEBUG == true {
		fmt.Println("Command:" + c.Cmd.Name + " Params: " + params)
	}

	RunCmd(c, params)
}

func envAndshow(command, params string) {
	switch command {
	case "@env":
		if len(params) == 0 {
			showConfigs()
		} else {
			fmt.Println(OptionSetting(params))
		}
	case "@show":
		if len(params) == 0 {
			DisplayHistory(false)
		} else {
			if params == "all" {
				DisplayHistory(true)
			} else {
				displayHistoryDetail(params)
			}
		}
	}
}

func insertCmd(params string) {
	param := strings.Split(params, " ")
	out, error := Execmd(strings.Replace(params, param[0]+" ", "", 1), RECORD, param[0])
	if error != "Error" {
		fmt.Println(out)
	} else {
		fmt.Println(error, out)
	}
}

func RunCmd(c *ishell.Context, params string) bool {
	inputs = ""

	switch c.Cmd.Name {
	case "@env", "@show":
		envAndshow(c.Cmd.Name, params)
	case "@del":
		DeleteHistory(params)
	case "@ins":
		insertCmd(params)
	case "@export":
		saveDotFile(params)
	case "@test":
		doTest(params)
	case "exit":
		close(cancelCTRLc)
		os.Chdir(prevDir)
		saveDotFile(".hiffer")
		os.Exit(0)
	case "@kill":
		err := os.RemoveAll(prevDir + OSDIR + TEMPDir)
		if err != nil {
			fmt.Println(err)
		}
		History = nil
	default:
		defaultCmd(c, params)
	}
	return true
}

func defaultCmd(c *ishell.Context, params string) {
	if strings.Index(params, "cd ") == 0 {
		err := os.Chdir(strings.Replace(params, "cd ", "", 1))
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err))
		} else {
			filelists = listFile(".")
			shell.AddCompleters(filelists)
		}
		changePrompt(c)
	} else {
		out, error := Execmd(params, RECORD, "0")
		if error == "Error" {
			fmt.Println(error, out)
		} else {
			fmt.Printf(out)
		}
		changePrompt(c)
	}
}

func changePrompt(c *ishell.Context) {
	if strings.Index(prompt, "`") != -1 {
		if DEBUG == true {
			fmt.Printf("prompt: ")
		}

		out, error := Execmd(strings.Replace(prompt, "`", "", -1), false, "0")
		if error == "Error" {
			fmt.Println(error, out)
		}
		out = strings.Replace(out, "\n", "", -1)
		out = strings.Replace(out, "\r", "", -1)
		c.SetPrompt(out + "> ")
	} else {
		c.SetPrompt(prompt + "> ")
	}
}

func convertCommand(command, status string) string {
	if len(Convert) == 0 {
		return command
	}
	for i := 0; i < len(Convert); i++ {
		if strings.Index(command, Convert[i].Word) != -1 && Convert[i].PreAfter == status {
			rep := regexp.MustCompile(Convert[i].Regex)
			command = rep.ReplaceAllString(command, Convert[i].Replace)
			return command
		}
	}
	return command
}

func OptionSetting(options string) string {
	params := strings.Split(options, "=")

	if len(params) < 2 || len(options) == 0 {
		return "error"
	}

	switch params[0] {
	case "DEBUG":
		return setTrueFalse(&DEBUG, params[1])
	case "RECORD":
		return setTrueFalse(&RECORD, params[1])
	case "TEST":
		TEST = params[1]
	case "SHELL":
		useShell = params[1]
	case "MAXHISTORYS":
		cnt, err := strconv.Atoi(params[1])
		if err == nil {
			maxHistorys = cnt
			if len(History) > maxHistorys {
				DeleteHistory("1-" + strconv.Itoa(len(History)-maxHistorys))
			}
		}
	case "PROMPT":
		prompt = params[1]
	case "TEMP":
		TEMPDir = params[1]
	default:
		return "error"
	}
	return ""
}

func setTrueFalse(truefalse *bool, strs string) string {
	if strs == "true" {
		*truefalse = true
		return ""
	}

	if strs == "false" {
		*truefalse = false
		return ""
	}
	return "value set failure (usecase [value=true/false])"
}

func showConfigs() {
	fmt.Println(" -- DEBUG --")
	if DEBUG == true {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	fmt.Println(" -- RECORD --")
	if RECORD == true {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	fmt.Println(" -- TEST --")
	fmt.Println(TEST)

	fmt.Println(" -- SHELL --")
	fmt.Println(useShell)

	fmt.Println(" -- MAXHISTORYS --")
	fmt.Println(maxHistorys)

	fmt.Println(" -- PROMPT --")
	fmt.Println(prompt)

	fmt.Println(" -- TEMP --")
	fmt.Println(TEMPDir)

	fmt.Println(" -- Convert --")
	DisplayConvert()
}

func DisplayHistory(allFlag bool) bool {

	if len(History) == 0 {
		return false
	}

	if allFlag == true && len(History) > termSize {
		var listHistory []string

		for i := 0; i < len(History); i++ {
			listHistory = append(listHistory, fmt.Sprintf("[%3d] Command: %30s Params: %s Sum %s\n", i+1, History[i].Command, History[i].Output, History[i].Sum))
		}
		displayMaxLineSplit(listHistory)

		return true
	}

	cnt := 0
	if len(History) > termSize {
		cnt = len(History) - termSize
	}
	for i := cnt; i < len(History); i++ {
		fmt.Printf("[%3d] Command: %30s Params: %s Sum %s\n", i+1, History[i].Command, History[i].Output, History[i].Sum)
	}
	return true
}

func DisplayConvert() bool {
	if len(Convert) == 0 {
		return false
	}
	for i := 0; i < len(Convert); i++ {
		fmt.Printf("[%3d] Word: %s PreAfeter: %s Regex: %s Replace: %s\n", i+1, Convert[i].Word, Convert[i].PreAfter, Convert[i].Regex, Convert[i].Replace)
	}
	return true
}

func Unset(s []historyData, min, max int) []historyData {
	return append(s[:min], s[max:]...)
}

func DeleteHistory(ranges string) bool {
	if strings.Index(ranges, "-") != -1 {
		params := strings.Split(ranges, "-")
		if len(params) == 2 {
			min, err := strconv.Atoi(params[0])
			if err != nil {
				return false
			}

			max, err := strconv.Atoi(params[1])
			if err != nil {
				return false
			}

			if min > 0 && len(History) >= max && min < max {
				for i := min - 1; i < max; i++ {
					DeleteOutput(i)
				}

				History = Unset(History, min-1, max)
				return true
			}
		}
	}

	cnt, err := strconv.Atoi(ranges)
	if err != nil {
		return false
	}

	if err == nil && cnt > 0 && len(History) >= cnt {
		DeleteOutput(cnt - 1)
		History = Unset(History, cnt-1, cnt)
		return true

	}
	return false
}

func DeleteOutput(index int) {
	if Exists(prevDir+OSDIR+TEMPDir+OSDIR+History[index].Output) == true {
		if err := os.Remove(prevDir + OSDIR + TEMPDir + OSDIR + History[index].Output); err != nil {
			fmt.Println(err)
		}
		if DEBUG == true {
			fmt.Printf("Delete: %s\n", History[index].Output)
		}
	}
}

func Insert(s []historyData, cnt int, command, params, sum string) []historyData {
	s = append(s[:cnt+1], s[cnt:]...)
	s[cnt] = historyData{Command: command, Output: params, Sum: sum}
	return s
}

func Execmd(command string, rFlag bool, insert string) (string, string) {
	var cmd *exec.Cmd
	var out string
	var err error

	if len(command) == 0 {
		return "Error", "No Command"
	}

	if command[:1] == "!" {
		rFlag = false
		command = strings.Replace(command, "!", "", 1)
	}

	insertCnt := 0

	if insert != "0" {
		insertCnt, err = strconv.Atoi(insert)
		if err != nil || insertCnt < 1 && len(History) < insertCnt {
			return "Error", "Don't Insert Number"
		}
	}

	command = convertCommand(command, "true")

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command(useShell, "-c", command)
	case "windows":
		cmd = exec.Command("cmd", "/C", command)
	}

	if rFlag == true {
		c := &Capturer{}
		c.StartCapturingStdout()

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()

		out = c.StopCapturingStdout()
	} else {
		outs, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
		out = string(outs)
	}

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest([]byte(out))
	if err == nil {
		if result.Charset == "Shift_JIS" {
			out, _ = sjis_to_utf8(out)
		}
	}

	if rFlag == true {
		rnd := RandStr(8)
		if outputToFile(rnd, out) == false {
			fmt.Println("Error: don't output file: ", prevDir+OSDIR+TEMPDir+OSDIR+rnd)
		} else {
			command = convertCommand(command, "false")
			if insertCnt == 0 {
				if len(History) >= maxHistorys {
					DeleteHistory("1")
				}
				History = append(History, historyData{Command: command, Output: rnd, Sum: GetMD5Hash(out)})
			} else {
				if len(History) >= maxHistorys {
					DeleteHistory("1")
				}
				History = Insert(History, insertCnt-1, command, rnd, GetMD5Hash(out))
			}
		}
	}
	return out, GetMD5Hash(out)
}

// 標準出力をキャプチャする
func (c *Capturer) StartCapturingStdout() {
	c.saved = os.Stdout
	var err error
	c.in, c.out, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = c.out
	c.bufferChannel = make(chan string)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, c.in)
		c.bufferChannel <- b.String()
	}()
}

// キャプチャを停止する
func (c *Capturer) StopCapturingStdout() string {
	c.out.Close()
	os.Stdout = c.saved
	return <-c.bufferChannel
}

//FYI: https://qiita.com/uchiko/items/1810ddacd23fd4d3c934
// ShiftJIS から UTF-8
func sjis_to_utf8(str string) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func outputToFile(filename, output string) bool {
	if Exists(prevDir+OSDIR+TEMPDir) == false {
		if err := os.MkdirAll(prevDir+OSDIR+TEMPDir, 0777); err != nil {
			fmt.Println(err)
		}
	}

	file, err := os.Create(prevDir + OSDIR + TEMPDir + OSDIR + filename)
	if err != nil {
		return false
	}
	defer file.Close()

	file.Write(([]byte)(output))
	return true
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func doTest(params string) bool {
	if len(params) == 0 {
		params = TEST
	}

	if params == "all" {
		params = "1-" + strconv.Itoa(len(History))
	}

	params = params + ","

	testCase := strings.Split(params, ",")
	for i := 0; i < len(testCase); i++ {
		if len(testCase[i]) > 0 && strings.Index(testCase[i], "-") != -1 {
			params := strings.Split(testCase[i], "-")
			if len(params) == 2 {
				min, err := strconv.Atoi(params[0])
				if err == nil {
					max, err := strconv.Atoi(params[1])

					if err == nil && min > 0 && len(History) >= max && min < max {
						for i := min - 1; i < max; i++ {
							checkOutput(i + 1)
						}
					}
				}
			}
		} else {
			if len(testCase[i]) > 0 {
				cnt, err := strconv.Atoi(testCase[i])
				if err == nil {
					checkOutput(cnt)
				}
			}
		}
	}
	return true
}

func checkOutput(cnt int) {
	if cnt > 0 && len(History) >= cnt {
		out, hash := Execmd(History[cnt-1].Command, false, "0")
		if hash != History[cnt-1].Sum {
			color.Red(" -- DIFF! %s -- \n", History[cnt-1].Command)
			diffToColor(diff.Diff(readOutput(prevDir+OSDIR+TEMPDir+OSDIR+History[cnt-1].Output), out))
		} else {
			color.Blue(" -- OK! %s -- \n", History[cnt-1].Command)
		}
	}
}

func diffToColor(strs string) {
	cnt := 1
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(strs, -1) {
		if len(v) > 0 {
			switch v[:1] {
			case "+":
				color.Magenta("%d: %s\n", cnt, v)
			case "-":
				color.Cyan("%d: %s\n", cnt, v)
			default:
				if DEBUG == true {
					fmt.Printf("%d: %s\n", cnt, v)
				}
			}
		}
		cnt = cnt + 1
	}
}

func readOutput(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func displayHistoryDetail(params string) bool {
	cnt, err := strconv.Atoi(params)
	if err == nil && cnt > 0 && cnt <= len(History) {
		strs := readOutput(prevDir + OSDIR + TEMPDir + OSDIR + History[cnt-1].Output)
		var output []string

		for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(strs, -1) {
			output = append(output, v)
		}
		displayMaxLineSplit(output)
		return true
	}
	return false
}

func switchRawcode(Rawcode uint16) {
	if Rawcode == 13 {
		inputs = ""
	}
	if Rawcode == 8 {
		if len(inputs) > 0 {
			inputs = inputs[0:(len(inputs) - 1)]
		}
	}
	if Rawcode != 9 && tabFlag == false {
		inputs = inputs + strings.ToLower(string(Rawcode))
		shell.AddCompleters(filelists)
	} else {
		tabFlag = true
	}
}

func listFile(path string) []string {
	var lists []string

	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		lists = append(lists, f.Name())
	}
	return lists
}

// FYI: https://github.com/notduncansmith/go-less
func (l *Lens) setBuffer(lines []string) {
	newBuf := make([]string, 0, len(lines))

	for i, v := range lines {
		newBuf = append(newBuf, l.getMargin(i)+v)
	}

	l.buffer = newBuf
}

func (l *Lens) len() int {
	return len(l.buffer)
}

func (l *Lens) size() int {
	return l.top - l.bottom
}

func (l *Lens) printToBuffer() {
	termbox.Clear(defaultColor, defaultColor)

	screen := make([]string, 0, l.len()+1)

	screen = append(screen, l.buffer[l.top:l.bottom]...)
	screen = append(screen, " -- -- PageDown:→,h PageUp:←,l Up:↑,k Down:↓,j,Enter Exit:Esc,q -- -- ")

	for lineIndex, line := range screen {
		x := 0

		for _, chr := range line {
			termbox.SetCell(x, lineIndex, chr, defaultColor, defaultColor)
			x++
		}
	}
}

func (l *Lens) lineDown() {
	l.bottom = l.bottom + termSize
	l.top = l.top + termSize

	if l.bottom >= l.len() {
		l.bottom = len(l.buffer)
		l.top = len(l.buffer) - termSize
	}

	l.printToBuffer()
	termbox.Flush()
}

func (l *Lens) lineUp() {
	l.bottom = l.bottom - termSize
	l.top = l.top - termSize

	if l.top <= 0 {
		l.bottom = termSize
		l.top = 0
	}

	l.printToBuffer()
	termbox.Flush()
}

func (l *Lens) oneDown() {
	if l.bottom < l.len() {
		l.bottom = l.bottom + 1
		l.top = l.top + 1
	}

	l.printToBuffer()
	termbox.Flush()
}

func (l *Lens) oneUp() {
	if l.top > 0 {
		l.bottom = l.bottom - 1
		l.top = l.top - 1
	}

	l.printToBuffer()
	termbox.Flush()
}

func (l *Lens) getMargin(lineNumber int) string {
	if l.showLineNumbers == "N" {
		return marginWithLineNumbers(lineNumber)
	}

	return marginWithoutLineNumbers()
}

func marginWithoutLineNumbers() string {
	m := "  "

	for len(m) < Margin {
		m = " " + m
	}

	return m
}

func marginWithLineNumbers(lineNumber int) string {
	m := strconv.Itoa(lineNumber)

	for len(m) < Margin {
		m = " " + m
	}

	return m + "  "
}

func (l *Lens) listen() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	l.bottom = termSize

	for {
		l.printToBuffer()
		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 13, 65516: //Enter, and Down
				l.oneDown()
			case 65517:
				l.oneUp() // Upper
			case 65515: // Left
				l.lineUp()
			case 32, 65514: // Right, and Space
				l.lineDown()
			case 27: //ESC
				return
			default:
			}

			switch ev.Ch {
			case 'j':
				l.oneDown()
			case 'k':
				l.oneUp()
			case 'l':
				l.lineUp()
			case 'h':
				l.lineDown()
			case 'q':
				return
			default:
			}
		}
	}
}

const (
	Margin = 2
)

func makeLens(name string, lines []string, showLineNumbers string) Lens {
	l := Lens{}

	l.showLineNumbers = showLineNumbers
	l.setBuffer(lines)
	l.margin = Margin
	l.top = 0
	l.fileName = name

	return l
}

// FYI: https://github.com/notduncansmith/go-less

func displayMaxLineSplit(strs []string) {
	if len(strs) < termSize {
		for i := 0; i < len(strs); i++ {
			fmt.Println(strs[i])
		}
		return
	}
	showLineNums := "F"

	l := makeLens("", strs, showLineNums)
	l.listen()
}
